package request

import (
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"time"
)

type RollbackVersion struct {
	RunnerConf *model.Runner `json:"runner_conf"`
	OldVersion string        `json:"old_version"`
}

func (c *Request) IsOpenCommand() bool {
	return c.SoftInfo.Command == "_cloud_func" || c.SoftInfo.Command == "_docs_info_text"
}

func (c *Request) RequestJSON() (string, error) {
	j, err := json.Marshal(c.Body)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func (c *Request) GetRequestFilePath(callerPath string) string {
	reqJson := callerPath + fmt.Sprintf("/.request/%v_%v.json",
		c.SoftInfo.Soft, time.Now().UnixNano())
	return reqJson
}
