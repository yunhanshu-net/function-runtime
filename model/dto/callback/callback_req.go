package callback

import (
	"fmt"
)

type OnApiCreatedReq struct {
	Method string `json:"method"`
	Router string `json:"router"`
}

type BeforeApiDeleteReq struct {
	Method string `json:"method"`
	Router string `json:"router"`
}

type AfterApiDeletedReq struct {
	Method string `json:"method"`
	Router string `json:"router"`
}

type BeforeRunnerCloseReq struct {
}

type AfterRunnerCloseReq struct {
}

type ChangeReq struct {
	Method string `json:"method"`
	Router string `json:"router"`
	Type   string `json:"type"`
}

func (c *ChangeReq) String() string {
	return fmt.Sprintf(`{"method": "%s", "router": "%s","type","%s"}`, c.Method, c.Router, c.Type)
}

type OnVersionChange struct {
	Change []ChangeReq `json:"change"`
}

type OnInputFuzzyReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type OnInputValidateReq struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type OnTableDeleteRowsReq struct {
	Ids []string `json:"ids"`
}

type OnTableUpdateRowReq struct {
	Ids []string `json:"ids"`
}

type OnTableSearchReq struct {
	Cond map[string]string `json:"cond"`
}
type OnFuzzyReq struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

//func (c *Callback) BindData(req interface{}) error {
//	err := json.Unmarshal([]byte(c.Body), &req)
//	if err != nil {
//		return err
//	}
//	return nil
//}
