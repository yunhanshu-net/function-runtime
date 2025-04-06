package request

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model"
)

type Context struct {
	//Conn     *nats.Conn `json:"-"`
	Msg      *nats.Msg `json:"-"`
	HttpData []byte
	Request  *Request `json:"request"`
	Type     string   `json:"type"`
}

func (c *Context) GetSubject() string {
	if c.Msg == nil {
		return fmt.Sprintf("runner.%s.%s.%s.run", c.Request.Runner.User, c.Request.Runner.Name, c.Request.Runner.Version)
	}
	return c.Msg.Subject
}

func (c *Context) GetData() []byte {
	if c.Type == "" || c.Type == "nats" {
		return c.Msg.Data
	}
	marshal, err := json.Marshal(c.Request)
	if err != nil {
		return nil
	}
	return marshal
}

type RollbackVersion struct {
	RunnerConf *model.Runner `json:"runner_conf"`
	OldVersion string        `json:"old_version"`
}

//func (c *RunnerRequest) IsOpenCommand() bool {
//	return c.RunnerInfo.Command == "_cloud_func" || c.RunnerInfo.Command == "_docs_info_text"
//}

func (c *RunnerRequest) RequestJSON() (string, error) {
	j, err := json.Marshal(c.Body)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

//func (c *RunnerRequest) GetRequestFilePath(callerPath string) string {
//	reqJson := callerPath + fmt.Sprintf("/.request/%v_%v.json",
//		c.RunnerInfo.Soft, time.Now().UnixNano())
//	return reqJson
//}
