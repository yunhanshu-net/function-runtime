package scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"github.com/yunhanshu-net/runcher/runtime"
	"strings"
	"sync"
)

const (
	highQPSThreshold = 3 // 每秒3请求视为高并发
)

type Scheduler struct {
	RunnerRoot     string
	natsConn       *nats.Conn
	closeSub       *nats.Subscription
	coderSub       *nats.Subscription
	runtimeRunners map[string]*runtime.Runners
	runnerLock     *sync.Mutex
	sockInfoLk     *sync.Mutex
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
		RunnerRoot:     conf.GetRunnerRoot(),
		natsConn:       conn,
		runnerLock:     &sync.Mutex{},
		runtimeRunners: make(map[string]*runtime.Runners),
		sockInfoLk:     &sync.Mutex{},
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
			logrus.Errorf("runner:%s close err:%s", m.GetRequestSubject(), err.Error())
			return
		}
		logrus.Infof("runner:%s close success", m.GetRequestSubject())

	})
	if err != nil {
		return err
	}
	s.closeSub = subscribe

	coderSub, err := s.natsConn.Subscribe("coder.>", func(msg *nats.Msg) {
		subjects := strings.Split(msg.Subject, ".")
		subject := subjects[1]
		if subject == "add_api" {
			s.AddApiByNats(msg)
		}

		if subject == "add_apis" {
			s.AddApiByNats(msg)
		}

	})
	if err != nil {
		return err
	}
	s.coderSub = coderSub

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

func (s *Scheduler) getAndSetRunner(r *model.Runner) *runtime.Runners {
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

func (s *Scheduler) getRunner(r *model.Runner) runner.Runner {
	return s.getAndSetRunner(r).GetOne()
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
	rt := s.getAndSetRunner(request.Runner)
	r := rt.GetOne()
	if r == nil {
		return nil, errors.New("runner not found")
	}
	if r.IsRunning() { //如果有运行中的实例，直接请求
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

func (s *Scheduler) SysCallback(callbackType string, r *model.Runner, body interface{}) (interface{}, error) {

	runnerRequest := &request.Request{
		Route:  "_sysCallback",
		Method: "POST",
		Body: &syscallback.Request{
			CallbackType: callbackType,
			Data:         body,
		},
	}
	runnerIns := s.getRunner(r)
	rsp, err := runnerIns.Request(context.Background(), runnerRequest)
	if err != nil {
		return nil, err
	}
	switch callbackType {
	case "SysOnVersionChange":
		decodeBody, err := response.DecodeBody[*syscallback.ResponseWith[*syscallback.SysOnVersionChangeResp]](rsp)
		if err != nil {
			return nil, err
		}
		if decodeBody.Code != 0 {
			return nil, fmt.Errorf(decodeBody.Msg)
		}
		//*syscallback.SysOnVersionChangeResp
		return decodeBody.Data, nil
	}

	return nil, fmt.Errorf("callbackType not found")
}
