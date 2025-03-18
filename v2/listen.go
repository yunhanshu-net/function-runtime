package v2

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

func (r *Runcher) connectRunner() error {

	subscribe, err := r.conn.Subscribe("runcher.>", func(msg *nats.Msg) {
		//接收来自runner的连接和关闭请求
		uid := msg.Header.Get("uuid")
		subject := msg.Header.Get("subject")
		if msg.Header.Get("connect") == "req" {
			rd := runnerReady{Err: nil, UUID: uid}
			ready, ok := r.waitUUIDRunnerReady[uid]
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
