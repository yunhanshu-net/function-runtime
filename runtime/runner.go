package runtime

import (
	"encoding/json"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"time"
)

// Runner runner 运行时
type Runner struct {
	UUID      string        `json:"uid"`
	StartTime time.Time     `json:"start_time"`
	Instance  runner.Runner `json:"instance"`
	Status    string        `json:"status"`
}

func (r *Runner) Request(req *request.Context) (*response.RunnerResponse, error) {
	var res response.RunnerResponse
	req.Request.UUID = r.UUID
	subject := req.Request.GetSubject()
	msg, err := req.Conn.Request(subject, req.Msg.Data, time.Second*20)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(msg.Data, &res.Response)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
