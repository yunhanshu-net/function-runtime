package scheduler

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/runcher/runner"
)

func (s *Scheduler) addApiByNats(req *coder.AddApiReq) (*coder.AddApiResp, error) {
	var resp = new(coder.AddApiResp)
	newRunner, err := runner.NewRunner(*req.Runner)

	if err != nil {
		return nil, err
	}
	resp, err = newRunner.AddApi(req.CodeApi)
	if err != nil {
		err = errors.WithMessage(err, "AddApi err")
		return nil, err
	}
	req.Runner.Version = req.Runner.GetNextVersion()
	callback, err := s.SysCallback("sysOnVersionChange", req.Runner, &syscallback.SysOnVersionChangeReq{})
	if err != nil {
		return nil, err
	}
	resp.SyscallChangeVersion = callback.(*syscallback.SysOnVersionChangeResp)
	return resp, nil
}

func (s *Scheduler) addApisByNats(req *coder.AddApisReq) (*coder.AddApisResp, error) {
	var resp = new(coder.AddApisResp)
	newRunner, err := runner.NewRunner(*req.Runner)
	if err != nil {
		return nil, err
	}
	resp, err = newRunner.AddApis(req.CodeApis)
	if err != nil {
		err = errors.WithMessage(err, "addApisByNats err")
		return nil, err
	}
	req.Runner.Version = req.Runner.GetNextVersion()
	callback, err := s.SysCallback("sysOnVersionChange", req.Runner, &syscallback.SysOnVersionChangeReq{})
	if err != nil {
		return nil, err
	}
	resp.SyscallChangeVersion = callback.(*syscallback.SysOnVersionChangeResp)
	return resp, nil
}

func (s *Scheduler) AddApiByNats(msg *nats.Msg) {
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
			logrus.Errorf("[AddApiByNats] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
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
	resp, err = s.addApiByNats(&req)
	if err != nil {
		return
	}

}

func (s *Scheduler) AddApisByNats(msg *nats.Msg) {
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
			logrus.Errorf("[AddApiByNats] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
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

	resp, err = s.addApisByNats(&req)
	if err != nil {
		return
	}
}

func (s *Scheduler) addBizPackage(r *coder.BizPackage) (*coder.BizPackageResp, error) {
	newRunner, err := runner.NewRunner(*r.Runner)
	if err != nil {
		return nil, err
	}
	rsp, err := newRunner.AddBizPackage(r)
	if err != nil {
		err = errors.WithMessage(err, "AddBizPackage err")
		return nil, err
	}
	return rsp, nil
}

func (s *Scheduler) AddBizPackage(msg *nats.Msg) {
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
			logrus.Errorf("[AddBizPackage] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
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
	resp, err = s.addBizPackage(&req)
	if err != nil {
		return
	}
}

func (s *Scheduler) createProject(r *coder.CreateProjectReq) (*coder.CreateProjectResp, error) {
	r.Runner.Version = "v0"
	newRunner, err := runner.NewRunner(r.Runner)
	if err != nil {
		return nil, err
	}
	logrus.Infof("[createProject] req:%+v", r)
	rsp, err := newRunner.CreateProject()
	if err != nil {
		err = errors.WithMessage(err, "createProject err")
		return nil, err
	}
	return rsp, nil
}

func (s *Scheduler) CreateProject(msg *nats.Msg) {
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
			logrus.Errorf("[CreateProject] msg.RespondMsg(rspMsg) err:%s err2:%s req:%+v", err.Error(), err2, req)
		}
	}()
	err = json.Unmarshal(msg.Data, &req)
	if err != nil {
		return
	}
	resp, err = s.createProject(&req)
	if err != nil {
		return
	}
}
