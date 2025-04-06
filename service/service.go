package service

import (
	"context"
	"github.com/yunhanshu-net/runcher/kernel"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
)

type RunnerHttp struct {
	R *kernel.Runcher
}

func (r *RunnerHttp) Get(context context.Context, req *request.Request) (*response.Response, error) {
	req.Request.Method = "GET"
	rr := &request.Context{
		Request: req,
		Type:    "http",
	}
	runnerResponse, err := r.R.Scheduler.Request(rr)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil

}
