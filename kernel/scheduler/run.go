package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/pkg/constants"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/sdk-go/pkg/dto/request"
	"runtime/debug"
	"strings"
	"time"
)

func (s *Scheduler) Run() error {

	functionSub, err := s.natsConn.Subscribe("function.run.>", func(msg *nats.Msg) {
		ctx := context.WithValue(context.Background(), constants.TraceID, msg.Header.Get(constants.TraceID))
		//logger.Infof(ctx, "function.run >%s uid:%s", msg.Subject, string(msg.Data))
		//接收runner关闭
		runner, err := runnerproject.NewRunner(msg.Header.Get("user"), msg.Header.Get("runner"), conf.GetRunnerRoot(), msg.Header.Get("version"))
		if err != nil {
			panic(err)
		}

		req := request.RunFunctionReq{
			Method:  msg.Header.Get("method"),
			Router:  msg.Header.Get("router"),
			TraceID: msg.Header.Get(constants.TraceID),
			Runner:  runner,
			//BodyType: "string",
		}
		if req.IsMethodGet() {
			req.UrlQuery = msg.Header.Get("url_query")
		} else {
			req.Body = string(msg.Data)
			req.BodyType = "string"
		}
		rsp := nats.NewMsg(msg.Subject)
		rsp.Header.Set("code", "0")
		response, err := s.Request(ctx, &req)
		if err != nil {
			rsp.Header.Set("code", "-1")
			rsp.Header.Set("msg", err.Error())
			err = msg.RespondMsg(rsp)
			if err != nil {
				logger.Error(ctx, "request error", err)
			}
			return
		}
		for k, v := range response.MetaData {
			rsp.Header.Set(k, fmt.Sprintf("%v", v))
		}
		marshal, err := json.Marshal(response)
		if err != nil {
			logger.Error(ctx, "response marshal error", err)
			panic(err)
		}
		rsp.Data = marshal

		err = msg.RespondMsg(rsp)
		if err != nil {
			logger.Error(ctx, "request error", err)
			return
		}
	})
	if err != nil {
		return err
	}
	s.functionSub = functionSub

	//监听runner的启动和关闭事件
	subscribe, err := s.natsConn.Subscribe("close.runner", func(msg *nats.Msg) {
		ctx := context.WithValue(context.Background(), constants.TraceID, msg.Header.Get(constants.TraceID))
		logger.Infof(ctx, "runner.close >%s uid:%s", msg.Subject, string(msg.Data))
		//接收runner关闭
		rn, err := runnerproject.NewRunner(msg.Header.Get("user"), msg.Header.Get("name"), conf.GetRunnerRoot(), msg.Header.Get("version"))
		if err != nil {
			//todo 错误处理
			panic(err)
		}
		err = s.stopRunner(rn)
		if err != nil {
			logger.Errorf(ctx, "runner:%s close err:%s", rn.GetRequestSubject(), err.Error())
			return
		}
		rsp := nats.NewMsg(msg.Subject)
		rsp.Header.Set("code", "0")
		err = msg.RespondMsg(rsp)
		if err != nil {
			logger.Errorf(ctx, "runner:%s close err:%s", rn.GetRequestSubject(), err.Error())
			return
		}
		logger.Infof(ctx, "runner:%s close success", rn.GetRequestSubject())
	})
	if err != nil {
		return err
	}
	s.closeSub = subscribe

	coderSub, err := s.natsConn.Subscribe("coder.>", func(msg *nats.Msg) {
		ctx := context.WithValue(context.Background(), constants.TraceID, msg.Header.Get(constants.TraceID))

		defer func() {
			if err := recover(); err != nil {
				fmt.Println(string(debug.Stack()))
			}
		}()
		subjects := strings.Split(msg.Subject, ".")
		subject := subjects[1]

		if subject == "addApis" {
			s.AddApisByNats(ctx, msg)
		}

		if subject == "createProject" {
			s.CreateProject(ctx, msg)
		}

		if subject == "addBizPackage" {
			s.AddBizPackage(ctx, msg)
		}
		if subject == "deleteProject" {
			s.DeleteProject(ctx, msg)
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
				logger.Debug(context.Background(), "关闭nats监控")
				return
			case <-time.After(5 * time.Second):
				// 检查连接状态
				if s.natsConn.Status() != nats.CONNECTED {
					logger.Infof(context.Background(), "NATS连接已断开")
				}
			}
		}
	}()

	return nil
}
