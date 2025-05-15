package coder

import (
	"context"
	"fmt"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"os"
	"path/filepath"
)

type GoCoder struct {
	*model.Runner

	runnerRoot   string
	IsDev        bool
	UseGoMod     bool
	SavePathRoot string
	ModuleRoot   string //github.com/yunhanshu-net/sdk-go prod
	CodePath     string //存放目录代码 /soft/user/name/code
	BinPath      string //可执行程序存放目录 soft/user/name/workplace/bin
	ExecPath     string //执行目录 /soft/user/name/workplace
	saveRoot     string //整个项目的根目录soft/
	savePath     string //当前项目的存储位置 /soft/user/name
}

func NewGoCoderV2() (*GoCoder, error) {
	root := conf.GetRunnerRoot()
	isDev := conf.IsDev()
	g := &GoCoder{
		saveRoot:     root,
		IsDev:        isDev,
		UseGoMod:     !isDev,
		SavePathRoot: root,
	}
	return g, nil
}

func (g *GoCoder) GetImportPath(addPkgPath string) string {
	if g.IsDev {
		return filepath.Join(fmt.Sprintf("github.com/yunhanshu-net/sdk-go/soft/%s/%s/code/api", g.User, g.Name), addPkgPath)
	}
	return filepath.Join(fmt.Sprintf("git.yunhanshu.net/%s/%s/api", g.User, g.Name), addPkgPath)
}

func (g *GoCoder) mkdirAll(ctx context.Context) error {
	err := os.MkdirAll(g.savePath, 0755) //初始化项目目录
	if err != nil {
		return err
	}

	err = os.MkdirAll(g.savePath+"/code/api", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.savePath+"/workplace/api-logs", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.savePath+"/workplace/bin", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.savePath+"/workplace/bin/releases", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.savePath+"/workplace/conf", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.savePath+"/workplace/data", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.savePath+"/workplace/logs", 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.savePath+"/workplace/metadata", 0755)
	if err != nil {
		return err
	}
	return nil
}

func (g *GoCoder) CreateProject(ctx context.Context) (*coder.CreateProjectResp, error) {
	logger.Infof(ctx, "Create project:%+v", g.Runner)
	err := g.mkdirAll(ctx)
	if err != nil {
		return nil, err
	}
	//GenMainGo(nil)

	return nil, nil
}
