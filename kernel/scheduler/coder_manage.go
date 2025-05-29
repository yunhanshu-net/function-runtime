package scheduler

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/pkg/dto/coder"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/runcher/runner"
)

func (s *Scheduler) addApisByNats(ctx context.Context, req *coder.AddApisReq) (*coder.AddApisResp, error) {
	var resp = new(coder.AddApisResp)
	newRunner, err := runner.NewRunner(*req.Runner)
	if err != nil {
		return nil, err
	}
	resp, err = newRunner.AddApis(ctx, req)
	if err != nil {
		err = errors.WithMessage(err, "addApisByNats err")
		return nil, err
	}
	return resp, nil
}
func (s *Scheduler) deleteProjectByNats(ctx context.Context, req *coder.DeleteProjectReq) (*coder.DeleteProjectResp, error) {
	var resp = new(coder.DeleteProjectResp)

	rn, err := runnerproject.NewRunner(req.User, req.Runner, conf.GetRunnerRoot())
	if err != nil {
		return nil, err
	}
	newRunner, err := runner.NewRunner(*rn)
	if err != nil {
		return nil, err
	}
	resp, err = newRunner.DeleteProject(ctx, req)
	if err != nil {
		err = errors.WithMessage(err, "deleteProjectByNats err")
		return nil, err
	}
	return resp, nil
}

func (s *Scheduler) AddApisByNats(ctx context.Context, msg *nats.Msg) {
	var req coder.AddApisReq
	var resp = new(coder.AddApisResp)
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
			logger.Errorf(ctx, "[AddApiByNats] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}
	newRunner, err := runnerproject.NewRunner(req.Runner.User, req.Runner.Name, conf.GetRunnerRoot())
	if err != nil {
		return
	}
	req.Runner = newRunner

	resp, err = s.addApisByNats(ctx, &req)
	if err != nil {
		return
	}
}

func (s *Scheduler) DeleteProject(ctx context.Context, msg *nats.Msg) {
	var req coder.DeleteProjectReq
	var resp = new(coder.DeleteProjectResp)
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
			logger.Errorf(ctx, "[AddApiByNats] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}

	resp, err = s.deleteProjectByNats(ctx, &req)
	if err != nil {
		return
	}
}

func (s *Scheduler) addBizPackage(ctx context.Context, r *coder.BizPackage) (*coder.BizPackageResp, error) {
	newRunner, err := runner.NewRunner(*r.Runner)
	if err != nil {
		return nil, err
	}
	rsp, err := newRunner.AddBizPackage(ctx, r)
	if err != nil {
		err = errors.WithMessage(err, "AddBizPackage err")
		return nil, err
	}
	return rsp, nil
}

func (s *Scheduler) AddBizPackage(ctx context.Context, msg *nats.Msg) {
	var req coder.BizPackage
	var resp = new(coder.BizPackageResp)
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
			logger.Errorf(ctx, "[AddBizPackage] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}
	newRunner, err := runnerproject.NewRunner(req.Runner.User, req.Runner.Name, conf.GetRunnerRoot())
	if err != nil {
		return
	}
	req.Runner = newRunner
	resp, err = s.addBizPackage(ctx, &req)
	if err != nil {
		return
	}
}

func (s *Scheduler) createProject(ctx context.Context, r *coder.CreateProjectReq) (*coder.CreateProjectResp, error) {
	r.Runner.Version = "v0"
	newRunner, err := runner.NewRunner(r.Runner)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "[createProject] req:%+v", r)
	rsp, err := newRunner.CreateProject(ctx)
	if err != nil {
		err = errors.WithMessage(err, "createProject err")
		return nil, err
	}
	return rsp, nil
}

func (s *Scheduler) CreateProject(ctx context.Context, msg *nats.Msg) {
	var req coder.CreateProjectReq
	var resp = new(coder.CreateProjectResp)
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
			logger.Errorf(ctx, "[CreateProject] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}
	resp, err = s.createProject(ctx, &req)
	if err != nil {
		return
	}
}
