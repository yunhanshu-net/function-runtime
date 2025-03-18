package scheduler

import (
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
	"os"
	"strings"
)

type Scheduler struct {
	RunnerRoot string
}

func NewDefaultScheduler() *Scheduler {
	runnerRoot := "./soft"
	if os.Getenv("RUNNER_ROOT") != "" {
		runnerRoot = strings.TrimSuffix(os.Getenv("RUNNER_ROOT"), "/") + "/soft"
	}

	return &Scheduler{
		RunnerRoot: runnerRoot,
	}
}

func (s *Scheduler) AddApi(runner *model.Runner, codeApi *codex.CodeApi) error {
	coder := NewCoder(runner.Language)
	err := coder.AddApi(s.RunnerRoot, runner, codeApi)
	if err != nil {
		return err
	}
	return nil
}
