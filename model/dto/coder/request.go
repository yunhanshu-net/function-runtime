package coder

import (
	"github.com/yunhanshu-net/runcher/model"
)

type AddApiReq struct {
	Runner  *model.Runner `json:"runner"`
	CodeApi *CodeApi      `json:"code_api"`
}

type AddApisReq struct {
	Runner   *model.Runner
	CodeApis []*CodeApi
}
