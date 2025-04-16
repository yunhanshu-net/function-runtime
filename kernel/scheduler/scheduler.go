package scheduler

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"github.com/yunhanshu-net/runcher/runtime"
	"sync"
	"time"
)

const (
	highQPSThreshold = 3 // 每秒3请求视为高并发
)

type sockRuntimeInfo struct {
	qpsWindow      map[int64]uint
	latestHandelTs time.Time
}

func (s *sockRuntimeInfo) shouldClose() bool {
	if time.Now().Sub(s.latestHandelTs).Seconds() > 5 {
		return true
	}
	return false
}

type Scheduler struct {
	natsConn *nats.Conn
	closeSub *nats.Subscription
	//runtimeRunner   map[string]runner.Runner
	runtimeRunners  map[string]*runtime.Runners
	runnerLock      *sync.Mutex
	sockRuntimeInfo map[string]*sockRuntimeInfo
	sockInfoLk      *sync.Mutex
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

func NewScheduler(conn *nats.Conn) *Scheduler {
	return &Scheduler{
		natsConn:   conn,
		runnerLock: &sync.Mutex{},
		//runtimeRunner:   make(map[string]runner.Runner),
		runtimeRunners:  make(map[string]*runtime.Runners),
		sockRuntimeInfo: make(map[string]*sockRuntimeInfo),
		sockInfoLk:      &sync.Mutex{},
	}
}

func (s *Scheduler) Run() error {
	logrus.Infof("Scheduler Run")

	//group := uuid.New().String()
	//监听runner的启动和关闭事件
	subscribe, err := s.natsConn.Subscribe("close.runner", func(msg *nats.Msg) {
		logrus.Infof("runner.close >%s uid:%s", msg.Subject, string(msg.Data))
		//接收runner关闭
		var m model.Runner
		m.Version = msg.Header.Get("version")
		m.User = msg.Header.Get("user")
		m.Name = msg.Header.Get("name")
		err := s.stopRunner(&m)
		if err != nil {
			logrus.Errorf("runner:%s close err:%s", m.GetRequestSubject(), err.Error())
			return
		}
		rsp := nats.NewMsg(msg.Subject)
		rsp.Header.Set("code", "0")
		err = msg.RespondMsg(rsp)
		if err != nil {
			panic(err)
		}
		logrus.Infof("runner:%s close success", m.GetRequestSubject())

	})
	if err != nil {
		panic(err)
	}
	s.closeSub = subscribe

	return nil
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

func (s *Scheduler) Close() error {
	for unix, v := range s.runtimeRunners {

		for _, r := range v.Running {
			err := r.Close()
			if err != nil {
				logrus.Errorf("runner:%s close err:%s", unix, err.Error())
			}
			logrus.Infof("runner:%s close success", unix)
		}
	}
	s.closeSub.Unsubscribe()
	return nil
}

//func (s *Scheduler) getAndSetRunner(r *model.Runner) runner.Runner {
//	s.runnerLock.Lock()
//	defer s.runnerLock.Unlock()
//	name := r.GetRequestSubject()
//	v, ok := s.runtimeRunner[name]
//	if ok {
//		return v
//	}
//	logrus.Infof("set runner")
//	newRunner := runner.NewRunner(*r)
//	s.runtimeRunner[name] = newRunner
//	return newRunner
//}

func (s *Scheduler) addRunningRunner(r runner.Runner) {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	name := r.GetInfo().GetRequestSubject()
	v, ok := s.runtimeRunners[name]
	if !ok {
		s.runtimeRunners[name] = runtime.NewRunners(r)
		return
	}
	v.Running = append(v.Running, r)
	return
}

func (s *Scheduler) getAndSetRunner1(r *model.Runner) *runtime.Runners {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	name := r.GetRequestSubject()
	runtimeRunner, ok := s.runtimeRunners[name]
	if !ok {
		rn := runner.NewRunner(*r)
		runners := runtime.NewRunners(rn)
		runners.StartLock[rn.GetID()] = &sync.Mutex{}
		s.runtimeRunners[name] = runners
		s.runtimeRunners[name] = runners

		return runners
	}
	return runtimeRunner
}

func (s *Scheduler) Request(request *request.RunnerRequest) (*response.Response, error) {

	//假如没带版本
	if request.Runner.Version == "" {
		version, err := request.Runner.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		request.Runner.Version = version
	}
	rt := s.getAndSetRunner1(request.Runner)
	r := rt.GetOne()
	if r != nil && r.IsRunning() { //如果有运行中的实例，直接请求
		return r.Request(context.Background(), request.Request)
	}
	qps := rt.GetCurrentQps()
	rt.AddQps(1)

	if qps >= highQPSThreshold && r.GetStatus() == runner.StatusClosed { //如果不在启动中，那就启动
		//	启动连接
		lock := rt.StartLock[r.GetID()].TryLock()
		if lock { //加锁成功！
			logrus.Infof("当前qps：%v尝试启动连接", qps)
			err := r.Connect(s.natsConn)
			if err != nil {
				logrus.Errorf("连接启动失败：%+v err:%s", r.GetInfo(), err)
				return nil, err
			}
			rt.StartLock[r.GetID()].Unlock()
		}
	}

	if rt.GetCurrentQps() >= highQPSThreshold && r.GetStatus() == runner.StatusConnecting { //如果在启动中
		rt.StartLock[r.GetID()].Lock()
		rt.StartLock[r.GetID()].Unlock()
	}

	runnerResponse, err := r.Request(context.Background(), request.Request)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil
}
