package coder

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request/manage"
	"github.com/yunhanshu-net/runcher/pkg/codex"
	"github.com/yunhanshu-net/runcher/runner/coder"
)

func (s *Coder) addApi(msg *nats.Msg) error {
	var req manage.AddApi
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return err
	}
	err = s.AddApi(req.Runner, req.CodeApi)
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

func (s *Coder) AddApi(runner *model.Runner, codeApi *codex.CodeApi) error {
	newCoder, err := coder.NewCoder(runner.Language)
	if err != nil {
		return err
	}
	err = newCoder.AddApi(s.RunnerRoot, runner, codeApi)
	if err != nil {
		return err
	}
	fmt.Println("add api success")
	return nil
}
