package scheduler

import (
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/runtime"
	"sync"
)

type Scheduler struct {
	conn             *nats.Conn
	upstreamSub      *nats.Subscription
	manageSub        *nats.Subscription
	receiveRunnerSub *nats.Subscription
	runnerLock       map[string]*sync.RWMutex
	runnerLockLock   *sync.RWMutex
	lk               *sync.RWMutex
	runners          map[string]*runtime.Runners
	waitRunnerReady  map[string]*waitReady
}

//func NewDefaultScheduler() *Scheduler {
//	runnerRoot := "./soft"
//	if os.Getenv("RUNNER_ROOT") != "" {
//		runnerRoot = strings.TrimSuffix(os.Getenv("RUNNER_ROOT"), "/") + "/soft"
//	}
//
//	return &Scheduler{
//		RunnerRoot: runnerRoot,
//	}
//}
//
//func (s *Scheduler) AddApi(runner *model.Runner, codeApi *codex.CodeApi) error {
//	newCoder, err := coder.NewCoder(runner.Language)
//	if err != nil {
//		return err
//	}
//	err = newCoder.AddApi(s.RunnerRoot, runner, codeApi)
//	if err != nil {
//		return err
//	}
//	fmt.Println("add api success")
//	return nil
//}
