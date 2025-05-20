package coder

import (
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/runcher/model/dto/syscallback"
	"github.com/yunhanshu-net/sdk-go/pkg/dto/api"
	"path/filepath"
	"strings"
)

type BizPackage struct {
	Runner *runnerproject.Runner `json:"runner"`

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
func (c *BizPackage) GetPackageAbsPath(apiPath string) (absPackagePath string) {
	return filepath.Join(apiPath, c.AbsPackagePath)
}

func (c *BizPackage) GetPackageName() string {
	return c.EnName
}

type CreateProjectReq struct {
	runnerproject.Runner
}
type CreateProjectResp struct {
	Version string `json:"version"`
}

type ApiChangeInfo struct {
	CurrentVersion string      `json:"current_version"` //此次更新的版本
	AddApi         []*api.Info `json:"add_api"`         //此次新增的api
	DelApi         []*api.Info `json:"del_api"`         //此次删除的api
	UpdateApi      []*api.Info `json:"update_api"`      //此次变更的api
}
type AddApisResp struct {
	Version       string               `json:"version"`
	ErrList       []*CodeApiCreateInfo `json:"err_list"`
	ApiChangeInfo *ApiChangeInfo       `json:"api_change_info"`
}

type AddApiResp struct {
	Version              string                              `json:"version"`
	Data                 interface{}                         `json:"data"`
	SyscallChangeVersion *syscallback.SysOnVersionChangeResp `json:"syscall_change_version"`
}

type BizPackageResp struct {
	Version string `json:"version"`
}
