package runtime

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"time"
)

// Runner runner 运行时
type Runner struct {
	Conn      *nats.Conn
	UUID      string        `json:"uid"`
	StartTime time.Time     `json:"start_time"`
	Instance  runner.Runner `json:"instance"`
	Status    string        `json:"status"`
}

func (r *Runner) Request(req *request.Context) (*response.Response, error) {
	var res response.Response
	req.Request.UUID = r.UUID
	subject := req.GetSubject()
	msg, err := r.Conn.Request(subject, req.GetData(), time.Second*20)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(msg.Data, &res.Response)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
