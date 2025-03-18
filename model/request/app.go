package request

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/model"
)

type Context struct {
	Conn    *nats.Conn     `json:"-"`
	Msg     *nats.Msg      `json:"-"`
	Request *RunnerRequest `json:"request"`
}

type RollbackVersion struct {
	RunnerConf *model.Runner `json:"runner_conf"`
	OldVersion string        `json:"old_version"`
}

//func (c *Request) IsOpenCommand() bool {
//	return c.RunnerInfo.Command == "_cloud_func" || c.RunnerInfo.Command == "_docs_info_text"
//}

func (c *Request) RequestJSON() (string, error) {
	j, err := json.Marshal(c.Body)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

//func (c *Request) GetRequestFilePath(callerPath string) string {
//	reqJson := callerPath + fmt.Sprintf("/.request/%v_%v.json",
//		c.RunnerInfo.Soft, time.Now().UnixNano())
//	return reqJson
//}
