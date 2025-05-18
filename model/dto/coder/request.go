package coder

import (
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
)

type AddApiReq struct {
	Runner  *runnerproject.Runner `json:"runner"`
	CodeApi *CodeApi              `json:"code_api"`
}

type AddApisReq struct {
	Runner   *runnerproject.Runner
	CodeApis []*CodeApi
}
