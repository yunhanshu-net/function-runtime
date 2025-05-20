package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/pkg/constants"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/x/cmdx"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/pkg/x/jsonx"
	"github.com/yunhanshu-net/pkg/x/stringsx"
	"github.com/yunhanshu-net/runcher/runner/coder"
	"github.com/yunhanshu-net/sdk-go/pkg/dto/request"
	"github.com/yunhanshu-net/sdk-go/pkg/dto/response"
)

const (
	StatusConnecting = "connecting"
	StatusRunning    = "running"
	StatusClosed     = "closed"
)

type Runner interface {
	coder.Coder
	IsRunning() bool
	Connect(ctx context.Context, conn *nats.Conn) error
	Close() error
	GetInfo() *runnerproject.Runner
	GetID() string
	GetStatus() string
	Request(ctx context.Context, req *request.RunFunctionReq) (*response.RunFunctionResp, error)
}

func NewRunner(runner runnerproject.Runner) (Runner, error) {

	if runner.Version == "" {
		version, err := runner.GetCurrentVersion()
		if err != nil {
			return nil, err
		}
		runner.Version = version
	}

	runnerCoder, err := coder.NewCoder(&runner)
	if err != nil {
		return nil, err
	}

	cmd := &cmdRunner{
		Coder:          runnerCoder,
		qpsWindow:      make(map[int64]uint),
		qpsLock:        &sync.Mutex{},
		id:             uuid.NewString(),
		detail:         &runner,
		connectLock:    &sync.Mutex{},
		connectingLock: &sync.Mutex{},
		status:         StatusClosed,
		connected:      false}

	if runner.Kind == "cmd" {
		return cmd, nil
	}
	return cmd, nil
}

type cmdRunner struct {
	id        string
	detail    *runnerproject.Runner
	connected bool
	coder.Coder
	qpsLock        *sync.Mutex
	qpsWindow      map[int64]uint
	latestHandelTs time.Time

	natsConn       *nats.Conn
	process        *os.Process
	status         string
	connectLock    *sync.Mutex
	connectingLock *sync.Mutex
}

func (r *cmdRunner) GetStatus() string {
	return r.status
}

func (r *cmdRunner) IsRunning() bool {
	return r.status == StatusRunning
}

func (r *cmdRunner) Connect(ctx context.Context, conn *nats.Conn) error {
	return r.connectNats(ctx, conn)
}

func (r *cmdRunner) connectNats(ctx context.Context, conn *nats.Conn) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	now := time.Now()

	lock := r.connectLock.TryLock()
	if !lock {
		return nil
	}
	if lock && r.connected {
		r.connectLock.Unlock()
		logger.Infof(ctx, "未启动连接:%s", r.detail.GetRequestSubject())
		return nil
	}
	if r.status == StatusConnecting {
		return nil
	}

	r.natsConn = conn
	r.status = StatusConnecting
	connectMsgCh := make(chan *nats.Msg, 1)
	subscribe, err := conn.ChanSubscribe(r.GetID(), connectMsgCh)
	if err != nil {
		r.status = StatusClosed
		return fmt.Errorf("NATS订阅失败: %w", err)
	}
	defer subscribe.Unsubscribe()
	defer r.connectLock.Unlock()

	runner := r.detail
	// Linux和macOS可以直接使用 && 连接命令
	args := []string{
		"./" + runner.GetBuildRunnerCurrentVersionName(),
		"connect",
		"--runner_id",
		r.GetID(),
	}

	go func() {
		_, cmd, err := cmdx.Run(ctx, runner.GetBinPath(), args)
		if err != nil {
			logger.Errorf(ctx, "connect nats %s", err.Error())
			return
		}
		r.process = cmd.Process
	}()

	select {
	case <-ctx.Done():
		r.status = StatusClosed
		if r.process != nil {
			r.process.Kill()
		}
		return ctx.Err()
	case <-time.After(time.Second * 5):
		r.status = StatusClosed
		if r.process != nil {
			r.process.Kill()
		}
		return fmt.Errorf("连接 %+v 超时", runner)
	case msg := <-connectMsgCh:
		newMsg := nats.NewMsg(msg.Subject)
		newMsg.Header.Set("code", "0")
		err := msg.RespondMsg(newMsg)
		if err != nil {
			return fmt.Errorf("响应连接消息失败: %w", err)
		}
		if msg.Header.Get("code") != "0" {
			return fmt.Errorf("连接 %+v 失败: %s", runner, msg.Header.Get("msg"))
		}
		logger.Infof(ctx, "runner: %s 启动成功, 耗时: %s", runner.GetRequestSubject(), time.Since(now))
		r.status = StatusRunning
		r.connected = true
	}
	return nil
}

func (r *cmdRunner) GetInfo() *runnerproject.Runner {
	return r.detail
}

func (r *cmdRunner) GetID() string {
	return r.id
}

func (r *cmdRunner) Close() error {
	if r.connected {
		r.connected = false
		r.connectLock.Lock()
		defer r.connectLock.Unlock()
		r.status = StatusClosed
	}
	return nil
}

func (r *cmdRunner) requestByFile(ctx context.Context, req *request.RunFunctionReq) (*response.RunFunctionResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	fileName := strconv.Itoa(int(time.Now().UnixMicro())) + ".json"
	requestJsonPath := r.detail.GetBinPath() + "/.request/" + fileName
	binPath := r.detail.GetBinPath()
	req.RunnerID = r.id
	req.Runner = r.detail
	err := jsonx.SaveFile(requestJsonPath, req)
	if err != nil {
		logger.Errorf(ctx, "保存请求文件失败: path=%s, error=%v", requestJsonPath, err)
		return nil, err
	}
	args := []string{
		filepath.Join(binPath, r.detail.GetBuildRunnerCurrentVersionName()),
		"run",
		"--file",        // 标志名
		requestJsonPath, // 值
		"--method",
		req.Method,
		"--router",
		req.Router,
		"--trace_id",
		req.TraceID,
	}
	cc := strings.Join(args, " ")

	run, _, err := cmdx.Run(ctx, binPath, args)
	if err != nil {
		logger.Infof(ctx, "执行命令如下: %s", cc)
		logger.Errorf(ctx, "cmdx.Run(ctx, binPath, args) err:%s", err)
		return nil, err
	}

	resList := stringsx.ParserHtmlTagContent(string(run), "Response")

	if len(resList) == 0 {
		logger.Errorf(ctx, "未找到Response标签")
		return nil, fmt.Errorf("请使用SDK开发软件，未找到正确的响应格式")
	}
	var res response.RunFunctionResp
	err = json.Unmarshal([]byte(resList[0]), &res)
	if err != nil {
		logger.Errorf(ctx, "解析响应JSON失败: error=%v", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &res, nil
}

func (r *cmdRunner) requestByNats(ctx context.Context, runnerRequest *request.RunFunctionReq) (*response.RunFunctionResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	var resp response.RunFunctionResp
	msg := nats.NewMsg(r.detail.GetRequestSubject())
	runnerRequest.Runner = r.detail
	marshal, err := json.Marshal(runnerRequest)
	if err != nil {
		return nil, err
	}
	msg.Data = marshal
	msg.Header.Set(constants.TraceID, runnerRequest.TraceID)
	respMsg, err := r.natsConn.RequestMsg(msg, time.Second*20)
	if err != nil {
		return nil, fmt.Errorf("NATS请求失败: %w", err)
	}
	if respMsg.Header.Get("code") != "0" {
		return nil, fmt.Errorf("业务错误: %s", respMsg.Header.Get("msg"))
	}

	err = json.Unmarshal(respMsg.Data, &resp)
	if err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}
	return &resp, nil
}

func (r *cmdRunner) shouldBeClose() bool {
	if time.Since(r.latestHandelTs).Seconds() > 5 {
		return true
	}
	return false
}

func (r *cmdRunner) Request(ctx context.Context, runnerRequest *request.RunFunctionReq) (*response.RunFunctionResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	//这里检查是否需要启动程序
	r.qpsLock.Lock()
	r.latestHandelTs = time.Now()
	r.qpsWindow[time.Now().Unix()]++
	r.qpsLock.Unlock()

	if !r.connected {
		one, err := r.requestByFile(ctx, runnerRequest)
		if err != nil {
			return nil, err
		}
		return one, nil
	} else {
		runnerRequest.RunnerID = r.GetID()
		//长连接
		rpc, err := r.requestByNats(ctx, runnerRequest)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") { //连接失效了
				logger.Warnf(ctx, "NATS连接已失效，尝试使用文件方式请求")
				return r.requestByFile(ctx, runnerRequest)
			}
			return nil, err
		}
		return rpc, nil
	}
}
