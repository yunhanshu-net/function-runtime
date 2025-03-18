package scheduler

import (
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
)

type Coder interface {
	AddApi(runnerRoot string, runner *model.Runner, codeApi *codex.CodeApi) error
}

func NewCoder(language string) Coder {
	switch language {
	case "go":
		return &Golang{}
	default:
		return nil
	}
}
