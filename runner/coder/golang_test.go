package coder

import (
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
	"testing"
)

var code = `package array

import "github.com/yunhanshu-net/sdk-go/runner"

// 1. 定义请求/响应结构体
type UniqueReq struct {
	Arr []string 
}
type UniqueResp struct {
	UniqueArr []string 
}

// 2. 注册路由
func init() {
	runner.Post("/array/unique", UniqueRunner)
}

// 3. 核心逻辑
func Unique(req *UniqueReq) (resp UniqueResp, err error) {
	seen := make(map[string]bool)
	for _, v := range req.Arr {
		if !seen[v] {
			resp.UniqueArr = append(resp.UniqueArr, v)
			seen[v] = true
		}
	}
	return resp, nil
}

// 4. HTTP处理函数
func UniqueRunner(ctx *runner.HttpContext) {
	var req UniqueReq
	if err := ctx.Request.ShouldBindJSON(&req); err != nil {
		ctx.Response.FailWithJSON(400, err.Error())
		return
	}
	resp, err := Unique(&req)
	if err != nil {
		ctx.Response.FailWithJSON(500, err.Error())
		return
	}
	ctx.Response.JSON(resp).Build()
}`

func TestCreateProject(t *testing.T) {
	g := &Golang{
		runnerRoot: "/Users/beiluo/Documents/code/sdk-go/soft",
	}
	r := &model.Runner{
		User:     "kuaishou",
		Name:     "api",
		Language: "go",
		Version:  "v1",
	}
	err := g.CreateProject(r)
	if err != nil {
		panic(err)
	}
}

func TestAddPackage(t *testing.T) {
	g := &Golang{
		runnerRoot: "/Users/beiluo/Documents/code/sdk-go/soft",
	}
	r := &model.Runner{
		User:     "kuaishou",
		Name:     "api",
		Language: "go",
		Version:  "v1",
	}
	packages := []*codex.BizPackage{

		{
			AbsPackagePath: "array/diff",
			Language:       "go",
			EnName:         "diff",
		},

		{
			AbsPackagePath: "array/excel",
			Language:       "go",
			EnName:         "excel",
		},

		{
			AbsPackagePath: "pdf",
			Language:       "go",
			EnName:         "pdf",
		},
		{
			AbsPackagePath: "pdf/tojpg",
			Language:       "go",
			EnName:         "tojpg",
		},
	}

	for _, pkg := range packages {
		err := g.AddBizPackage(r, pkg)
		if err != nil {
			panic(err)
		}
	}
}

func TestAddApi(t *testing.T) {
	g := &Golang{
		runnerRoot: "/Users/beiluo/Documents/code/sdk-go/soft",
	}
	r := &model.Runner{
		User:     "kuaishou",
		Name:     "api",
		Language: "go",
		Version:  "v1",
	}
	api := &codex.CodeApi{
		Language:       "go",
		EnName:         "unique",
		Package:        "array",
		AbsPackagePath: "array",
		Code:           code,
	}

	err := g.AddApi(g.runnerRoot, r, api)
	if err != nil {
		panic(err)
	}
}
