package v2

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model/request/manage"
)

func (r *Runcher) AddApi(msg *nats.Msg) error {
	var req manage.AddApi
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return err
	}
	err = r.Scheduler.AddApi(req.Runner, req.CodeApi)
	if err != nil {
		return err
	}

	////获取一下新增的接口参数详情
	//r.runRequest()

	newMsg := nats.NewMsg(msg.Subject)
	newMsg.Header.Set("code", "0")
	err = msg.RespondMsg(newMsg)
	if err != nil {
		panic(err)
	}
	return nil
}
