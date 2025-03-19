package runner

import (
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/store"
	"github.com/yunhanshu-net/runcher/runner/coder"
	"sync"
)

type RuntimeInfo struct {
	RunCount int
	AvgQPS   int
	AvgMem   int
	Lock     *sync.Mutex
}

type Control struct {
	RunningRunners map[string]*RuntimeInfo
}

func NewRunner(runner *model.Runner) Runner {
	switch runner.ToolType {
	case "windows", "linux", "macos", "可执行程序(linux)", "可执行程序(windows)":
		return NewCmd(runner)
	case "website":
		return NewWebSite(runner)
	case "docker":
		//	todo 待实现
	case "python":
		//	todo 待实现

	}
	return NewCmd(runner)
}

// Runner runcher 引擎可以调度任何实现Runner 接口的程序（可执行程序|批处理文件|python代码|lua|ruby|nodejs|docker镜像|java jar包|so文件）
type Runner interface {
	Install(store store.FileStore) (installInfo *response.InstallInfo, err error)                             //安装程序
	GetInstallPath() (path string)                                                                            //获取安装路径
	UnInstall() (unInstallInfo *response.UnInstallInfo, err error)                                            //卸载
	UpdateVersion(up *model.UpdateVersion, fileStore store.FileStore) (*response.UpdateVersion, error)        //更新版本
	RollbackVersion(r *request.RollbackVersion, fileStore store.FileStore) (*response.RollbackVersion, error) //版本回滚
	Request(ctx *request.Context) (*response.RunnerResponse, error)                                           //运行程序

	GetUUID() string

	coder.Coder
	// StartKeepAlive ctx with uuid
	StartKeepAlive(ctx *request.Context) error
	Stop() error
	GetInstance() interface{}
}
