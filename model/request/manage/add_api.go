package manage

import (
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
)

type AddApi struct {
	Runner  *model.Runner  `json:"runner"`
	CodeApi *codex.CodeApi `json:"code_api"`
}

type AddApis struct {
	Runner   *model.Runner
	CodeApis []*codex.CodeApi
}
