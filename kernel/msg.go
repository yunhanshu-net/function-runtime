package kernel

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"strings"
	"time"
)

func (b *Executor) RunnerListen() error {
	connectKey := fmt.Sprintf("runner.connect.*.*")
	closeKey := fmt.Sprintf("runner.close.*.*")
	connectSub, err := b.nats.Subscribe(connectKey, func(msg *nats.Msg) {
		fmt.Printf("connectKey:%s 建立长连接\n", connectKey)
		b.RuncherConnectLock.Lock()
		defer b.RuncherConnectLock.Unlock()
		subject := msg.Subject
		split := strings.Split(subject, ".")
		if len(split) < 4 {
			fmt.Println("RunnerListen: key err", split)
			return
		}
		user := split[2]
		runner := split[3]
		if msg.Header.Get("status") == "0" {
			h := nats.Header{}
			h.Set("status", "0")
			err := msg.RespondMsg(&nats.Msg{Header: h})
			if err != nil {
				fmt.Println("RunnerListen: RespondMsg key err", err.Error())
				return
			}
			if _, ok := b.RuncherConnect[fmt.Sprintf("%s.%s", user, runner)]; !ok {
				b.RuncherConnect[fmt.Sprintf("%s.%s", user, runner)] = make(chan ConnectStatus, 1)
			}
			b.RuncherConnect[fmt.Sprintf("%s.%s", user, runner)] <- ConnectStatus{Success: true}

		}
	})
	if err != nil {
		return err
	}
	closeSub, err := b.nats.Subscribe(closeKey, func(msg *nats.Msg) {
		fmt.Printf("closeKey:%s 释放长连接\n", closeKey)

		b.RuncherCloseLock.Lock()
		defer b.RuncherCloseLock.Unlock()
		subject := msg.Subject
		split := strings.Split(subject, ".")
		if len(split) < 4 {
			fmt.Println("RunnerListen: key err", split)
			return
		}
		user := split[2]
		runner := split[3]
		if msg.Header.Get("status") == "0" {
			if _, ok := b.RuncherClose[fmt.Sprintf("%s.%s", user, runner)]; !ok {
				b.RuncherClose[fmt.Sprintf("%s.%s", user, runner)] = make(chan CloseStatus, 1)
			}
			b.RuncherClose[fmt.Sprintf("%s.%s", user, runner)] <- CloseStatus{Success: true}
		}
	})
	if err != nil {
		return err
	}
	b.connectSub = connectSub
	b.closeSub = closeSub
	return nil
}

// WaitRunnerConnected 等待指定runner成功建立连接
func (b *Executor) WaitRunnerConnected(ctx context.Context, user, app string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second * 5):
		return fmt.Errorf("wait runner connected timeout")
	case status := <-b.RuncherConnect[fmt.Sprintf("%s.%s", user, app)]:
		if !status.Success {
			return fmt.Errorf("wait runner connected err:" + status.Message)
		}
		return nil
	}
}
