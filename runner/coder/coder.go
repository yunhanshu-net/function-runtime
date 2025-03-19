package coder

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
)

type Coder interface {
	AddApi(runnerRoot string, runner *model.Runner, codeApi *codex.CodeApi) error
}

func NewCoder(language string) (Coder, error) {
	switch language {
	case "go":
		return &Golang{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}
