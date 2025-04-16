package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/jsonx"
	"github.com/yunhanshu-net/runcher/pkg/stringsx"
	"github.com/yunhanshu-net/runcher/runner/coder"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
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
	WaitReady()
	Request(ctx context.Context, req *request.Request) (*response.Response, error)
}

func NewRunner(runner model.Runner) Runner {
	runnerCoder, _ := coder.NewCoder(&runner)

	cmd := &cmdRunner{
		Coder:       runnerCoder,
		qpsWindow:   make(map[int64]uint),
		qpsLock:     &sync.Mutex{},
		id:          uuid.NewString(),
		detail:      &runner,
		close:       make(chan *protocol.Message),
		connectLock: &sync.Mutex{},
		status:      StatusClosed,
		connected:   false}

	cmd.cond = &sync.Cond{L: cmd.connectLock}

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

	natsConn *nats.Conn
	//nats        *nats.Conn
	conn        client.XClient
	process     *os.Process
	status      string //
	connectLock *sync.Mutex

	connectingLock *sync.Mutex

	cond  *sync.Cond
	close chan *protocol.Message
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

func (r *cmdRunner) WaitReady() {
	r.connectingLock.Lock()
	r.connectingLock.Unlock()
}

func (r *cmdRunner) connectUnixSock() error {
	lock := r.connectLock.TryLock()
	if !lock {
		return nil
	}
	if lock && r.connected {
		r.connectLock.Unlock()
		logrus.Infof("未启动连接:%s", r.detail.GetUnixFileName())
		return nil
	}

	r.status = StatusConnecting
	defer r.connectLock.Unlock()
	runner := r.detail

	req := request.RunnerRequest{
		Runner: runner,
		UUID:   r.id,
		TransportConfig: &request.TransportConfig{
			IdleTime: 10,
		},
		Request: nil,
	}
	now := time.Now()
	path := runner.GetRequestPath() + "/" + uuid.New().String() + ".json"
	err := jsonx.SaveFile(path, req)
	if err != nil {
		return err
	}

	cc := fmt.Sprintf("cd %s && ./%s _connect %s", runner.GetBinPath(), runner.GetBuildRunnerCurrentVersionName(), path)
	// Linux和macOS可以直接使用 && 连接命令
	cmd := exec.Command("sh", "-c", cc)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Errorf("cmd StdoutPipe err:%s", err.Error())
		return err
	}
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("cmd run err:%s", err.Error())
		return err
	}
	r.process = cmd.Process
	scanner := bufio.NewScanner(stdoutPipe)
	// 主进程等待结果或超时
	for {
		select {
		case <-time.After(10 * time.Second):
			cmd.Process.Kill()
			fmt.Println("超时：未检测到连接成功")
			return fmt.Errorf("connect timeout")
		default:

			//todo 这里有bug，需要改进，下面scanner.Scan()会阻塞，导致上面time.After超时了也不会触发的
			if scanner.Scan() {
				line := scanner.Text()
				// 检测到目标字符串后触发操作
				if line == "<connect-ok></connect-ok>" {
					unixPath := runner.GetUnixPathFile()
					t1 := time.Now()

					d, err := client.NewPeer2PeerDiscovery("unix@"+unixPath, "")
					if err != nil {
						return err
					}
					option := client.DefaultOption
					option.SerializeType = protocol.JSON
					r.conn = client.NewBidirectionalXClient("Rpc", client.Failtry, client.RandomSelect, d, option, r.close)
					err = r.conn.Call(context.Background(), "Ping", &request.Ping{}, &request.Ping{})
					if err != nil {
						panic(err)
					}
					r.status = StatusRunning
					r.connected = true
					go func() {
						ticker := time.NewTicker(time.Second * 20)
						for {
							select {
							case v := <-r.close:
								r.connected = false
								r.status = StatusClosed
								logrus.Infof("服务端已关闭连接，客户端监听到消息已经关闭该连接 Payload:%s,metadata:%s", string(v.Payload), v.Metadata)
								r.conn = nil
								return
							case <-ticker.C:
								if r.shouldBeClose() {
									logrus.Infof("runcher主动关闭连接")
									err := r.Close()
									if err != nil {
										logrus.Error(err.Error())
									}
									return
								}
							}
						}
					}()

					logrus.Infof("连接启动成功：%s total-cost:%s net:%s",
						r.detail.GetUnixFileName(),
						time.Since(now).String(),
						time.Since(t1).String())
					// 通知主流程继续
					//return // 结束监听
					return nil
				}
			} else {
				err := scanner.Err()
				if err != nil {
					cmd.Process.Kill()
					return err
				}
			}
		}
	}

	return nil
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
		panic(err)
		return err
	}
	defer subscribe.Unsubscribe()

	defer r.connectLock.Unlock()
	runner := r.detail
	//path :=runner.GetUnixPath()

	req := request.RunnerRequest{
		Runner: runner,
		UUID:   r.id,
		TransportConfig: &request.TransportConfig{
			IdleTime: 10,
		},
		Request: nil,
	}
	//now := time.Now()
	path := runner.GetRequestPath() + "/" + uuid.New().String() + ".json"
	err = jsonx.SaveFile(path, req)
	if err != nil {
		panic(err)
		return err
	}

	cc := fmt.Sprintf("cd %s && ./%s _connect %s", runner.GetBinPath(), runner.GetBuildRunnerCurrentVersionName(), path)
	// Linux和macOS可以直接使用 && 连接命令
	cmd := exec.Command("sh", "-c", cc)
	err = cmd.Start()
	if err != nil {
		logrus.Errorf("cmd run err:%s", err.Error())
		panic(err)
		return err
	}
	r.process = cmd.Process

	select {
	case <-time.After(time.Second * 5):
		panic("time out")
		return fmt.Errorf("connect %+v timeout", runner)
	case msg := <-connectMsgCh:
		newMsg := nats.NewMsg(msg.Subject)
		newMsg.Header.Set("code", "0")
		err := msg.RespondMsg(newMsg)
		if err != nil {
			panic(err)
		}
		if msg.Header.Get("code") != "0" {
			return fmt.Errorf("connectNats connect %+v err:%s", runner, msg.Header.Get("msg"))
		}
		logrus.Infof("runner:%s 启动成功：cost:%s", runner.GetRequestSubject(), time.Now().Sub(now))
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

func (r *cmdRunner) closeReq() error {
	req := &request.RunnerRequest{Request: nil, Runner: nil}
	resp := &response.Response{}
	err := r.conn.Call(context.Background(), "Close", req, resp)
	if err != nil {
		return err
	}
	return nil
}

//func (r *cmdRunner) Close() error {
//	if r.connected {
//		r.connected = false
//		r.connectLock.Lock()
//		defer r.connectLock.Unlock()
//		r.status = StatusClosed
//		err := r.closeReq()
//		if err != nil {
//			panic(err)
//		}
//		r.conn.Close()
//		//最好把unix sock文件也删除了
//	}
//	return nil
//}

func (r *cmdRunner) Close() error {
	if r.connected {
		r.connected = false
		r.connectLock.Lock()
		defer r.connectLock.Unlock()
		r.status = StatusClosed
		//err := r.closeReq()
		//if err != nil {
		//	panic(err)
		//}
		//r.conn.Close()
		//最好把unix sock文件也删除了
	}
	return nil
}

func (r *cmdRunner) requestByFile(req *request.Request) (*response.Response, error) {
	fileName := strconv.Itoa(int(time.Now().UnixMicro())) + ".json"
	//req.Runner.WorkPath = c.GetBinPath() //软件安装目录
	requestJsonPath := r.detail.GetBinPath() + "/.request/" + fileName
	binPath := r.detail.GetBinPath()
	reqCall := request.RunnerRequest{
		Request: req,
		Runner:  r.detail,
	}
	err := jsonx.SaveFile(requestJsonPath, reqCall) //todo 存储请求参数
	if err != nil {
		return nil, err
	}

	cc := fmt.Sprintf("cd %s && ./%s %s .request/%s",
		binPath, r.detail.GetBuildRunnerCurrentVersionName(), req.Route, fileName)
	// Linux和macOS可以直接使用 && 连接命令
	cmd := exec.Command("sh", "-c", cc)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		logrus.Errorf("cmd run err:%s cc=:%s", err.Error(), cc)
		return nil, err
	}
	outString := out.String()
	if outString == "" {
		return nil, fmt.Errorf("out.String() ==== nil cmd程序输出的结果为空，请检测程序是否正确")
	}

	resList := stringsx.ParserHtmlTagContent(outString, "Response")

	if len(resList) == 0 {
		return nil, fmt.Errorf("soft call err 请使用sdk开发软件")
	}
	var res response.Response
	err = json.Unmarshal([]byte(resList[0]), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *cmdRunner) requestByRpc(runnerRequest *request.Request) (*response.Response, error) {
	req := &request.RunnerRequest{Request: runnerRequest, Runner: r.detail}
	var resp response.Response
	err := r.conn.Call(context.Background(), "Call", req, &resp)
	if err != nil {
		logrus.Errorf("requestByRpc err:%s", err)
		return nil, err
	}
	return &resp, nil
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
		return nil, fmt.Errorf("requestByNats RequestMsg err:%s", err)
	}
	if respMsg.Header.Get("code") != "0" {
		return nil, fmt.Errorf("requestByNats RequestMsg biz err:%s", respMsg.Header.Get("msg"))
	}

	err = json.Unmarshal(respMsg.Data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *cmdRunner) shouldBeClose() bool {
	if time.Now().Sub(r.latestHandelTs).Seconds() > 5 {
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

			}
			return nil, err
		}
		return rpc, nil
	}
}
