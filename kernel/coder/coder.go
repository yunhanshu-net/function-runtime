package coder

import (
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
	manageSub, err := s.conn.Subscribe("coder.>", func(msg *nats.Msg) {
		subjects := strings.Split(msg.Subject, ".")
		subject := subjects[1]
		if subject == "add_api" {
			s.addApiByNats(msg)
		}

		if subject == "add_apis" {
			s.addApiByNats(msg)
		}

	})
	if err != nil {
		return err
	}
	s.manageSub = manageSub
	return nil
}
