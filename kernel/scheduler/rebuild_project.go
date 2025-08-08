package scheduler

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/function-runtime/conf"
	"github.com/yunhanshu-net/function-runtime/pkg/dto/coder"
	"github.com/yunhanshu-net/function-runtime/runner"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/logger"
)

func (s *Scheduler) RebuildProjectByNats(ctx context.Context, msg *nats.Msg) {
	var req coder.RebuildProjectReq
	var resp = new(coder.RebuildProjectResp)
	var err error
	defer func() {
		rspMsg := nats.NewMsg(msg.Subject)
		if err != nil {
			rspMsg.Header.Set("code", "-1")
			rspMsg.Header.Set("msg", err.Error())
		} else {
			rspMsg.Header.Set("code", "0")
		}
		marshal, _ := json.Marshal(resp)
		rspMsg.Data = marshal
		err2 := msg.RespondMsg(rspMsg)
		if err2 != nil {
			logger.Errorf(ctx, "[RebuildProjectByNats] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}

	resp, err = s.rebuildProjectByNats(ctx, &req)
	if err != nil {
		return
	}
}

func (s *Scheduler) rebuildProjectByNats(ctx context.Context, req *coder.RebuildProjectReq) (*coder.RebuildProjectResp, error) {

	rn, err := runnerproject.NewRunner(req.User, req.Name, conf.GetRunnerRoot())
	if err != nil {
		return nil, err
	}
	newRunner, err := runner.NewRunner(*rn)
	if err != nil {
		return nil, err
	}
	rsp, err := newRunner.RebuildProject(ctx, req)
	if err != nil {
		err = errors.WithMessage(err, "DeleteApis err")
		return nil, err
	}
	return rsp, nil
}
