package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/runtime"
	"sync"
)

type waitReady struct {
	ready  chan runnerReady
	runner *runtime.Runner
}

func NewScheduler(conn *nats.Conn) *Scheduler {
	r := &Scheduler{
		runnerLock:      make(map[string]*sync.RWMutex),
		lk:              &sync.RWMutex{},
		runners:         make(map[string]*runtime.Runners),
		waitRunnerReady: make(map[string]*waitReady),
		conn:            conn,
	}

	return r
}

func (r *Scheduler) connectUpstream() error {
	upstreamSub, err := r.conn.Subscribe("upstream.>", func(msg *nats.Msg) {
		var req request.RunnerRequest
		fmt.Printf("read subject:%s msg:%s\n", msg.Subject, string(msg.Data))
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

func (r *Scheduler) handelMsg(reqCtx *request.Context) error {
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
	msg.Header.Set("code", "0")
	err = reqCtx.Msg.RespondMsg(msg)
	if err != nil {
		panic(err)
	}
	return nil
}
func (r *Scheduler) Run() error {

	err := r.connectUpstream()
	if err != nil {
		return err
	}

	err = r.connectRunner()
	if err != nil {
		return err
	}

	return nil
}

func (r *Scheduler) Close() {
}

func (r *Scheduler) connectRunner() error {

	subscribe, err := r.conn.Subscribe("runcher.>", func(msg *nats.Msg) {
		//接收来自runner的连接和关闭请求
		uid := msg.Header.Get("uuid")
		subject := msg.Header.Get("subject")
		if msg.Header.Get("connect") == "req" {
			rd := runnerReady{Err: nil, UUID: uid}
			ready, ok := r.waitRunnerReady[uid]
			if ok {
				ready.ready <- rd
				fmt.Printf("connect: uid%v subject:%s\n", uid, subject)
			} else {
			}
			newMsg := nats.NewMsg(msg.Subject)
			newMsg.Header.Set("status", "success")
			msg.RespondMsg(newMsg)
		}

		if msg.Header.Get("close") == "req" {
			//runner 关闭连接
			//r.waitRunnerReady["uuid"] <- runnerReady{Err: nil, UUID: uuid}
			fmt.Printf("close: uid:%v subject:%s\n", uid, subject)
			r.removeRunner(subject, uid)
			newMsg := nats.NewMsg(msg.Subject)
			newMsg.Header.Set("status", "success")
			msg.RespondMsg(newMsg)
		}

	})
	if err != nil {
		return err
	}
	r.receiveRunnerSub = subscribe
	return nil
}
