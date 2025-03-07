package runtime

import (
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"github.com/yunhanshu-net/runcher/transport"
	"time"
)

// Runner runner 运行时
type Runner struct {
	StartTime time.Time
	Instance  runner.Runner
	Transport transport.Transport
	Status    string
}

func (r *Runner) Request(req *request.RunnerRequest) (*response.RunnerResponse, error) {
	return r.Instance.Request(req, &runner.Context{Transport: r.Transport, Status: r.Status})
}
