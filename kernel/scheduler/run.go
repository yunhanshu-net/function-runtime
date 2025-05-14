package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/pkg/constants"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"go.uber.org/zap"
	"runtime/debug"
	"strings"
	"time"
)

func (s *Scheduler) Run() error {

	functionSub, err := s.natsConn.Subscribe("function.run.>", func(msg *nats.Msg) {

		logger.Infof("function.run >%s uid:%s", msg.Subject, string(msg.Data))
		//接收runner关闭
		req := request.RunnerRequest{
			Request: &request.Request{
				Method: msg.Header.Get("method"),
				Route:  msg.Header.Get("router"),
				Body:   string(msg.Data),
			},
			Runner: &model.Runner{
				Name:     msg.Header.Get("runner"),
				User:     msg.Header.Get("user"),
				Version:  msg.Header.Get("version"),
				Language: "go",
			},
		}
		ctx := context.WithValue(context.Background(), constants.TraceID, msg.Header.Get(constants.TraceID))
		//var req request.RunnerRequest
		rsp := nats.NewMsg(msg.Subject)
		rsp.Header.Set("code", "0")
		response, err := s.Request(ctx, &req)
		if err != nil {
			rsp.Header.Set("code", "-1")
			rsp.Header.Set("msg", err.Error())
			err = msg.RespondMsg(rsp)
			if err != nil {
				logger.Error("request error", zap.Error(err))
				return
			}
		}
		for k, v := range response.MetaData {
			rsp.Header.Set(k, fmt.Sprintf("%v", v))
		}
		switch response.Body.(type) {
		case string:
			rsp.Data = []byte(response.Body.(string))
		default:
			rsp.Data, err = json.Marshal(response.Body)
			if err != nil {
				logger.Error("request error", zap.Error(err))
			}
		}

		err = msg.RespondMsg(rsp)
		if err != nil {
			logger.Error("request error", zap.Error(err))
			return
		}
		logger.Info("request success", zap.String("uid", msg.Subject))
	})
	if err != nil {
		return err
	}
	s.functionSub = functionSub

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
		fmt.Printf("runner:%s close success\n", m.GetRequestSubject())

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
