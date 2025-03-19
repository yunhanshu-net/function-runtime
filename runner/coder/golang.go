package coder

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/pkg/codex"
	"github.com/yunhanshu-net/runcher/pkg/osx"
	"github.com/yunhanshu-net/runcher/status"
	"os"
	"os/exec"
)

type Golang struct {
}

func (g *Golang) AddApi(runnerRoot string, runner *model.Runner, codeApi *codex.CodeApi) error {
	currentVersionPath := runner.GetInstallPath(runnerRoot)
	nextVersionPath, err := runner.GetNextVersionInstallPath(runnerRoot)
	if err != nil {
		return err
	}

	fileSavePath, absFile := codeApi.GetFileSaveFullPath(nextVersionPath)
	if osx.DirExists(fileSavePath) {
		return status.ErrorCodeApiFileExist.WithMessage(fileSavePath)
	} else {
		err = os.MkdirAll(fileSavePath, 0755)
		if err != nil {
			return err
		}
	}

	//todo先判断文件是否存在
	if osx.FileExists(absFile) {
		return status.ErrorCodeApiFileExist.WithMessage(absFile)
	}

	err = os.MkdirAll(nextVersionPath, 0755)
	if err != nil {
		return err
	}

	err = osx.CopyDirectory(currentVersionPath, nextVersionPath) //把当前项目代码保存一份复制到下一个版本
	if err != nil {
		return err
	}

	codeFile, err := os.Create(absFile)
	if err != nil {
		return err
	}
	defer codeFile.Close()
	_, err = codeFile.WriteString(codeApi.Code)
	if err != nil {
		return err
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = nextVersionPath
	err = tidy.Run()
	if err != nil {
		return err
	}
	// 创建命令
	cmd := exec.Command("go", "build", "-o", runner.Name)

	// 设置工作目录（可选）
	cmd.Dir = nextVersionPath // 当前目录，可以根据需要修改为项目路径

	output, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	exists := osx.FileExists(nextVersionPath + "/" + runner.Name)
	if !exists {
		return status.ErrorCodeApiBuildError.WithMessage(absFile)
	}
	return nil
}
