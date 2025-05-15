package scheduler

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/runcher/runner"
	"github.com/yunhanshu-net/runcher/runtime"
	"sync"
)

const (
	highQPSThreshold = 3 // 每秒3请求视为高并发
)

// Scheduler 调度器结构体
type Scheduler struct {
	RunnerRoot     string
	natsConn       *nats.Conn
	closeSub       *nats.Subscription
	coderSub       *nats.Subscription
	functionSub    *nats.Subscription
	runtimeRunners map[string]*runtime.Runners
	runnerLock     *sync.Mutex
	sockInfoLk     *sync.Mutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

func (s *Scheduler) closeRunner(path string) error {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	v, ok := s.runtimeRunners[path]
	if ok {
		for _, r := range v.Running {
			r.Close()
		}
	}
	return nil
}

// NewScheduler 创建新的调度器实例
func NewScheduler(conn *nats.Conn) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		RunnerRoot:     conf.GetRunnerRoot(),
		natsConn:       conn,
		runnerLock:     &sync.Mutex{},
		runtimeRunners: make(map[string]*runtime.Runners),
		sockInfoLk:     &sync.Mutex{},
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (s *Scheduler) stopRunner(runner *model.Runner) error {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	subject := runner.GetRequestSubject()
	v, ok := s.runtimeRunners[subject]
	if ok {
		for _, r := range v.Running {
			r.Close()
		}
	}
	return nil
}

//// Run 运行调度器
//func (s *Scheduler) Run() error {
//	// 订阅主题
//	sub, err := s.natsConn.Subscribe("scheduler.>", func(msg *nats.Msg) {
//		// 处理消息
//		// ...
//	})
//	if err != nil {
//		return fmt.Errorf("订阅主题失败: %w", err)
//	}
//
//	// 启动监控协程
//
//
//	return nil
//}

// Close 关闭调度器
func (s *Scheduler) Close() error {
	s.cancel()
	s.wg.Wait()
	for unix, v := range s.runtimeRunners {
		for _, r := range v.Running {
			err := r.Close()
			if err != nil {
				logger.Errorf(context.Background(), "runner:%s close err:%s", unix, err.Error())
			}
			logger.Infof(context.Background(), "runner:%s close success", unix)
		}
	}
	s.closeSub.Unsubscribe()
	s.coderSub.Unsubscribe()
	s.functionSub.Unsubscribe()
	return nil
}

func (s *Scheduler) getAndSetRunner(r *model.Runner) (*runtime.Runners, error) {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	name := r.GetRequestSubject()
	runtimeRunner, ok := s.runtimeRunners[name]
	if !ok {
		rn, err := runner.NewRunner(*r)
		if err != nil {
			return nil, err
		}
		runners := runtime.NewRunners(rn)
		runners.StartLock[rn.GetID()] = &sync.Mutex{}
		s.runtimeRunners[name] = runners
		s.runtimeRunners[name] = runners

		return runners, nil
	}
	return runtimeRunner, nil
}

func (s *Scheduler) getRunner(r *model.Runner) (runner.Runner, error) {
	setRunner, err := s.getAndSetRunner(r)
	if err != nil {
		return nil, err
	}
	return setRunner.GetOne(), nil
}

func (s *Scheduler) Request(ctx context.Context, request *request.RunnerRequest) (*response.Response, error) {

	//假如没带版本
	if request.Runner.Version == "" {
		version, err := request.Runner.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		request.Runner.Version = version
	}
	rt, err := s.getAndSetRunner(request.Runner)
	if err != nil {
		return nil, err
	}
	r := rt.GetOne()
	if r == nil {
		return nil, errors.New("runner not found")
	}
	if r.IsRunning() { //如果有运行中的实例，直接请求
		return r.Request(ctx, request.Request)
	}
	qps := rt.GetCurrentQps()
	rt.AddQps(1)

	if qps >= highQPSThreshold && r.GetStatus() == runner.StatusClosed { //如果不在启动中，那就启动
		//	启动连接
		lk := rt.StartLock[r.GetID()]
		lock := lk.TryLock()
		if lock { //加锁成功！
			logger.Infof(ctx, "当前qps：%v尝试启动连接", qps)
			err := r.Connect(ctx, s.natsConn)
			if err != nil {
				logger.Errorf(ctx, "连接启动失败：%+v err:%s", r.GetInfo(), err)
				return nil, err
			}
			lk.Unlock()
		}
	}

	if rt.GetCurrentQps() >= highQPSThreshold && r.GetStatus() == runner.StatusConnecting { //如果在启动中
		lk := rt.StartLock[r.GetID()]
		lk.Lock()
		lk.Unlock()
	}

	runnerResponse, err := r.Request(ctx, request.Request)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil
}
