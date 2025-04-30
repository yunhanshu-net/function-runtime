package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"github.com/yunhanshu-net/runcher/service/http2nats"
)

func AddApi(c *gin.Context) {
	var (
		r   coder.AddApiReq
		rsp *coder.AddApiResp
		err error
	)
	defer func() {
		logrus.Infof("[AddApi] req:%+v rsp:%+v err:%v", r, rsp, err)
	}()
	err = c.ShouldBindJSON(&r)
	if err != nil {
		response.FailWithMessage(c, "参数错误")
		return
	}
	//err = cmd.Runcher.Coder.AddApi(&r)

	newRunner, err := runner.NewRunner(*r.Runner)
	if err != nil {
		panic(err)
	}
	rsp, err = newRunner.AddApi(r.CodeApi)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, rsp)
}

func AddApis(c *gin.Context) {
	var (
		r   coder.AddApisReq
		rsp *coder.AddApisResp
		err error
	)
	defer func() {
		logrus.Infof("[AddApis] req:%+v rsp:%+v err:%v ", r, rsp, err)
	}()
	err = c.ShouldBindJSON(&r)
	if err != nil {
		response.FailWithMessage(c, "参数错误")
		return
	}

	newRunner, err := runner.NewRunner(*r.Runner)
	if err != nil {
		panic(err)
	}
	rsp, err = newRunner.AddApis(r.CodeApis)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithData(c, rsp)
}

func Manage(c *gin.Context) {
	data, err := http2nats.GinRequest(c)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	c.Data(200, "application/json", data)
}
