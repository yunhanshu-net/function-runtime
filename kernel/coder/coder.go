package coder

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"os"
	"strings"
)

type Coder struct {
	RunnerRoot string
	conn       *nats.Conn
	manageSub  *nats.Subscription
}

func NewDefaultCoder(conn *nats.Conn) *Coder {
	runnerRoot := "./soft"
	if os.Getenv("RUNNER_ROOT") != "" {
		runnerRoot = strings.TrimSuffix(os.Getenv("RUNNER_ROOT"), "/") + "/soft"
	}

	return &Coder{
		conn:       conn,
		RunnerRoot: runnerRoot,
	}
}

func (s *Coder) Run() error {
	err := s.connectManage()
	if err != nil {
		return err
	}
	//监听nats消息
	return nil
}

func (s *Coder) connectManage() error {
	manageSub, err := s.conn.Subscribe("manage.>", func(msg *nats.Msg) {

		subjects := strings.Split(msg.Subject, ".")
		subject := subjects[1]
		var err error
		if subject == "add_api" {
			fmt.Println("add_api:", string(msg.Data))
			err = s.addApi(msg)
			if err != nil {
				fmt.Println(err)
			}
		}

		newMsg := nats.NewMsg(msg.Subject)
		if err != nil {
			newMsg.Header.Set("code", "-1")
			newMsg.Header.Set("msg", err.Error())
		} else {
			newMsg.Header.Set("code", "0")
		}
		err = msg.RespondMsg(newMsg)
		if err != nil {
			panic(err)
		}

	})
	if err != nil {
		return err
	}
	s.manageSub = manageSub
	return nil
}
