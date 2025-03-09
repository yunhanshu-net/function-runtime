package v2

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/runtime"
	"sync"
)

type Runcher struct {
	conn        *nats.Conn
	sub         *nats.Subscription
	upstreamSub *nats.Subscription

	runnerLock map[string]*sync.RWMutex
	lk         *sync.RWMutex
	runners    map[string]*runtime.Runners

	waitUUIDRunnerReady map[string]*waitReady
	//closeReq        chan string
}

type waitReady struct {
	ready  chan runnerReady
	runner *runtime.Runner
}

func NewRuncher() *Runcher {
	//opt := server.Options{}
	//s, err := server.NewServer()
	//if err != nil {
	//	panic(err)
	//}

	r := &Runcher{
		runnerLock:          make(map[string]*sync.RWMutex),
		lk:                  &sync.RWMutex{},
		runners:             make(map[string]*runtime.Runners),
		waitUUIDRunnerReady: make(map[string]*waitReady),
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
