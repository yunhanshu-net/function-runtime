package v2

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/runtime"
	"github.com/yunhanshu-net/runcher/v2/scheduler"
	"sync"
)

type Runcher struct {
	conn                *nats.Conn
	receiveRunnerSub    *nats.Subscription
	upstreamSub         *nats.Subscription
	runnerLock          map[string]*sync.RWMutex
	lk                  *sync.RWMutex
	runners             map[string]*runtime.Runners
	waitUUIDRunnerReady map[string]*waitReady

	Scheduler *scheduler.Scheduler

	//closeReq        chan string
}

type waitReady struct {
	ready  chan runnerReady
	runner *runtime.Runner
}

func NewRuncher() *Runcher {
	r := &Runcher{
		runnerLock:          make(map[string]*sync.RWMutex),
		lk:                  &sync.RWMutex{},
		runners:             make(map[string]*runtime.Runners),
		waitUUIDRunnerReady: make(map[string]*waitReady),
		Scheduler:           scheduler.NewDefaultScheduler(),
	}
	return r
}

func (r *Runcher) connectUpstream() error {
	upstreamSub, err := r.conn.Subscribe("upstream.>", func(msg *nats.Msg) {
		var req request.RunnerRequest
		//fmt.Printf("read msg,%s\n", msg.Subject)
		err := json.Unmarshal(msg.Data, &req)
		if err != nil {
			panic(err)
		}

		msgCtx := &request.Context{
			Msg:     msg,
			Request: &req,
		}
		err = r.handelMsg(msgCtx)
		if err != nil {
			panic(err)
		}

	})
	if err != nil {
		return err
	}
	r.upstreamSub = upstreamSub
	return nil
}

func (r *Runcher) handelMsg(reqCtx *request.Context) error {
	runnerResponse, err := r.request(reqCtx)
	if err != nil {
		panic(err)
	}
	msg := nats.NewMsg(reqCtx.Msg.Subject)
	marshal, err := json.Marshal(runnerResponse)
	if err != nil {
		panic(err)
	}
	msg.Data = marshal
	err = reqCtx.Msg.RespondMsg(msg)
	if err != nil {
		panic(err)
	}
	return nil
}
func (r *Runcher) Run() error {
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}
	r.conn = conn

	err = r.connectUpstream()
	if err != nil {
		return err
	}

	err = r.connectRunner()
	if err != nil {
		return err
	}

	return nil
}

func (r *Runcher) Close() {
}
