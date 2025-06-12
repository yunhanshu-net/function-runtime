package coder

import (
	"fmt"
	"github.com/yunhanshu-net/function-go/pkg/dto/api"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"path/filepath"
)

type BizPackage struct {
	Runner *runnerproject.Runner `json:"runner"`

	AbsPackagePath string `json:"abs_package_path"`
	Language       string `json:"language"`
	EnName         string `json:"en_name"`
	CnName         string `json:"cn_name"`
	Desc           string `json:"desc"`
}

func (c *BizPackage) GetSubPackagePath() string {
	return c.AbsPackagePath
}

func (c *BizPackage) GetPackageAbsPath(apiPath string) (absPackagePath string) {
	return filepath.Join(apiPath, c.GetSubPackagePath())
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

func (a *ApiChangeInfo) GetAddApisDesc() string {
	add := ""
	for _, v := range a.AddApi {
		add += fmt.Sprintf("%s:(%s)\n", v.ChineseName, v.Router)
	}
	return add
}
func (a *ApiChangeInfo) GetDelApisDesc() string {
	add := ""
	for _, v := range a.DelApi {
		add += fmt.Sprintf("%s(%s)\n", v.ChineseName, v.Router)
	}
	return add
}
func (a *ApiChangeInfo) GetUpdateApisDesc() string {
	add := ""
	for _, v := range a.DelApi {
		add += fmt.Sprintf("%s(%s)\n", v.ChineseName, v.Router)
	}
	return add
}

func (a *ApiChangeInfo) GetChangeLog() string {
	return fmt.Sprintf("新增API:%s\n删除API:%s\n变更API:%s", a.GetAddApisDesc(), a.GetDelApisDesc(), a.GetUpdateApisDesc())
}

type AddApisResp struct {
	Hash          string               `json:"hash"`
	Version       string               `json:"version"`
	ErrList       []*CodeApiCreateInfo `json:"err_list"`
	ApiChangeInfo *ApiChangeInfo       `json:"api_change_info"`
}

type DeleteAPIsResp struct {
	Hash    string      `json:"hash"`
	DelApis []*api.Info `json:"del_apis"`
	Version string      `json:"version"`
}

func (a *DeleteAPIsResp) GetDelApisDesc() string {
	add := ""
	for _, v := range a.DelApis {
		add += fmt.Sprintf("%s(%s)\n", v.ChineseName, v.Router)
	}
	return add
}

type BizPackageResp struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
}
