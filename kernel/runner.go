package kernel

import (
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/runner"
	"github.com/yunhanshu-net/runcher/runtime"
	"github.com/yunhanshu-net/runcher/transport"
	"time"
)

func (r *Executor) HandelRunnerStatus(msg *transport.Msg) {

}

func (r *Executor) NewRunner(req *request.RunnerRequest) (*runtime.Runner, error) {

	transportInstance, err := r.GetTransport(req.TransportConfig.Type)
	if err != nil {
		return nil, err
	}
	instance := runner.NewRunner(req.Runner)
	return &runtime.Runner{
		StartTime: time.Now(),
		Instance:  instance,
		Transport: transportInstance,
	}, nil
}
