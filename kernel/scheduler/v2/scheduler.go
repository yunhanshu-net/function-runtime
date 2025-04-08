package v2

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	v1 "github.com/yunhanshu-net/runcher/runner/v1"
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
	runtimeRunner   map[string]v1.Runner
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

func NewScheduler() *Scheduler {
	return &Scheduler{
		runnerLock:      &sync.Mutex{},
		runtimeRunner:   make(map[string]v1.Runner),
		sockRuntimeInfo: make(map[string]*sockRuntimeInfo),
		sockInfoLk:      &sync.Mutex{},
	}
}

func (s *Scheduler) Run() error {
	return nil
	//tk := time.NewTicker(time.Second * 5)
	//for {
	//	select {
	//	case <-tk.C:
	//		s.sockInfoLk.Lock()
	//		for path, info := range s.sockRuntimeInfo {
	//			if info.shouldClose() {
	//				err := s.closeRunner(path)
	//				if err != nil {
	//					logrus.Errorf("close runner:%s err:%s", path, err)
	//				}
	//			}
	//		}
	//		s.sockInfoLk.Unlock()
	//	}
	//}
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

func (s *Scheduler) getAndSetRunner(runner *model.Runner) v1.Runner {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	name := runner.GetUnixFileName()
	v, ok := s.runtimeRunner[name]
	if ok {
		return v
	}
	logrus.Infof("set runner")
	newRunner := v1.NewRunner(*runner)
	s.runtimeRunner[name] = newRunner
	return newRunner
}

func (s *Scheduler) Request(request *request.Request) (*response.RunnerResponse, error) {

	//假如没带版本
	if request.Runner.Version == "" {
		version, err := request.Runner.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		request.Runner.Version = version
	}
	sockRunner := request.Runner.GetUnixFileName()
	s.sockInfoLk.Lock()
	currentWindow, ok := s.sockRuntimeInfo[sockRunner]
	ts := time.Now().Unix()
	if !ok {
		currentWindow = &sockRuntimeInfo{
			latestHandelTs: time.Now(),
			qpsWindow:      map[int64]uint{ts: 1},
		}
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
	runner := s.getAndSetRunner(request.Runner)
	if runner.IsRunning() {
		return runner.Request(context.Background(), request.Request)
	}

	logrus.Infof("当前qps：%v", qps)
	if qps >= highQPSThreshold && runner.GetStatus() != v1.RunnerStatusConnecting { //如果不在启动中，那就启动
		//	启动连接
		logrus.Infof("当前qps：%v尝试启动连接", qps)
		err := runner.Connect()
		if err != nil {
			logrus.Errorf("连接启动失败：%+v err:%s", runner.GetInfo(), err)
			return nil, err
		}
	}
	runnerResponse, err := runner.Request(context.Background(), request.Request)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil
}
