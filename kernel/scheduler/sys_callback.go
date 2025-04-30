package scheduler

import (
	"context"
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/runcher/model/request"
)

func (s *Scheduler) SysCallback(callbackType string, r *model.Runner, body interface{}) (interface{}, error) {

	runnerRequest := &request.Request{
		Route:  "/_sysCallback/" + callbackType,
		Method: "POST",
		Body:   body}
	runnerIns, err := s.getRunner(r)
	if err != nil {
		return nil, err
	}
	rsp, err := runnerIns.Request(context.Background(), runnerRequest)
	if err != nil {
		return nil, err
	}
	if !rsp.OK() {
		return nil, fmt.Errorf(rsp.Msg)
	}
	switch callbackType {
	case "SysOnVersionChange":
		change, err := sysOnVersionChange(r)
		if err != nil {
			return nil, err
		}
		//*syscallback.SysOnVersionChangeResp
		return change, nil
	}

	return nil, fmt.Errorf("callbackType not found")
}

func sysOnVersionChange(runner *model.Runner) (*syscallback.SysOnVersionChangeResp, error) {
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
		oldVersionPath = runner.GetApiPath() + "/" + oldVersion + ".json"
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
