// Package scheduler runcher内核的系统调用，内核态，用户侧无感知
package scheduler

import (
	"context"
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/pkg/jsonx"
)

const (
	sysCallbackSysOnVersionChange = "sysOnVersionChange"
)

func (s *Scheduler) SysCallback(ctx context.Context, callbackType string, r *model.Runner, body interface{}) (interface{}, error) {

	runnerRequest := &request.Request{
		Route:  "/_sysCallback/" + callbackType,
		Method: "POST",
		Body:   jsonx.JSONString(body)}
	runnerIns, err := s.getRunner(r)
	if err != nil {
		return nil, err
	}
	rsp, err := runnerIns.Request(ctx, runnerRequest)
	if err != nil {
		return nil, err
	}
	fmt.Println(rsp)
	//if !rsp.OK() {
	//	return nil, fmt.Errorf(rsp.Msg)
	//}
	switch callbackType {
	case sysCallbackSysOnVersionChange:
		change, err := sysOnVersionChange(ctx, r)
		if err != nil {
			return nil, err
		}
		//*syscallback.SysOnVersionChangeResp
		return change, nil
	}

	return nil, fmt.Errorf("callbackType not found")
}

func sysOnVersionChange(ctx context.Context, runner *model.Runner) (*syscallback.SysOnVersionChangeResp, error) {
	var resp syscallback.SysOnVersionChangeResp
	resp.CurrentVersion = runner.Version
	versions, err := runner.GetLatestVersions(2)
	if err != nil {
		return nil, err
	}
	if len(versions) < 1 {
		return nil, fmt.Errorf("no versions found")
	}
	var oldVersion, newVersion string
	var oldVersionPath string
	newVersion = versions[0]
	newVersionPath := runner.GetApiPath() + "/" + newVersion + ".json"
	if len(versions) > 1 {
		oldVersion = versions[1]
		if oldVersion != "v0" {
			oldVersionPath = runner.GetApiPath() + "/" + oldVersion + ".json"
		}
	}
	add, del, updated, err := runner.DiffApi(oldVersionPath, newVersionPath)
	if err != nil {
		return nil, err
	}
	resp.AddApi = add
	resp.DelApi = del
	resp.UpdateApi = updated
	return &resp, nil
}
