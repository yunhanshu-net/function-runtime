package scheduler

import (
	"github.com/nats-io/nats.go"
)

var natsConn *nats.Conn
var subFunctionServer *nats.Subscription
var subRunner *nats.Subscription
var scheduler *Scheduler

func getConn() *nats.Conn {
	if natsConn == nil {
		conn, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}
		natsConn = conn
	}
	return natsConn
}

// PushFunctionServer 消息推送给FunctionServer
func PushFunctionServer(msg *nats.Msg) error {
	msg.Subject = "function-server.receiver"
	return getConn().PublishMsg(msg)
}

// PushRunner 消息推送给runner
//func PushRunner(runnerRequest *request.RunFunctionReq, runner *runnerproject.Runner) error {
//	sub := fmt.Sprintf("runner.%s.%s.%s.run", runner.User, runner.Name, runner.Version)
//	msg := nats.NewMsg(sub)
//	runnerRequest.Runner = runner
//	marshal, err := json.Marshal(runnerRequest)
//	if err != nil {
//		return err
//	}
//	msg.Data = marshal
//	msg.Header.Set(constants.TraceID, runnerRequest.TraceID)
//	return getConn().PublishMsg(msg)
//}
//
//// OnSubFunctionServer 接收到FunctionServer的消息后推送消息处理
//func OnSubFunctionServer(msg *nats.Msg) {
//
//	ctx := context.WithValue(context.Background(), constants.TraceID, msg.Header.Get(constants.TraceID))
//	//logger.Infof(ctx, "function.run >%s uid:%s", msg.Subject, string(msg.Data))
//	//接收runner关闭
//	logger.Infof(ctx, "Got message: %s", string(msg.Data))
//	runner, err := runnerproject.NewRunner(msg.Header.Get("user"), msg.Header.Get("runner"), conf.GetRunnerRoot(), msg.Header.Get("version"))
//	if err != nil {
//		panic(err)
//	}
//
//	req := request.RunFunctionReq{
//		Method:  msg.Header.Get("method"),
//		Router:  msg.Header.Get("router"),
//		TraceID: msg.Header.Get(constants.TraceID),
//		Runner:  runner,
//		//BodyType: "string",
//	}
//	if req.IsMethodGet() {
//		req.UrlQuery = msg.Header.Get("url_query")
//	} else {
//		req.Body = string(msg.Data)
//		req.BodyType = "string"
//	}
//	rsp := nats.NewMsg(msg.Subject)
//	rsp.Header.Set("code", "0")
//	//推送消息让runner处理
//	err = scheduler.RequestSync(ctx, &req)
//	if err != nil {
//		panic(err)
//	}
//
//}
//
//// SubFunctionServer function-runtime接收来自function-server的消息的主题
//func SubFunctionServer(handler nats.MsgHandler) error {
//	subscribe, err := getConn().Subscribe("function-server.sub", handler)
//	if err != nil {
//		return err
//	}
//	subFunctionServer = subscribe
//	return nil
//}
//
//// OnSubRunner 接收到runner的消息后转发给FunctionServer
//func OnSubRunner(msg *nats.Msg) {
//	err := PushFunctionServer(msg)
//	if err != nil {
//		panic(err)
//	}
//}
//
//// SubRunner function-runtime接收来自function-server的消息的主题
//func SubRunner(handler nats.MsgHandler) error {
//	subscribe, err := getConn().Subscribe("function-runner.sub", handler)
//	if err != nil {
//		return err
//	}
//	subRunner = subscribe
//	return nil
//}
//
//func Close() {
//	if subRunner != nil {
//		subRunner.Unsubscribe()
//	}
//	if subFunctionServer != nil {
//		subFunctionServer.Unsubscribe()
//	}
//	if natsConn != nil {
//		natsConn.Close()
//	}
//}

type MessageCenter struct {
	natsConn *nats.Conn
	natsSub  *nats.Subscription
}

// PublishFunctionServer function-server的接收主题
func (c *MessageCenter) PublishFunctionServer(msg *nats.Msg) error {
	//function-server.receiver
	err := c.natsConn.PublishMsg(msg)
	if err != nil {
		return err
	}
	return nil
}

// SubFunctionRuntimeReceiver function-runtime接收来自function-server的消息的主题
func (c *MessageCenter) SubFunctionRuntimeReceiver(handler nats.MsgHandler) error {
	subscribe, err := c.natsConn.Subscribe("function-runtime.receiver", handler)
	if err != nil {
		return err
	}
	c.natsSub = subscribe

	return nil
}

// PublishToRunner 给runner发消息
func (c *MessageCenter) PublishToRunner(msg *nats.Msg) error {

	err := c.natsConn.PublishMsg(msg)
	if err != nil {
		return err
	}
	return nil
}

// SubFunctionRuntimeReceiverFromFunctionGO 接收来自runner的消息
func (c *MessageCenter) SubFunctionRuntimeReceiverFromFunctionGO(handler nats.MsgHandler) error {
	subscribe, err := c.natsConn.Subscribe("function-run.receiver", handler)
	if err != nil {
		return err
	}
	c.natsSub = subscribe

	return nil
}
