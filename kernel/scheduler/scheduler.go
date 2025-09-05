package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/function-go/pkg/dto/request"
	"github.com/yunhanshu-net/function-go/pkg/dto/response"
	"github.com/yunhanshu-net/function-runtime/conf"
	"github.com/yunhanshu-net/function-runtime/runner"
	"github.com/yunhanshu-net/function-runtime/runtime"
	"github.com/yunhanshu-net/pkg/constants"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/logger"
	"sync"
)

const (
	highQPSThreshold = 0 // 每秒3请求视为高并发
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

	msgCenter *MessageCenter
}

// NewScheduler 创建新的调度器实例
func NewScheduler(conn *nats.Conn) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		msgCenter: &MessageCenter{
			natsConn: conn,
		},
		RunnerRoot:     conf.GetRunnerRoot(),
		natsConn:       conn,
		runnerLock:     &sync.Mutex{},
		runtimeRunners: make(map[string]*runtime.Runners),
		sockInfoLk:     &sync.Mutex{},
		ctx:            ctx,
		cancel:         cancel,
	}

	//订阅来自FunctionServer的消息,server 发起运行请求
	err := s.msgCenter.SubFunctionRuntimeReceiver(func(msg *nats.Msg) {

		//回复FunctionServer
		newMsg := nats.NewMsg("function-server.receiver")
		err := s.msgCenter.PublishFunctionServer(newMsg)
		if err != nil {
			panic(err)
		}
		ctx := context.WithValue(context.Background(), constants.TraceID, msg.Header.Get(constants.TraceID))
		//logger.Infof(ctx, "function.run >%s uid:%s", msg.Subject, string(msg.Data))
		//接收runner关闭
		logger.Infof(ctx, "Got message: %s", string(msg.Data))
		runner, err := runnerproject.NewRunner(msg.Header.Get("user"), msg.Header.Get("runner"), conf.GetRunnerRoot(), msg.Header.Get("version"))
		if err != nil {
			panic(err)
		}

		req := request.RunFunctionReq{
			Method:  msg.Header.Get("method"),
			Router:  msg.Header.Get("router"),
			TraceID: msg.Header.Get(constants.TraceID),
			Runner:  runner,
			//BodyType: "string",
		}
		if req.IsMethodGet() {
			req.UrlQuery = msg.Header.Get("url_query")
		} else {
			req.Body = string(msg.Data)
			req.BodyType = "string"
		}
		rsp := nats.NewMsg(msg.Subject)
		rsp.Header.Set("code", "0")
		response, err := s.Request(ctx, &req)
		if err != nil {
			rsp.Header.Set("code", "-1")
			rsp.Header.Set("msg", err.Error())
			err = msg.RespondMsg(rsp)
			if err != nil {
				logger.Error(ctx, "request error", err)
			}
			return
		}
		for k, v := range response.MetaData {
			rsp.Header.Set(k, fmt.Sprintf("%v", v))
		}
		marshal, err := json.Marshal(response)
		if err != nil {
			logger.Error(ctx, "response marshal error", err)
			panic(err)
		}
		rsp.Data = marshal

		//这里回复
		err = msg.RespondMsg(rsp)
		if err != nil {
			logger.Error(ctx, "request error", err)
			return
		}

	})
	if err != nil {
		return nil
	}
	scheduler = s
	return s
}

func (s *Scheduler) stopRunner(runner *runnerproject.Runner) error {
	s.runnerLock.Lock()
	defer s.runnerLock.Unlock()
	subject := runner.GetRequestSubject()
	v, ok := s.runtimeRunners[subject]
	if ok {
		for _, r := range v.Running {
			err := r.Close()
			if err != nil {
				fmt.Printf("close runner error: %s\n", err)
			}
		}
	}
	return nil
}

// Close 关闭调度器
func (s *Scheduler) Close() error {
	//todo 这里要runner自己释放，不能强杀，因为runner可能还在处理任务，需要把释放权交给runner，优雅退出
	//s.cancel()
	//s.wg.Wait()
	//for unix, v := range s.runtimeRunners {
	//	for _, r := range v.Running {
	//		err := r.Close()
	//		if err != nil {
	//			logger.Errorf(context.Background(), "runner:%s close err:%s", unix, err.Error())
	//		}
	//		logger.Infof(context.Background(), "runner:%s close success", unix)
	//	}
	//}
	s.closeSub.Unsubscribe()
	s.coderSub.Unsubscribe()
	s.functionSub.Unsubscribe()
	s.msgCenter.natsSub.Unsubscribe()
	return nil
}

func (s *Scheduler) getAndSetRunner(r *runnerproject.Runner) (*runtime.Runners, error) {
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

func (s *Scheduler) getRunner(r *runnerproject.Runner) (runner.Runner, error) {
	setRunner, err := s.getAndSetRunner(r)
	if err != nil {
		return nil, err
	}
	return setRunner.GetOne(), nil
}

func (s *Scheduler) Request(ctx context.Context, request *request.RunFunctionReq) (*response.RunFunctionResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
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
		return r.Request(ctx, request)
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

	runnerResponse, err := r.Request(ctx, request)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil
}
func (s *Scheduler) RequestSync(ctx context.Context, request *request.RunFunctionReq) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	rt, err := s.getAndSetRunner(request.Runner)
	if err != nil {
		return err
	}
	r := rt.GetOne()
	if r == nil {
		return errors.New("runner not found")
	}
	if r.IsRunning() { //如果有运行中的实例，直接请求
		return r.RequestSync(ctx, request)
	}
	qps := rt.GetCurrentQps()
	rt.AddQps(1)

	if r.GetStatus() == runner.StatusClosed { //如果不在启动中，那就启动
		//	启动连接
		lk := rt.StartLock[r.GetID()]
		lock := lk.TryLock()
		if lock { //加锁成功！
			logger.Infof(ctx, "当前qps：%v尝试启动连接", qps)
			err := r.Connect(ctx, s.natsConn)
			if err != nil {
				logger.Errorf(ctx, "连接启动失败：%+v err:%s", r.GetInfo(), err)
				return err
			}
			lk.Unlock()
		}
	}

	if r.GetStatus() == runner.StatusConnecting { //如果在启动中
		lk := rt.StartLock[r.GetID()]
		lk.Lock()
		lk.Unlock()
	}

	return r.RequestSync(ctx, request)
}
