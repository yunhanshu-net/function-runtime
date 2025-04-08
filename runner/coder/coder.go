package coder

import (
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
)

type Coder interface {
	AddBizPackage(codeBizPackage *codex.BizPackage) error
	AddApi(codeApi *codex.CodeApi) error
	AddApis(codeApis []*codex.CodeApi) (errs []*codex.CodeApiCreateInfo, err error)
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
