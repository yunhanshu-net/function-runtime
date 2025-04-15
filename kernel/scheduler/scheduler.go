package scheduler

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
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
	natsConn   *nats.Conn
	closeSub   *nats.Subscription
	connectSub *nats.Subscription
	//natsConn *nats.Conn
	runtimeRunner   map[string]runner.Runner
	runnerLock      *sync.Mutex
	sockRuntimeInfo map[string]*sockRuntimeInfo
	sockInfoLk      *sync.Mutex
}

func (s *Scheduler) closeRunner(path string) error {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	v, ok := s.runtimeRunner[path]
	if ok {
		return v.Close()
	}
	return nil
}

func NewScheduler(conn *nats.Conn) *Scheduler {
	return &Scheduler{
		natsConn:        conn,
		runnerLock:      &sync.Mutex{},
		runtimeRunner:   make(map[string]runner.Runner),
		sockRuntimeInfo: make(map[string]*sockRuntimeInfo),
		sockInfoLk:      &sync.Mutex{},
	}
}

func (s *Scheduler) Run() error {
	logrus.Infof("Scheduler Run")

	//group := uuid.New().String()
	//监听runner的启动和关闭事件
	subscribe, err := s.natsConn.Subscribe("close.runner", func(msg *nats.Msg) {
		logrus.Infof("runner.close >%s", msg.Subject)
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
	v, ok := s.runtimeRunner[subject]
	if ok {
		return v.Close()
	}
	return nil
}

func (s *Scheduler) Close() error {
	for unix, v := range s.runtimeRunner {
		err := v.Close()
		if err != nil {
			logrus.Errorf("runner:%s close err:%s", unix, err.Error())
		}
		logrus.Infof("runner:%s close success", unix)
	}
	return nil
}

func (s *Scheduler) getAndSetRunner(r *model.Runner) runner.Runner {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	name := r.GetRequestSubject()
	v, ok := s.runtimeRunner[name]
	if ok {
		return v
	}
	logrus.Infof("set runner")
	newRunner := runner.NewRunner(*r)
	s.runtimeRunner[name] = newRunner
	return newRunner
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
	sockRunner := request.Runner.GetRequestSubject()
	s.sockInfoLk.Lock()
	currentWindow, ok := s.sockRuntimeInfo[sockRunner]
	ts := time.Now().Unix()
	if !ok {
		currentWindow = &sockRuntimeInfo{latestHandelTs: time.Now(), qpsWindow: map[int64]uint{ts: 1}}
		s.sockRuntimeInfo[sockRunner] = currentWindow
	}

	qps, ok := currentWindow.qpsWindow[ts]
	if !ok {
		currentWindow.latestHandelTs = time.Now()
		currentWindow.qpsWindow[ts] = 1
	} else {
		currentWindow.latestHandelTs = time.Now()
		currentWindow.qpsWindow[ts]++
	}

	s.sockInfoLk.Unlock()
	r := s.getAndSetRunner(request.Runner)
	if r.IsRunning() {
		return r.Request(context.Background(), request.Request)
	}

	//logrus.Infof("当前qps：%v", qps)
	if qps >= highQPSThreshold && r.GetStatus() != runner.StatusConnecting { //如果不在启动中，那就启动
		//	启动连接
		logrus.Infof("当前qps：%v尝试启动连接", qps)
		err := r.Connect(s.natsConn)
		if err != nil {
			logrus.Errorf("连接启动失败：%+v err:%s", r.GetInfo(), err)
			return nil, err
		}
	}
	runnerResponse, err := r.Request(context.Background(), request.Request)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil
}
