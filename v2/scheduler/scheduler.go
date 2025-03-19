package scheduler

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
	"github.com/yunhanshu-net/runcher/runner/coder"
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
	newCoder, err := coder.NewCoder(runner.Language)
	if err != nil {
		return err
	}
	err = newCoder.AddApi(s.RunnerRoot, runner, codeApi)
	if err != nil {
		return err
	}
	fmt.Println("add api success")
	return nil
}
