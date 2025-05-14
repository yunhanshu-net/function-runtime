package coder

import (
	"context"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
)

type Coder interface {
	AddBizPackage(ctx context.Context, codeBizPackage *coder.BizPackage) (*coder.BizPackageResp, error)
	AddApi(ctx context.Context, codeApi *coder.CodeApi) (*coder.AddApiResp, error)
	AddApis(ctx context.Context, codeApis []*coder.CodeApi) (resp *coder.AddApisResp, err error)
	CreateProject(ctx context.Context) (*coder.CreateProjectResp, error)
}

func NewCoder(runner *model.Runner) (Coder, error) {
	goCoder, err := NewGoCoder(runner)
	if err != nil {
		return nil, err
	}
	switch runner.Language {
	case "go":
		return goCoder, nil
	default:
		runner.Language = "go"
		return goCoder, nil
	}
}
