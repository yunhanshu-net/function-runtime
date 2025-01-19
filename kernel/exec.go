package kernel

import (
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/store"
	"github.com/yunhanshu-net/runcher/runner"
	"time"
)

// Executor 引擎负责调度，管理和执行 各种Runner
type Executor struct {
	FileStore    store.FileStore
	runnerStatus map[string]*runnerStatus
}

type runnerStatus struct {
	running   bool
	startTime time.Time
}

func NewExecutor(fileStore store.FileStore) *Executor {
	return &Executor{FileStore: fileStore}
}

func (b *Executor) startKeepAlive() {

}

// Request 执行请求
func (b *Executor) Request(call *request.Request, runnerConf *model.Runner) (*response.Response, error) {
	newRunner := runner.NewRunner(runnerConf)
	//call.RunnerInfo.WorkPath = newRunner.GetInstallPath()        //软件安装目录
	//err := jsonx.SaveFile(call.RunnerInfo.RequestJsonPath, call) //todo 存储请求参数
	//if err != nil {
	//	return nil, err
	//}
	status, ok := b.runnerStatus[call.RunnerInfo.Key()]
	if ok {
		call.IsRunning = status.running
	}
	//todo 这里判断是否需要建立长连接

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
