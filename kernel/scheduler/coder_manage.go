package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/api"
	"github.com/yunhanshu-net/runcher/model/dto/callback"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/runcher/runner"
	"go/types"
)

func (s *Scheduler) addApiByNats(ctx context.Context, req *coder.AddApiReq) (*coder.AddApiResp, error) {
	var resp = new(coder.AddApiResp)
	newRunner, err := runner.NewRunner(*req.Runner)

	if err != nil {
		return nil, err
	}
	resp, err = newRunner.AddApi(ctx, req.CodeApi)
	if err != nil {
		err = errors.WithMessage(err, "AddApi err")
		return nil, err
	}

	req.Runner.Version = req.Runner.GetNextVersion()

	sysCallbackResp, err := s.SysCallback(ctx, "sysOnVersionChange", req.Runner, &syscallback.SysOnVersionChangeReq{})
	if err != nil {
		return nil, err
	}

	apiDiff, ok := sysCallbackResp.(*syscallback.SysOnVersionChangeResp)
	if !ok {
		return nil, fmt.Errorf("addApiByNats sysCallbackResp.(*syscallback.SysOnVersionChangeResp)")
	}

	call := func(callbackType string, apis []*api.Info) error {
		if apis == nil {
			return nil
		}
		runnerIns := req.Runner
		if callbackType == CallbackTypeBeforeApiDelete {
			//这里应该调用旧版本的runner来执行回调函数，因为新版本删除了该api，那么该api的回调也不复存在，需要回到旧版本进行执行
			oldVersion, err := runnerIns.GetOldVersion()
			if err != nil {
				return err
			}
			runnerIns = oldVersion
		}
		for _, info := range apis {
			if !info.ExistCallback(callbackType) { //不存在回调，跳过
				continue
			}
			bd := &callback.OnApiCreatedReq{Method: info.Method, Router: info.Router}

			r := &callback.Request{
				Method: info.Method,
				Router: info.Router,
				Type:   callbackType,
				Body:   bd,
			}

			rsp := &callback.ResponseWith[*callback.OnApiCreatedReq, types.Nil]{}
			err = s.UserCallback(runnerIns, r, rsp)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = call(CallbackTypeOnApiCreated, apiDiff.AddApi)
	if err != nil {
		return nil, err
	}

	err = call(CallbackTypeOnApiUpdated, apiDiff.UpdateApi)
	if err != nil {
		return nil, err
	}

	err = call(CallbackTypeBeforeApiDelete, apiDiff.DelApi)
	if err != nil {
		return nil, err
	}

	resp.SyscallChangeVersion = sysCallbackResp.(*syscallback.SysOnVersionChangeResp)
	return resp, nil
}

func (s *Scheduler) addApisByNats(ctx context.Context, req *coder.AddApisReq) (*coder.AddApisResp, error) {
	var resp = new(coder.AddApisResp)
	newRunner, err := runner.NewRunner(*req.Runner)
	if err != nil {
		return nil, err
	}
	resp, err = newRunner.AddApis(ctx, req.CodeApis)
	if err != nil {
		err = errors.WithMessage(err, "addApisByNats err")
		return nil, err
	}
	req.Runner.Version = req.Runner.GetNextVersion()
	callback, err := s.SysCallback(ctx, "sysOnVersionChange", req.Runner, &syscallback.SysOnVersionChangeReq{})
	if err != nil {
		return nil, err
	}
	resp.SyscallChangeVersion = callback.(*syscallback.SysOnVersionChangeResp)
	return resp, nil
}

func (s *Scheduler) AddApiByNats(ctx context.Context, msg *nats.Msg) {
	var req coder.AddApiReq
	var resp = new(coder.AddApiResp)
	//var callbackResp = new(syscallback.Response[*syscallback.SysOnVersionChangeResp])
	var err error
	defer func() {
		rspMsg := nats.NewMsg(msg.Subject)
		if err != nil {
			rspMsg.Header.Set("code", "-1")
			rspMsg.Header.Set("msg", err.Error())
		} else {
			rspMsg.Header.Set("code", "0")
		}
		//resp.SyscallChangeVersion = callbackResp.Data
		marshal, _ := json.Marshal(resp)
		rspMsg.Data = marshal
		err2 := msg.RespondMsg(rspMsg)
		if err2 != nil {
			logger.ErrorContextf(ctx, "[AddApiByNats] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}
	newRunner, err := model.NewRunner(req.Runner.User, req.Runner.Name)
	if err != nil {
		return
	}
	req.Runner = newRunner
	resp, err = s.addApiByNats(ctx, &req)
	if err != nil {
		return
	}

}

func (s *Scheduler) AddApisByNats(ctx context.Context, msg *nats.Msg) {
	var req coder.AddApisReq
	var resp = new(coder.AddApisResp)
	var callbackResp = new(syscallback.ResponseWith[*syscallback.SysOnVersionChangeResp])

	var err error
	//var errs []*coder.CodeApiCreateInfo
	defer func() {
		rspMsg := nats.NewMsg(msg.Subject)
		if err != nil {
			rspMsg.Header.Set("code", "-1")
			rspMsg.Header.Set("msg", err.Error())
		} else {
			rspMsg.Header.Set("code", "0")
		}
		resp.SyscallChangeVersion = callbackResp.Data
		marshal, _ := json.Marshal(resp)
		rspMsg.Data = marshal
		err2 := msg.RespondMsg(rspMsg)
		if err2 != nil {
			logger.Errorf("[AddApiByNats] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}
	newRunner, err := model.NewRunner(req.Runner.User, req.Runner.Name)
	if err != nil {
		return
	}
	req.Runner = newRunner

	resp, err = s.addApisByNats(ctx, &req)
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
	//var callbackResp = new(syscallback.ResponseWith[*syscallback.SysOnVersionChangeResp])

	var err error
	//var errs []*coder.CodeApiCreateInfo
	defer func() {
		rspMsg := nats.NewMsg(msg.Subject)
		if err != nil {
			rspMsg.Header.Set("code", "-1")
			rspMsg.Header.Set("msg", err.Error())
		} else {
			rspMsg.Header.Set("code", "0")
		}
		//resp.SyscallChangeVersion = callbackResp.Data
		marshal, _ := json.Marshal(resp)
		rspMsg.Data = marshal
		err2 := msg.RespondMsg(rspMsg)
		if err2 != nil {
			logger.Errorf("[AddBizPackage] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}
	newRunner, err := model.NewRunner(req.Runner.User, req.Runner.Name)
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
	logger.Infof("[createProject] req:%+v", r)
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
	//var errs []*coder.CodeApiCreateInfo
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
			logger.Errorf("[CreateProject] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
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
