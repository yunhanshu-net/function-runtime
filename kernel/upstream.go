package kernel

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"os"
	"strings"
)

// 上游的数据
func (r *Runcher) upstreamFunc(msg *nats.Msg) {
	var (
		req request.Request
		rsp = new(response.Response)
		err error
	)

	rspMsp := nats.NewMsg(msg.Subject)
	if msg.Data != nil {
		err = json.Unmarshal(msg.Data, &req)
		if err != nil {
			rspMsp.Header.Set("error", "参数错误")
			msg.RespondMsg(rspMsp)
			return
		}
	}

	subs := strings.Split(msg.Subject, ".")
	user := subs[2]
	soft := subs[3]
	cmd := subs[4]
	runnerMeta := &model.Runner{
		AppCode:    soft,
		TenantUser: user,
		OssPath:    msg.Header.Get("fs_path"),
		ToolType:   msg.Header.Get("type"),
		Version:    msg.Header.Get("version"),
	}
	req.SoftInfo.User = user
	req.SoftInfo.Soft = soft
	req.SoftInfo.Command = cmd
	req.SoftInfo.Command = strings.TrimPrefix(req.SoftInfo.Command, "/")
	req.Method = msg.Header.Get("method")

	run := runner.NewRunner(runnerMeta)

	req.SoftInfo.RequestJsonPath = strings.ReplaceAll(req.GetRequestFilePath(run.GetInstallPath()), "\\", "/")
	defer func() {
		go os.Remove(req.SoftInfo.RequestJsonPath)
	}()
	rsp, err = r.executor.Request(&req, runnerMeta)
	if err != nil {
		rspMsp.Header.Set("error", err.Error())
		msg.RespondMsg(rspMsp)
		return
	}
	data, err := json.Marshal(rsp)
	if err != nil {
		fmt.Println(err)
	}
	//cost := rsp.CallCostTime
	//timex.Println(cost)
	//执行引擎发起调用到程序执行结束的总耗时
	rspMsp.Data = data
	err = msg.RespondMsg(rspMsp)
	if err != nil {
		fmt.Println(err)
	}
}
