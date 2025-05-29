package coder

import (
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
)

type AddApiReq struct {
	Runner  *runnerproject.Runner `json:"runner"`
	CodeApi *CodeApi              `json:"code_api"`
}

type AddApisReq struct {
	Runner   *runnerproject.Runner `json:"runner"`
	CodeApis []*CodeApi            `json:"code_apis"`
	Msg      string                `json:"msg"`
}

type DeleteApisReq struct {
	Runner   *runnerproject.Runner `json:"runner"`
	CodeApis []*CodeApi            `json:"code_apis"`
	Msg      string                `json:"msg"`
}

type DeleteProjectReq struct {
	User    string `json:"user"`
	Runner  string `json:"runner"`
	Method  string `json:"method"`
	Router  string `json:"router"`
	Body    string `json:"body"`
	Version string `json:"version"`
}

type DeleteProjectResp struct {
}
