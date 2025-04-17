package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
)

//AddBizPackage(codeBizPackage *coder.BizPackage) error

func AddBizPackage(c *gin.Context) {
	var (
		r   coder.BizPackage
		rsp *coder.BizPackageResp
		err error
	)
	defer func() {
		logrus.Infof("[AddBizPackage] req:%+v rsp:%+v err:%v", r, rsp, err)
	}()
	err = c.ShouldBindJSON(&r)
	if err != nil {
		response.FailWithMessage(c, "参数错误")
		return
	}
	newRunner := runner.NewRunner(*r.Runner)
	rsp, err = newRunner.AddBizPackage(&r)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithData(c, rsp)
}

func CreateProject(c *gin.Context) {
	var (
		r   coder.CreateProjectReq
		rsp *coder.CreateProjectResp
		err error
	)
	defer func() {
		logrus.Infof("[AddBizPackage] req:%+v rsp:%+v err:%v", r, rsp, err)
	}()
	err = c.ShouldBindJSON(&r)
	if err != nil {
		response.FailWithMessage(c, "参数错误")
		return
	}

	newRunner := runner.NewRunner(r.Runner)
	project, err := newRunner.CreateProject()
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	response.OkWithData(c, project)
}
