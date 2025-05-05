package scheduler

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/constants"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"go.uber.org/zap"
	"runtime/debug"
	"strings"
	"time"
)

func (s *Scheduler) Run() error {

	//group := uuid.New().String()
	//监听runner的启动和关闭事件
	subscribe, err := s.natsConn.Subscribe("close.runner", func(msg *nats.Msg) {
		logger.Infof("runner.close >%s uid:%s", msg.Subject, string(msg.Data))
		//接收runner关闭
		var m model.Runner
		m.Version = msg.Header.Get("version")
		m.User = msg.Header.Get("user")
		m.Name = msg.Header.Get("name")
		err := s.stopRunner(&m)
		if err != nil {
			logger.Errorf("runner:%s close err:%s", m.GetRequestSubject(), err.Error())
			return
		}
		rsp := nats.NewMsg(msg.Subject)
		rsp.Header.Set("code", "0")
		err = msg.RespondMsg(rsp)
		if err != nil {
			logger.Errorf("runner:%s close err:%s", m.GetRequestSubject(), err.Error())
			return
		}
		logger.Infof("runner:%s close success", m.GetRequestSubject())

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
		ctx := context.WithValue(context.Background(), constants.TraceID, msg.Header.Get(constants.TraceID))
		if subject == "addApi" {
			s.AddApiByNats(ctx, msg)
		}

		if subject == "addApis" {
			s.AddApiByNats(ctx, msg)
		}

		if subject == "createProject" {
			s.CreateProject(ctx, msg)
		}

		if subject == "addBizPackage" {
			s.AddBizPackage(ctx, msg)
		}

	})
	if err != nil {
		return err
	}
	s.coderSub = coderSub

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				logger.Debug("关闭nats监控")
				return
			case <-time.After(5 * time.Second):
				// 检查连接状态
				if s.natsConn.Status() != nats.CONNECTED {
					logger.Error("NATS连接已断开", zap.String("status", s.natsConn.Status().String()))
				}
			}
		}
	}()

	return nil
}
