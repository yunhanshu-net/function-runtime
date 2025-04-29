package scheduler

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/runcher/runner"
)

func (s *Scheduler) addApiByNats(req *coder.AddApiReq) (*coder.AddApiResp, error) {
	var resp = new(coder.AddApiResp)
	newRunner := runner.NewRunner(*req.Runner)

	resp, err := newRunner.AddApi(req.CodeApi)
	if err != nil {
		err = errors.WithMessage(err, "AddApi err")
		return nil, err
	}
	callback, err := s.SysCallback("SysOnVersionChange", req.Runner, &syscallback.SysOnVersionChangeReq{})
	if err != nil {
		return nil, err
	}
	resp.SyscallChangeVersion = callback.(*syscallback.SysOnVersionChangeResp)
	return resp, nil
}

func (s *Scheduler) addApisByNats(req *coder.AddApisReq) (*coder.AddApisResp, error) {
	var resp = new(coder.AddApisResp)
	newRunner := runner.NewRunner(*req.Runner)

	resp, err := newRunner.AddApis(req.CodeApis)
	if err != nil {
		err = errors.WithMessage(err, "addApisByNats err")
		return nil, err
	}
	callback, err := s.SysCallback("SysOnVersionChange", req.Runner, &syscallback.SysOnVersionChangeReq{})
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
	resp, err = s.addApisByNats(&req)
	if err != nil {
		return
	}
}

//func (s *Coder) AddApi(api *coder.AddApiReq) error {
//	newRunner := runner.NewRunner(*api.Runner)
//	addApi, err := newRunner.AddApi(api.CodeApi)
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *Coder) AddApis(api *coder.AddApisReq) (errs []*coder.CodeApiCreateInfo, err error) {
//	newRunner := runner.NewRunner(*api.Runner)
//	errs, err = newRunner.AddApis(api.CodeApis)
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return nil, err
//	}
//	return errs, nil
//}
//
//func (s *Coder) AddBizPackage(api *coder.BizPackage) (err error) {
//	newRunner := runner.NewRunner(*api.Runner)
//	err = newRunner.AddBizPackage(api)
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (s *Coder) CreateProject(r *model.Runner) error {
//	newRunner := runner.NewRunner(*r)
//	err := newRunner.CreateProject()
//	//todo 这里要调用生命周期函数
//	if err != nil {
//		return err
//	}
//	return nil
//}
