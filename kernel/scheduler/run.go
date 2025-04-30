package scheduler

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model"
	"runtime/debug"
	"strings"
)

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
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(string(debug.Stack()))
			}
		}()
		subjects := strings.Split(msg.Subject, ".")
		subject := subjects[1]
		if subject == "addApi" {
			s.AddApiByNats(msg)
		}

		if subject == "addApis" {
			s.AddApiByNats(msg)
		}

		if subject == "createProject" {
			s.CreateProject(msg)
		}

		if subject == "addBizPackage" {
			s.AddBizPackage(msg)
		}

	})
	if err != nil {
		return err
	}
	s.coderSub = coderSub

	return nil
}
