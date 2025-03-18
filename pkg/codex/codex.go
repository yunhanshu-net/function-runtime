package codex

import "strings"

type CodeApi struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Package  string `json:"package"`
	EnName   string `json:"en_name"`
	CnName   string `json:"cn_name"`
	Desc     string `json:"desc"`
}

func (c *CodeApi) GetFileSaveFullPath(sourceCodeDir string) (fullPath string, absFilePath string) {
	fullPath = strings.TrimSuffix(sourceCodeDir, "/") + "/api/" + c.Package
	absFilePath = fullPath + "/" + c.GetFileName()
	return
}

func (c *CodeApi) GetFileName() string {
	return c.EnName + "." + c.Language
}
