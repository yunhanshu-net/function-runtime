package request

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/runcher/pkg/constants"
)

type Request struct {
	UUID       string              `json:"uuid"`
	TraceID    string              `json:"trace_id"`
	Route      string              `json:"route"`
	Method     string              `json:"method"`
	Headers    map[string]string   `json:"headers"`
	Body       interface{}         `json:"body"`        //请求json
	BodyString string              `json:"body_string"` //请求json
	FileMap    map[string][]string `json:"file_map"`
}

func (r *Request) WithContext(ctx context.Context) *Request {
	value := ctx.Value(constants.TraceID)
	if value != nil {
		r.TraceID = value.(string)
	}
	return r
}

type RunnerRequest struct {
	UUID string `json:"uuid"`

	Runner          *runnerproject.Runner `json:"runner"`
	TransportConfig *TransportConfig      `json:"transport_config"`

	Request *Request `json:"request"`
	//Body    interface{} `json:"body"`
}

type TransportConfig struct {
	IdleTime int                    `json:"idle_time"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (r *RunnerRequest) GetSubject() string {
	return fmt.Sprintf("runner.%s.%s.%s.run", r.Runner.User, r.Runner.Name, r.Runner.Version)
}
func (r *RunnerRequest) Bytes() []byte {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return jsonBytes
}
