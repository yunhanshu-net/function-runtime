package coder

import (
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
)

type Coder interface {
	AddBizPackage(codeBizPackage *coder.BizPackage) error
	AddApi(codeApi *coder.CodeApi) error
	AddApis(codeApis []*coder.CodeApi) (errs []*coder.CodeApiCreateInfo, err error)
	CreateProject() error
}

func NewCoder(runner *model.Runner) (Coder, error) {
	switch runner.Language {
	case "go":
		return &Golang{runnerRoot: conf.GetRunnerRoot(), runner: runner}, nil
	default:
		return &Golang{runnerRoot: conf.GetRunnerRoot(), runner: runner}, nil
	}
}
