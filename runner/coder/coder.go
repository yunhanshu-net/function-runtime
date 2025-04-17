package coder

import (
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
)

type Coder interface {
	AddBizPackage(codeBizPackage *coder.BizPackage) (*coder.BizPackageResp, error)
	AddApi(codeApi *coder.CodeApi) (*coder.AddApiResp, error)
	AddApis(codeApis []*coder.CodeApi) (resp *coder.AddApisResp, err error)
	CreateProject() (*coder.CreateProjectResp, error)
}

func NewCoder(runner *model.Runner) (Coder, error) {

	if runner.Version == "" {
		version, err := runner.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		runner.Version = version
	}
	switch runner.Language {
	case "go":
		return &Golang{runnerRoot: conf.GetRunnerRoot(), runner: runner}, nil
	default:
		runner.Language = "go"
		return &Golang{runnerRoot: conf.GetRunnerRoot(), runner: runner}, nil
	}
}
