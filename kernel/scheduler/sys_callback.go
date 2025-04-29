package scheduler

import (
	"context"
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
)

func (s *Scheduler) SysCallback(callbackType string, r *model.Runner, body interface{}) (interface{}, error) {

	runnerRequest := &request.Request{
		Route:  "/_sysCallback/" + callbackType,
		Method: "POST",
		Body:   body}
	runnerIns := s.getRunner(r)
	rsp, err := runnerIns.Request(context.Background(), runnerRequest)
	if err != nil {
		return nil, err
	}
	switch callbackType {
	case "SysOnVersionChange":
		if !rsp.OK() {
			return nil, fmt.Errorf(rsp.Msg)
		}
		var sysOnVersionChangeResp syscallback.SysOnVersionChangeResp
		err := response.DecodeBody(rsp, &sysOnVersionChangeResp)
		if err != nil {
			return nil, err
		}
		//*syscallback.SysOnVersionChangeResp
		return &sysOnVersionChangeResp, nil
	}

	return nil, fmt.Errorf("callbackType not found")
}
