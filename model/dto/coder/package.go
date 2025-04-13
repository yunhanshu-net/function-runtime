package coder

import "strings"

type BizPackage struct {
	AbsPackagePath string `json:"abs_package_path"`
	Language       string `json:"language"`
	EnName         string `json:"en_name"`
	CnName         string `json:"cn_name"`
	Desc           string `json:"desc"`
}

func (c *BizPackage) GetPackageSaveFullPath(sourceCodeDir string) (savePath string, absPackagePath string) {
	savePath = strings.TrimSuffix(sourceCodeDir, "/") + "/api"
	absPackagePath = savePath + "/" + c.AbsPackagePath
	return savePath, absPackagePath
}

func (c *BizPackage) GetPackageName() string {
	return c.EnName
}
