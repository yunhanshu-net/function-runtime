package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/jsonx"
	"github.com/yunhanshu-net/runcher/pkg/stringsx"
	"github.com/yunhanshu-net/runcher/runner/coder"
)

const (
	StatusConnecting = "connecting"
	StatusRunning    = "running"
	StatusClosed     = "closed"
)

type Runner interface {
	coder.Coder
	IsRunning() bool
	Connect(conn *nats.Conn) error
	Close() error
	GetInfo() *model.Runner
	GetID() string
	GetStatus() string
	Request(ctx context.Context, req *request.Request) (*response.Response, error)
}

func NewRunner(runner model.Runner) Runner {
	runnerCoder, _ := coder.NewCoder(&runner)

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
		return cmd
	}
	return cmd
}

type cmdRunner struct {
	id        string
	detail    *model.Runner
	connected bool
	coder.Coder
	qpsLock        *sync.Mutex
	qpsWindow      map[int64]uint
	latestHandelTs time.Time

	natsConn    *nats.Conn
	process     *os.Process
	status      string
	connectLock *sync.Mutex

	connectingLock *sync.Mutex
}

func (r *cmdRunner) GetStatus() string {
	return r.status
}

func (r *cmdRunner) IsRunning() bool {
	return r.status == StatusRunning
}

func (r *cmdRunner) Connect(conn *nats.Conn) error {
	return r.connectNats(conn)
}

func (r *cmdRunner) connectNats(conn *nats.Conn) error {
	now := time.Now()

	lock := r.connectLock.TryLock()
	if !lock {
		return nil
	}
	if lock && r.connected {
		r.connectLock.Unlock()
		logrus.Infof("未启动连接:%s", r.detail.GetRequestSubject())
		return nil
	}
	if r.status == StatusConnecting {
		return nil
	}

	r.natsConn = conn
	r.status = StatusConnecting
	connectMsgCh := make(chan *nats.Msg, 1)
	subscribe, err := conn.ChanSubscribe(r.id, connectMsgCh)
	if err != nil {
		r.status = StatusClosed
		return fmt.Errorf("NATS订阅失败: %w", err)
	}
	defer subscribe.Unsubscribe()
	defer r.connectLock.Unlock()

	runner := r.detail

	req := request.RunnerRequest{
		Runner:          runner,
		UUID:            r.id,
		TransportConfig: &request.TransportConfig{IdleTime: 10},
		Request:         nil,
	}

	path := runner.GetRequestPath() + "/" + uuid.New().String() + ".json"
	err = jsonx.SaveFile(path, req)
	if err != nil {
		return fmt.Errorf("保存请求文件失败: %w", err)
	}

	cc := fmt.Sprintf("cd %s && ./%s _connect %s", runner.GetBinPath(), runner.GetBuildRunnerCurrentVersionName(), path)
	// Linux和macOS可以直接使用 && 连接命令
	cmd := exec.Command("sh", "-c", cc)
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("命令执行失败: %s", err.Error())
		return fmt.Errorf("启动runner失败: %w", err)
	}
	r.process = cmd.Process

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
		logrus.Infof("runner: %s 启动成功, 耗时: %s", runner.GetRequestSubject(), time.Since(now))
		r.status = StatusRunning
		r.connected = true
	}
	return nil
}

func (r *cmdRunner) GetInfo() *model.Runner {
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

func (r *cmdRunner) requestByFile(req *request.Request) (*response.Response, error) {
	fileName := strconv.Itoa(int(time.Now().UnixMicro())) + ".json"
	requestJsonPath := r.detail.GetBinPath() + "/.request/" + fileName
	binPath := r.detail.GetBinPath()
	reqCall := request.RunnerRequest{
		Request: req,
		Runner:  r.detail,
	}
	err := jsonx.SaveFile(requestJsonPath, reqCall)
	if err != nil {
		return nil, err
	}

	cc := fmt.Sprintf("cd %s && ./%s %s .request/%s",
		binPath, r.detail.GetBuildRunnerCurrentVersionName(), req.Route, fileName)
	// Linux和macOS可以直接使用 && 连接命令
	cmd := exec.Command("sh", "-c", cc)
	logrus.Debugf("执行命令: %s", cc)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		logrus.Errorf("命令执行失败: %s, 命令: %s", err.Error(), cc)
		return nil, err
	}
	outString := out.String()
	if outString == "" {
		return nil, fmt.Errorf("命令输出为空，请检查程序是否正确")
	}

	resList := stringsx.ParserHtmlTagContent(outString, "Response")

	if len(resList) == 0 {
		return nil, fmt.Errorf("请使用SDK开发软件，未找到正确的响应格式")
	}
	var res response.Response
	err = json.Unmarshal([]byte(resList[0]), &res)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &res, nil
}

func (r *cmdRunner) requestByNats(runnerRequest *request.Request) (*response.Response, error) {
	req := &request.RunnerRequest{Request: runnerRequest, Runner: r.detail}
	var resp response.Response
	msg := nats.NewMsg(r.detail.GetRequestSubject())
	msg.Header.Set("body", runnerRequest.BodyString)
	marshal, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	msg.Data = marshal
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

func (r *cmdRunner) Request(ctx context.Context, runnerRequest *request.Request) (*response.Response, error) {
	//这里检查是否需要启动程序
	r.qpsLock.Lock()
	r.latestHandelTs = time.Now()
	r.qpsWindow[time.Now().Unix()]++
	r.qpsLock.Unlock()

	if !r.connected {
		one, err := r.requestByFile(runnerRequest)
		if err != nil {
			return nil, err
		}
		return one, nil
	} else {
		//长连接
		rpc, err := r.requestByNats(runnerRequest)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") { //连接失效了
				logrus.Warnf("NATS连接已失效，尝试使用文件方式请求")
				return r.requestByFile(runnerRequest)
			}
			return nil, err
		}
		return rpc, nil
	}
}
