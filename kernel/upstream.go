package kernel

// 上游的数据
//func (r *Runcher) upstreamFunc(msg *nats.Msg) {
//	var (
//		req request.Request
//		rsp = new(response.Response)
//		err error
//	)
//
//	rspMsp := nats.NewMsg(msg.Subject)
//	if msg.Data != nil {
//		err = json.Unmarshal(msg.Data, &req)
//		if err != nil {
//			rspMsp.Header.Set("error", "参数错误")
//			msg.RespondMsg(rspMsp)
//			return
//		}
//	}
//
//	subs := strings.Split(msg.Subject, ".")
//	user := subs[2]
//	soft := subs[3]
//	cmd := subs[4]
//	runnerMeta := &model.Runner{
//		Name:    soft,
//		User: user,
//		OssPath:    msg.Header.Get("fs_path"),
//		ToolType:   msg.Header.Get("type"),
//		Version:    msg.Header.Get("version"),
//	}
//	req.RunnerInfo.User = user
//	req.RunnerInfo.Soft = soft
//	req.RunnerInfo.Command = cmd
//	req.RunnerInfo.Command = strings.TrimPrefix(req.RunnerInfo.Command, "/")
//	req.Method = msg.Header.Get("method")
//
//	run := runner.NewRunner(runnerMeta)
//
//	req.RunnerInfo.RequestJsonPath = strings.ReplaceAll(req.GetRequestFilePath(run.GetInstallPath()), "\\", "/")
//	defer func() {
//		go os.Remove(req.RunnerInfo.RequestJsonPath)
//	}()
//	rsp, err = r.executor.Request(&req, runnerMeta)
//	if err != nil {
//		rspMsp.Header.Set("error", err.Error())
//		msg.RespondMsg(rspMsp)
//		return
//	}
//	data, err := json.Marshal(rsp)
//	if err != nil {
//		fmt.Println(err)
//	}
//	//cost := rsp.CallCostTime
//	//timex.Println(cost)
//	//执行引擎发起调用到程序执行结束的总耗时
//	rspMsp.Data = data
//	err = msg.RespondMsg(rspMsp)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
