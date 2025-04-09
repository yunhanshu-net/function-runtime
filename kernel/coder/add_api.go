package coder

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model/request/manage"
)

func (s *Coder) addApi(msg *nats.Msg) error {
	var req manage.AddApi
	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		return err
	}
	//newRunner := runner.NewRunner(req.Runner)
	//err = newRunner.AddApi(req.CodeApi)
	if err != nil {
		return err
	}
	return nil
}
