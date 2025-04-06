package coder

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/codes"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
	"github.com/yunhanshu-net/runcher/pkg/osx"
	"github.com/yunhanshu-net/runcher/status"
	"os"
	"os/exec"
	"strings"
)

type Golang struct {
	runnerRoot string
	runner     *model.Runner
}

func (g *Golang) AddApi(codeApi *codex.CodeApi) error {
	//nextVersion := runner.GetNextVersion()
	runnerRoot := g.runnerRoot
	runner := g.runner
	//currentVersionWorkPath := runner.GetInstallPath(runnerRoot)
	pathInfo := runner.GetPaths(runnerRoot)
	//nextVersionWorkPath, err := runner.GetNextVersionInstallPath(runnerRoot)
	//if err != nil {
	//	return err
	//}
	codeApi.Language = g.runner.Language

	addFileSavePath, addFileAbsFile := codeApi.GetFileSaveFullPath(pathInfo.NextVersionPath)
	if osx.DirExists(addFileSavePath) { //先判断package是否存在
		return status.ErrorCodeApiFileExist.WithMessage(addFileSavePath)
	}

	//先判断是否存在
	if osx.FileExists(addFileAbsFile) {
		return status.ErrorCodeApiFileExist.WithMessage(addFileAbsFile)
	}

	//假如当前版本是v1，新增一个api，直接编译当前v1目录的main.go文件，
	//失败了就把当前目录改为v1_err把v1_bak复制一份为v1，回滚为失败前的代码
	//编译成功后把v1改为v2 把v1_bak改为v1 把当前v2目录copy一份到v2_bak保证每次编译成功后都有源码的备份文件
	//err := g.createFile(addFileAbsFile, codeApi.Code)
	//if err != nil {
	//	return err
	//}
	//err = g.buildRunner(pathInfo.CurrentVersionPath, runner.GetBuildPath(runnerRoot), runner.GetBuildRunnerName())
	//if err != nil {
	//	//失败了就把当前目录改为v1_err把v1_bak复制一份为v1，回滚为失败前的代码
	//	err := os.Rename(pathInfo.CurrentVersionPath, pathInfo.CurrentVersionErrPath)
	//	if err != nil {
	//		fmt.Printf(" os.Rename(pathInfo.CurrentVersionPath, pathInfo.CurrentVersionErrPath) err:%s:%+v\n", err, pathInfo)
	//	}
	//	err = osx.CopyDirectory(pathInfo.CurrentVersionBakPath, pathInfo.CurrentVersionPath) //把v1_bak复制一份为v1，回滚为失败前的代码
	//	if err != nil {
	//		return err
	//	}
	//
	//	return err
	//}
	//
	////下面三个操作可以异步后台慢慢执行
	//err = os.Rename(pathInfo.CurrentVersionPath, pathInfo.NextVersionPath)
	//if err != nil {
	//	return err
	//}
	//err = os.Rename(pathInfo.CurrentVersionBakPath, pathInfo.CurrentVersionPath)
	//if err != nil {
	//	return err
	//}
	//err = osx.CopyDirectory(pathInfo.NextVersionPath, pathInfo.NextVersionBakPath)
	//if err != nil {
	//	return err
	//}

	//
	//创建新版本工作目录
	err := os.MkdirAll(pathInfo.NextVersionPath, 0755)
	if err != nil {
		return err
	}

	//copy 旧代码到新版本工作目录
	err = osx.CopyDirectory(pathInfo.CurrentVersionPath, pathInfo.NextVersionPath) //把当前项目代码保存一份复制到下一个版本
	if err != nil {
		return err
	}

	err = g.createFile(addFileAbsFile, codeApi.Code)
	if err != nil {
		return err
	}
	err = g.buildRunner(pathInfo.NextVersionPath, runner.GetBuildPath(runnerRoot), runner.GetBuildRunnerName())
	if err != nil {
		return err
	}
	//return nil
	return nil
}

func (g *Golang) createFile(filePath string, content string) error {
	codeFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer codeFile.Close()
	_, err = codeFile.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func (g *Golang) buildRunner(workDir string, buildPath string, runnerName string) error {
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = workDir
	err := tidy.Run()
	if err != nil {
		return err
	}
	// 创建命令
	cmd := exec.Command("go", "build", "-o", "../../bin/"+runnerName)

	// 设置工作目录（可选）
	cmd.Dir = workDir // 当前目录，可以根据需要修改为项目路径

	output, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	exists := osx.FileExists(buildPath + "/" + runnerName)
	if !exists {
		return status.ErrorCodeApiBuildError.WithMessage(workDir)
	}
	return nil
}

func (g *Golang) AddBizPackage(codeBizPackage *codex.BizPackage) error {
	runner := g.runner
	currentVersionPath := runner.GetInstallPath(g.runnerRoot)
	_, absPkgPath := codeBizPackage.GetPackageSaveFullPath(currentVersionPath)
	if osx.DirExists(absPkgPath) { //先判断Package是否存在
		return status.ErrorCodeApiFileExist.WithMessage(absPkgPath)
	}
	//不存在才可以创建
	err := os.MkdirAll(absPkgPath, 0755)
	if err != nil {
		return err
	}

	//在pkg下创建_init.go文件
	err = g.createFile(absPkgPath+"/init_.go", fmt.Sprintf(codes.InitCodeTemplate, codeBizPackage.EnName))
	if err != nil {
		return err
	}
	manager := codex.NewGolangProjectManager(currentVersionPath)
	err = manager.AddPackages([]codex.PackageInfo{
		{
			Alias:      strings.ReplaceAll(codeBizPackage.AbsPackagePath, "/", "_"),
			ImportPath: fmt.Sprintf("git.yunhanshu.net/%s/%s/api/%s", runner.User, runner.Name, codeBizPackage.AbsPackagePath),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *Golang) CreateProject() error {
	runner := g.runner
	err := os.MkdirAll(runner.GetToolPath(g.runnerRoot), 0755) //初始化项目目录
	if err != nil {
		return err
	}

	//go.mod
	//main.go
	//bin
	//	.request
	//	user_app_v1
	//  user_app_v2
	//version
	//	-v1
	//		-api
	//  -v2
	//		-api

	codePath := runner.GetToolPath(g.runnerRoot) + "/version/" + runner.Version
	err = os.MkdirAll(codePath, 0755) //初始版本目录
	if err != nil {
		return err
	}
	err = os.MkdirAll(codePath+"/api", 0755) //初始api目录
	if err != nil {
		return err
	}

	err = os.MkdirAll(runner.GetToolPath(g.runnerRoot)+"/bin/.request", 0755) //初始化可执行程序目录
	if err != nil {
		return err
	}

	//创建 go.mod 文件
	err = g.createFile(codePath+"/go.mod", fmt.Sprintf(`
module git.yunhanshu.net/%s/%s

go 1.23`, runner.User, runner.Name))
	if err != nil {
		return err
	}
	//创建main文件
	manager := codex.NewGolangProjectManager(codePath)
	err = manager.CreateMain(nil)
	if err != nil {
		return err
	}
	err = g.buildRunner(codePath, runner.GetBuildPath(g.runnerRoot), runner.GetBuildRunnerCurrentVersionName())
	if err != nil {
		return err
	}
	return nil
}
