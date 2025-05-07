package coder

import (
	"context"
	"fmt"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/dto/coder"
	"github.com/yunhanshu-net/runcher/pkg/codes"
	"github.com/yunhanshu-net/runcher/pkg/codex"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/runcher/pkg/osx"
	"github.com/yunhanshu-net/runcher/status"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Golang struct {
	runnerRoot string
	runner     *model.Runner
}

func (g *Golang) AddApi(ctx context.Context, codeApi *coder.CodeApi) (*coder.AddApiResp, error) {
	runner := g.runner
	pathInfo, err := g.addApi(ctx, codeApi)
	if err != nil {
		return nil, err
	}
	err = g.buildRunner(ctx, pathInfo.NextVersionPath, runner.GetBuildPath(g.runnerRoot), runner.GetBuildRunnerName())
	if err != nil {
		logger.ErrorContextf(ctx, "程序编译失败：%s", err.Error())
		return nil, err
	}
	return &coder.AddApiResp{Version: runner.GetNextVersion()}, nil
}

func (g *Golang) createFile(ctx context.Context, filePath string, content string) error {
	codeFile, err := os.Create(filePath)
	if err != nil {
		logger.Infof("[createFile] Create filePath:%s err: %v", filePath, err)
		return err
	}
	defer codeFile.Close()
	_, err = codeFile.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func (g *Golang) buildRunner(ctx context.Context, workDir string, buildPath string, runnerName string) error {
	logger.InfoContextf(ctx, "workDir:%s\nbuildPath:%s\n runnerName:%s\n", workDir, buildPath, runnerName)
	// 1. 检查 workDir 是否是有效的 Go 模块目录
	if !osx.FileExists(filepath.Join(workDir, "go.mod")) {
		return fmt.Errorf("workDir %s is not a Go module root", workDir)
	}

	// 2. 确保构建目录存在
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %v", err)
	}

	version := exec.Command("go", "version")
	version.Dir = workDir
	fmt.Println("version:")
	versionBt, err := version.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go version: %v\n%s", err, string(versionBt))
	}

	// 3. 执行 go mod tidy
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = workDir
	if output, err := tidy.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %v\n%s", err, string(output))
	}

	// 4. 构建二进制文件
	outputPath := filepath.Join(buildPath, runnerName)
	cmd := exec.Command("go", "build", "-o", outputPath)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	s := string(output)
	if err != nil {
		logger.ErrorContextf(ctx, "Build failed error ：%s output %s", err, s)
		return fmt.Errorf("build failed: %v\n%s", err, s)
	}

	// 5. 记录构建日志（正常输出）
	logger.InfoContextf(ctx, "Build output output:%s", s)

	// 6. 验证生成的文件
	if !osx.FileExists(outputPath) {
		return status.ErrorCodeApiBuildError.WithMessage(fmt.Sprintf("binary not found at %s", outputPath))
	}

	return nil
}

func (g *Golang) AddBizPackage(ctx context.Context, codeBizPackage *coder.BizPackage) (*coder.BizPackageResp, error) {
	runner := g.runner
	currentVersionPath := runner.GetInstallPath(g.runnerRoot)
	_, absPkgPath := codeBizPackage.GetPackageSaveFullPath(currentVersionPath)
	if osx.DirExists(absPkgPath) { //先判断Package是否存在
		return nil, status.ErrorCodeApiFileExist.WithMessage(absPkgPath)
	}
	//不存在才可以创建
	err := os.MkdirAll(absPkgPath, 0755)
	if err != nil {
		return nil, err
	}

	//在pkg下创建_init.go文件
	err = g.createFile(ctx, absPkgPath+"/init_.go", fmt.Sprintf(codes.InitCodeTemplate, codeBizPackage.EnName))
	if err != nil {
		return nil, err
	}
	manager := codex.NewGolangProjectManager(currentVersionPath)
	err = manager.AddPackages([]codex.PackageInfo{
		{
			Alias:      strings.ReplaceAll(codeBizPackage.AbsPackagePath, "/", "_"),
			ImportPath: fmt.Sprintf("git.yunhanshu.net/%s/%s/api/%s", runner.User, runner.Name, codeBizPackage.AbsPackagePath),
		},
	})
	if err != nil {
		return nil, err
	}

	return &coder.BizPackageResp{Version: runner.Version}, nil
}

func (g *Golang) CreateProject(ctx context.Context) (*coder.CreateProjectResp, error) {
	logger.Infof("Create project:%+v", g.runner)
	runner := g.runner
	err := os.MkdirAll(runner.GetToolPath(g.runnerRoot), 0755) //初始化项目目录
	if err != nil {
		return nil, err
	}

	//go.mod
	//main.go
	//bin
	//	api-logs //存储每个版本的api配置
	//	data
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
		return nil, err
	}
	err = os.MkdirAll(codePath+"/api", 0755) //初始api目录
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(runner.GetToolPath(g.runnerRoot)+"/bin/.request", 0755) //初始化可执行程序目录
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(runner.GetToolPath(g.runnerRoot)+"/bin/data", 0755) //初始化数据库目录
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(runner.GetToolPath(g.runnerRoot)+"/bin/conf", 0755) //初始化配置文件目录
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(runner.GetToolPath(g.runnerRoot)+"/bin/logs", 0755) //初始化logs目录
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(runner.GetToolPath(g.runnerRoot)+"/bin/api-logs", 0755) //初始化api-logs目录
	if err != nil {
		return nil, err
	}

	pkg := fmt.Sprintf("git.yunhanshu.net/%s/%s", runner.User, runner.Name)

	//创建 go.mod 文件
	//	err = g.createFile(ctx, codePath+"/go.mod", fmt.Sprintf(`
	//module git.yunhanshu.net/%s/%s
	//
	//go 1.23`, runner.User, runner.Name))
	//go mod init

	tidy := exec.Command("go", "mod", "init", pkg)
	tidy.Dir = codePath
	err = tidy.Run()
	if err != nil {
		return nil, err
	}
	exists := osx.FileExists(codePath + "/go.mod")
	if !exists {
		return nil, fmt.Errorf("go mod init err:%s", err)
	}

	//创建main文件
	manager := codex.NewGolangProjectManager(codePath)
	err = manager.CreateMain(nil)
	if err != nil {
		return nil, err
	}
	err = g.buildRunner(ctx, codePath, runner.GetBuildPath(g.runnerRoot), runner.GetBuildRunnerCurrentVersionName())
	if err != nil {
		return nil, err
	}
	return &coder.CreateProjectResp{Version: runner.Version}, nil
}

func (g *Golang) addApi(ctx context.Context, api *coder.CodeApi) (*model.RunnerPath, error) {
	runnerRoot := g.runnerRoot
	runner := g.runner
	if g.runner.Version == "" {
		version, err := runner.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		runner.Version = version
	}
	pathInfo := runner.GetPaths(runnerRoot)
	api.Language = g.runner.Language

	_, addFileAbsFile := api.GetFileSaveFullPath(pathInfo.NextVersionPath)
	//if osx.DirExists(addFileSavePath) { //先判断package是否存在
	//	return nil, status.ErrorCodeApiFileExist.WithMessage(addFileSavePath)
	//}

	//先判断文件是否已经存在
	if osx.FileExists(addFileAbsFile) {
		return nil, status.ErrorCodeApiFileExist.WithMessage(addFileAbsFile)
	}

	exists := osx.DirExists(pathInfo.NextVersionPath)
	if !exists {
		//创建新版本工作目录，把当前项目代码复制到下一个版本
		err := os.MkdirAll(pathInfo.NextVersionPath, 0755)
		if err != nil {
			return nil, err
		}
		err = osx.CopyDirectory(pathInfo.CurrentVersionPath, pathInfo.NextVersionPath) //把当前项目代码保存一份复制到下一个版本
		if err != nil {
			return nil, err
		}
	}

	err := g.createFile(ctx, addFileAbsFile, api.Code) //创建新文件
	if err != nil {
		return nil, err
	}

	return &pathInfo, nil
}

func (g *Golang) AddApis(ctx context.Context, codeApis []*coder.CodeApi) (resp *coder.AddApisResp, err error) {

	resp = new(coder.AddApisResp)
	if len(codeApis) == 0 {
		return nil, fmt.Errorf("no api")
	}
	var errs []*coder.CodeApiCreateInfo
	var pathInfo *model.RunnerPath
	for _, codeApi := range codeApis {
		info, err := g.addApi(ctx, codeApi)
		if err != nil {
			errs = append(errs, &coder.CodeApiCreateInfo{
				Language:       codeApi.Language,
				Package:        codeApi.Package,
				AbsPackagePath: codeApi.AbsPackagePath,
				EnName:         codeApi.EnName,
				CnName:         codeApi.CnName,
				Msg:            err.Error(),
				Status:         "failed",
			})
		}
		if info != nil {
			pathInfo = info
		}
	}
	if pathInfo == nil {
		return nil, fmt.Errorf("pathInfo is nil")
	}
	resp.ErrList = errs
	resp.Version = g.runner.GetNextVersion()

	err = g.buildRunner(ctx, pathInfo.NextVersionPath, g.runner.GetBuildPath(g.runnerRoot), g.runner.GetBuildRunnerName())
	if err != nil {
		return nil, err
	}
	return resp, nil
}
