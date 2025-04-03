package coder

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
)

type Coder interface {
	AddBizPackage(codeBizPackage *codex.BizPackage) error
	AddApi(codeApi *codex.CodeApi) error
	CreateProject() error
}

func NewCoder(runner *model.Runner) (Coder, error) {
	switch runner.Language {
	case "go":
		return &Golang{runnerRoot: conf.GetRunnerRoot(), runner: runner}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", runner.Language)
	}
}
