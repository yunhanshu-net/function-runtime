package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/x/cmdx"
	"github.com/yunhanshu-net/runcher/pkg/constants"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/pkg/jsonx"
	"github.com/yunhanshu-net/runcher/pkg/stringsx"
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

	//req := request.RunFunctionReq{
	//	Runner: runner,
	//UUID:            r.id,
	//TransportConfig: &request.TransportConfig{IdleTime: 10},
	//Request:         nil,
	//}

	//path := runner.GetRequestPath() + "/" + uuid.New().String() + ".json"
	//err = jsonx.SaveFile(path, req)
	//if err != nil {
	//	return fmt.Errorf("保存请求文件失败: %w", err)
	//}

	//cc := fmt.Sprintf("cd %s && ./%s _connect %s", runner.GetBinPath(), runner.GetBuildRunnerCurrentVersionName(), path)
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

	//cmd := exec.Command("sh", "-c", cc)
	//err = cmd.Start()
	//if err != nil {
	//	logger.Errorf(ctx, "命令执行失败: %s", err.Error())
	//	return fmt.Errorf("启动runner失败: %w", err)
	//}

	select {
	case <-time.After(time.Second * 5):
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
		fmt.Printf("runner: %s 启动成功, 耗时: %s\n", runner.GetRequestSubject(), time.Since(now))
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
	//logger.Infof("开始通过文件方式处理请求: route=%s", req.Route)
	//logger.Debugf("请求内容: %+v", req)

	fileName := strconv.Itoa(int(time.Now().UnixMicro())) + ".json"
	requestJsonPath := r.detail.GetBinPath() + "/.request/" + fileName
	//binPath := r.detail.GetBinPath()
	binPath := r.detail.GetBinPath()
	//reqCall := request.RunnerRequest{Request: req.WithContext(ctx), Runner: r.detail}
	//req:=&request.RunFunctionReq{
	//	RunnerID: r.id,
	//	Runner: r.detail,
	//
	//}
	req.RunnerID = r.id
	req.Runner = r.detail

	//logger.Debugf("准备保存请求文件: path=%s", requestJsonPath)
	//reqJson, _ := json.MarshalIndent(reqCall, "", "  ")

	err := jsonx.SaveFile(requestJsonPath, req)
	if err != nil {
		logger.Errorf(ctx, "保存请求文件失败: path=%s, error=%v", requestJsonPath, err)
		return nil, err
	}
	//logger.Debugf("请求文件保存成功: path=%s", requestJsonPath)

	//cc := fmt.Sprintf("cd %s && ./%s %s .request/%s",
	//	binPath, r.detail.GetBuildRunnerCurrentVersionName(), req.Router, fileName)
	//// Linux和macOS可以直接使用 && 连接命令
	//cmd := exec.Command("sh", "-c", cc)
	//logger.Infof(ctx, "执行命令: \n%s\n req_body: %+v\n", cc, req.Body)
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//err = cmd.Run()
	//if err != nil {
	//	logger.Infof(ctx, "命令执行失败: error=%v, command=%s", err, cc)
	//	return nil, err
	//}
	//
	//outString := out.String()
	//if outString == "" {
	//	logger.Errorf(ctx, "命令输出为空")
	//	return nil, fmt.Errorf("命令输出为空，请检查程序是否正确")
	//}
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

	//resJson, _ := json.MarshalIndent(res, "", "  ")
	//logger.Infof(ctx, "响应内容: %s", string(resJson))
	return &res, nil
}

//func (r *cmdRunner) requestByFileV2(ctx context.Context, req *request.) (*response.Response, error) {
//	//logger.Infof("开始通过文件方式处理请求: route=%s", req.Route)
//	//logger.Debugf("请求内容: %+v", req)
//
//	fileName := strconv.Itoa(int(time.Now().UnixMicro())) + ".json"
//	requestJsonPath := r.detail.GetBinPath() + "/.request/" + fileName
//	binPath := r.detail.GetBinPath()
//	reqCall := request.RunnerRequest{Request: req.WithContext(ctx), Runner: r.detail}
//
//	//logger.Debugf("准备保存请求文件: path=%s", requestJsonPath)
//	//reqJson, _ := json.MarshalIndent(reqCall, "", "  ")
//
//	err := jsonx.SaveFile(requestJsonPath, reqCall)
//	if err != nil {
//		logger.Errorf(ctx, "保存请求文件失败: path=%s, error=%v", requestJsonPath, err)
//		return nil, err
//	}
//	//logger.Debugf("请求文件保存成功: path=%s", requestJsonPath)
//
//	cc := fmt.Sprintf("cd %s && ./%s %s .request/%s",
//		binPath, r.detail.GetBuildRunnerCurrentVersionName(), req.Route, fileName)
//	// Linux和macOS可以直接使用 && 连接命令
//	cmd := exec.Command("sh", "-c", cc)
//	logger.Infof(ctx, "执行命令: \n%s\n req_body: %+v\n", cc, req.Body)
//	var out bytes.Buffer
//	cmd.Stdout = &out
//	err = cmd.Run()
//	if err != nil {
//		logger.Infof(ctx, "命令执行失败: error=%v, command=%s", err, cc)
//		return nil, err
//	}
//
//	outString := out.String()
//	if outString == "" {
//		logger.Errorf(ctx, "命令输出为空")
//		return nil, fmt.Errorf("命令输出为空，请检查程序是否正确")
//	}
//
//	resList := stringsx.ParserHtmlTagContent(outString, "Response")
//
//	if len(resList) == 0 {
//		logger.Errorf(ctx, "未找到Response标签")
//		return nil, fmt.Errorf("请使用SDK开发软件，未找到正确的响应格式")
//	}
//
//	var res response.Response
//	err = json.Unmarshal([]byte(resList[0]), &res)
//	if err != nil {
//		logger.Errorf(ctx, "解析响应JSON失败: error=%v", err)
//		return nil, fmt.Errorf("解析响应失败: %w", err)
//	}
//
//	resJson, _ := json.MarshalIndent(res, "", "  ")
//	logger.Infof(ctx, "响应内容: %s", string(resJson))
//	return &res, nil
//}

func (r *cmdRunner) requestByNats(ctx context.Context, runnerRequest *request.RunFunctionReq) (*response.RunFunctionResp, error) {
	//req := &request.RunFunctionReq{Runner: r.detail}
	var resp response.RunFunctionResp
	msg := nats.NewMsg(r.detail.GetRequestSubject())
	//msg.Header.Set("body", runnerRequest.BodyString)
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
