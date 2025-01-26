package kernel

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/store"
	"github.com/yunhanshu-net/runcher/runner"
	"sync"
	"time"
)

// Executor 引擎负责调度，管理和执行 各种Runner
type Executor struct {
	FileStore    store.FileStore
	runnerStatus map[string]*runnerStatus
	nats         *nats.Conn
	connectSub   *nats.Subscription
	closeSub     *nats.Subscription

	RuncherConnectLock *sync.Mutex
	RuncherConnect     map[string]chan ConnectStatus
	RuncherCloseLock   *sync.Mutex
	RuncherClose       map[string]chan CloseStatus
}

type ConnectStatus struct {
	Success bool
	Message string
}

type CloseStatus struct {
	Success bool
	Message string
}

type runnerStatus struct {
	running         bool
	waitExecRequest []*request.Request
	startTime       time.Time
	mu              *sync.Mutex // 用于并发安全
}

// Push 方法将请求添加到 waitExecRequest 切片中
func (rs *runnerStatus) Push(req *request.Request) {
	rs.mu.Lock()         // 锁定以确保并发安全
	defer rs.mu.Unlock() // 解锁
	rs.waitExecRequest = append(rs.waitExecRequest, req)
}

// Pop 方法从 waitExecRequest 切片中弹出一个请求
func (rs *runnerStatus) Pop() *request.Request {
	rs.mu.Lock()         // 锁定以确保并发安全
	defer rs.mu.Unlock() // 解锁
	if len(rs.waitExecRequest) == 0 {
		return nil // 如果切片为空，返回 nil
	}
	// 获取最后一个请求并移除它
	req := rs.waitExecRequest[len(rs.waitExecRequest)-1]
	rs.waitExecRequest = rs.waitExecRequest[:len(rs.waitExecRequest)-1]
	return req
}

func NewExecutor(fileStore store.FileStore) *Executor {
	return &Executor{
		FileStore:          fileStore,
		RuncherConnectLock: &sync.Mutex{},
		RuncherCloseLock:   &sync.Mutex{},
		runnerStatus:       make(map[string]*runnerStatus),
		RuncherConnect:     make(map[string]chan ConnectStatus),
		RuncherClose:       make(map[string]chan CloseStatus),
	}
}

func (b *Executor) startKeepAlive() {

}

type RunnerStatus struct {
	Running bool
}

func (b *Executor) RunnerStatus(runnerKey string) RunnerStatus {
	status, ok := b.runnerStatus[runnerKey]
	if ok {
		return RunnerStatus{Running: status.running}
	}
	return RunnerStatus{}
}

func (b *Executor) ShouldStartKeepAlive() bool {
	return false
}

// Request 执行请求
func (b *Executor) Request(call *request.Request, runnerConf *model.Runner) (*response.Response, error) {
	newRunner := runner.NewRunner(runnerConf)
	//call.RunnerInfo.WorkPath = newRunner.GetInstallPath()        //软件安装目录
	//err := jsonx.SaveFile(call.RunnerInfo.RequestJsonPath, call) //todo 存储请求参数
	//if err != nil {
	//	return nil, err
	//}
	//status, ok := b.runnerStatus[call.RunnerInfo.Key()]
	//if ok {
	//	call.IsRunning = status.running
	//}
	call.IsRunning = b.RunnerStatus(call.RunnerInfo.Key()).Running

	//todo 这里判断是否需要建立长连接
	err := newRunner.StartKeepAlive(context.Background())
	rspCall, err := newRunner.Request(call)
	if err != nil {
		return nil, err
	}
	return rspCall, nil
}

// Install 安装软件
func (b *Executor) Install(runnerConf *model.Runner) (*response.InstallInfo, error) {
	newRunner := runner.NewRunner(runnerConf)
	info, err := newRunner.Install(b.FileStore)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// UpdateVersion 更新软件
func (b *Executor) UpdateVersion(updateRunner *model.UpdateVersion) (*response.UpdateVersion, error) {
	newRunner := runner.NewRunner(updateRunner.RunnerConf)
	info, err := newRunner.UpdateVersion(updateRunner, b.FileStore)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (b *Executor) Close() error {
	b.closeSub.Unsubscribe()
	b.connectSub.Unsubscribe()
	return nil
}
