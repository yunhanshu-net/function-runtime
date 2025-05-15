package scheduler

import (
	"context"
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/pkg/jsonx"
	"github.com/yunhanshu-net/runcher/pkg/logger"
)

const (
	// 页面事件
	CallbackTypeOnPageLoad = "OnPageLoad" // 页面加载时

	// API 生命周期
	CallbackTypeOnApiCreated    = "OnApiCreated"    // API创建完成时
	CallbackTypeOnApiUpdated    = "OnApiUpdated"    // API更新时
	CallbackTypeBeforeApiDelete = "BeforeApiDelete" // API删除前
	CallbackTypeAfterApiDeleted = "AfterApiDeleted" // API删除后

	// 运行器(Runner)生命周期
	CallbackTypeBeforeRunnerClose = "BeforeRunnerClose" // 运行器关闭前
	CallbackTypeAfterRunnerClose  = "AfterRunnerClose"  // 运行器关闭后

	// 版本控制
	CallbackTypeOnVersionChange = "OnVersionChange" // 版本变更时

	// 输入交互
	CallbackTypeOnInputFuzzy    = "OnInputFuzzy"    // 输入模糊匹配
	CallbackTypeOnInputValidate = "OnInputValidate" // 输入校验

	// 表格操作
	CallbackTypeOnTableDeleteRows = "OnTableDeleteRows" // 删除表格行
	CallbackTypeOnTableUpdateRow  = "OnTableUpdateRow"  // 更新表格行
	CallbackTypeOnTableSearch     = "OnTableSearch"     // 表格搜索
)

func (s *Scheduler) UserCallback(r *model.Runner, callReq interface{}, resp interface{}) error {

	logger.Debugf(context.Background(), "[UserCallback] /_callback body:%s", jsonx.JSONString(callReq))
	runnerRequest := &request.Request{
		Route:  "/_callback",
		Method: "POST",
		Body:   jsonx.JSONString(callReq)}
	runnerIns, err := s.getRunner(r)
	if err != nil {
		return err
	}
	rsp, err := runnerIns.Request(context.Background(), runnerRequest)
	if err != nil {
		return err
	}
	if !rsp.OK() {
		return fmt.Errorf(rsp.Msg)
	}
	body, err := rsp.DecodeBody()
	if err != nil {
		return err
	}
	if body.Err() != nil {
		return body.Err()
	}
	if resp == nil {
		return nil
	}
	err = body.DecodeData(resp)
	if err != nil {
		logger.Errorf(context.Background(), "[UserCallback] /_callback err:%s sys_resp:%+v user_resp:%+v", err.Error(), rsp, resp)
		return err
	}
	logger.Debugf(context.Background(), "[UserCallback] /_callback sys_resp:%+v user_resp:%+v", rsp, resp)
	return nil

	//if !rsp.OK() {
	//	return nil, fmt.Errorf(rsp.Msg)
	//}

}
