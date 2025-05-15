package coder

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"path/filepath"
)

type GoCoder struct {
	*model.Runner

	runnerRoot     string
	IsDev          bool
	UseGoMod       bool
	ImportPathRoot string
	ModuleRoot     string
	saveRoot       string
	savePath       string //当前项目的存储位置
}

func NewGoCoderV2() (*GoCoder, error) {
	//root := conf.GetRunnerRoot()
	//isDev := conf.IsDev()
	//debugImport := ""
	//g := &GoCoder{
	//	saveRoot:       root,
	//	IsDev:          isDev,
	//	UseGoMod:       !isDev,
	//	ImportPathRoot: root,
	//}
	return &GoCoder{}, nil
}

func (g *GoCoder) GetImportPath(addPkgPath string) string {
	if g.IsDev {
		return filepath.Join(fmt.Sprintf("github.com/yunhanshu-net/sdk-go/soft/%s/%s/debug/api", g.User, g.Name), addPkgPath)
	}
	return filepath.Join(fmt.Sprintf("git.yunhanshu.net/%s/%s/api", g.User, g.Name), addPkgPath)
}
