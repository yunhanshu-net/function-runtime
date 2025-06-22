package coder

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/yunhanshu-net/function-go/pkg/dto/api"
	"github.com/yunhanshu-net/function-go/pkg/dto/response"
	"github.com/yunhanshu-net/function-go/pkg/dto/usercall"
	"github.com/yunhanshu-net/function-runtime/conf"
	"github.com/yunhanshu-net/function-runtime/pkg/dto/coder"
	"github.com/yunhanshu-net/function-runtime/status"
	"github.com/yunhanshu-net/pkg/constants"
	usercallConst "github.com/yunhanshu-net/pkg/constants/usercall"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/logger"
	"github.com/yunhanshu-net/pkg/x/cmdx"
	"github.com/yunhanshu-net/pkg/x/jsonx"
	"github.com/yunhanshu-net/pkg/x/osx"
	"github.com/yunhanshu-net/pkg/x/slicesx"
	"github.com/yunhanshu-net/pkg/x/stringsx"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

type GoCoder struct {
	*runnerproject.Runner

	IsDev    bool
	UseGoMod bool
	//CurrentBuildName string //user_name_v1
	//NextBuildName    string //user_name_v2
	ModuleRoot   string //github.com/yunhanshu-net/function-go prod
	SaveRoot     string //整个项目的根目录soft/
	SavePath     string //当前项目的存储位置 /soft/$user/$name
	CodePath     string //存放目录代码 /soft/$user/$name/code/cmd
	MainPath     string //执行目录 /soft/$user/$name/code/cmd/app
	MainFile     string //执行目录 /soft/$user/$name/code/cmd/app/main.go
	ApiPath      string //执行目录 /soft/$user/$name/code/api
	BinPath      string //可执行程序存放目录 soft/$user/$name/workplace/bin
	ReleasesPath string //可执行程序存放目录 soft/$user/$name/workplace/bin/releases
	MetaDataPath string //元数据存放目录 soft/$user/$name/workplace/metadata
	ApiLogsPath  string //api日志存放目录 soft/$user/$name/workplace/api-logs
	RequestPath  string //soft/$user/$name/workplace/bin/.request
}

func (g *GoCoder) GetNextBuildName() string {
	nextVersion := g.Runner.GetNextVersion()
	return fmt.Sprintf("%s_%s_%s", g.Runner.User, g.Runner.Name, nextVersion)
}

func (g *GoCoder) GetCurrentBuildName() string {
	return fmt.Sprintf("%s_%s_%s", g.Runner.User, g.Runner.Name, g.Runner.Version)
}

func NewGoCoderV2(runner *runnerproject.Runner) (*GoCoder, error) {
	root := conf.GetRunnerRoot()
	isDev := conf.IsDev()
	g := &GoCoder{
		IsDev:        isDev,
		Runner:       runner,
		UseGoMod:     !isDev,
		SaveRoot:     root,
		SavePath:     filepath.Join(root, runner.User, runner.Name),
		CodePath:     filepath.Join(root, runner.User, runner.Name, "code"),
		ApiPath:      filepath.Join(root, runner.User, runner.Name, "code", "api"),
		MainPath:     filepath.Join(root, runner.User, runner.Name, "code", "cmd", "app"),
		MainFile:     filepath.Join(root, runner.User, runner.Name, "code", "cmd", "app", "main.go"),
		BinPath:      filepath.Join(root, runner.User, runner.Name, "workplace", "bin"),
		ReleasesPath: filepath.Join(root, runner.User, runner.Name, "workplace", "bin", "releases"),
		MetaDataPath: filepath.Join(root, runner.User, runner.Name, "workplace", "metadata"),
		ApiLogsPath:  filepath.Join(root, runner.User, runner.Name, "workplace", "api-logs"),
		RequestPath:  filepath.Join(root, runner.User, runner.Name, "workplace", "bin", ".request"),
	}
	return g, nil
}

func (g *GoCoder) name() {

}

func (g *GoCoder) mkdirAll(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	err := os.MkdirAll(g.SavePath, 0755) //初始化项目目录
	err = os.MkdirAll(g.ApiPath, 0755)
	err = os.MkdirAll(g.RequestPath, 0755)
	err = os.MkdirAll(g.MainPath, 0755)
	err = os.MkdirAll(g.ReleasesPath, 0755)
	err = os.MkdirAll(g.MetaDataPath, 0755)
	err = os.MkdirAll(g.BinPath, 0755) //初始化项目目录
	err = os.MkdirAll(g.SavePath+"/workplace/api-logs", 0755)
	err = os.MkdirAll(g.SavePath+"/workplace/conf", 0755)
	err = os.MkdirAll(g.SavePath+"/workplace/data", 0755)
	err = os.MkdirAll(g.SavePath+"/workplace/logs", 0755)
	if err != nil {
		return err
	}
	return nil
}

func (g *GoCoder) GetApiSavePath(absPath string) string {
	return filepath.Join(g.ApiPath, absPath)
}

type UserCallback struct {
	Method string      `json:"method"`
	Router string      `json:"router"`
	Type   string      `json:"type"`
	Body   interface{} `json:"body"`
}

func (g *GoCoder) UserCall(ctx context.Context, req *usercall.Request) (resp *usercall.Response, err error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	args := []string{
		g.Runner.GetBinPath() + "/" + g.Runner.GetBuildRunnerCurrentVersionName(),
		"usercall",
		"--method", req.Method,
		"--router", req.Router,
		"--type", req.Type,
		"--trace_id", ctx.Value(constants.TraceID).(string),
	}

	if req.Body == nil {
		args = append(args, "--file", "noBody")
	} else {
		err := jsonx.SaveFile(filepath.Join(g.Runner.GetRequestPath(), uuid.New().String()+".json"), req)
		if err != nil {
			return nil, err
		}
	}
	run, cmd, err := cmdx.Run(ctx, g.Runner.GetBinPath(), args)
	if err != nil {
		logger.Errorf(ctx, "exec err:%s: cmd:%s", string(run), strings.Join(args, " "))
		return nil, err
	}
	content := stringsx.ParserHtmlTagContent(string(run), "Response")
	if len(content) == 0 {
		return nil, fmt.Errorf(string(run))
	}
	defer cmd.Process.Kill()
	var res response.RunFunctionRespWithData[*usercall.Response]
	err = json.Unmarshal([]byte(content[0]), &res)
	if err != nil {
		return nil, err
	}
	if res.Code == 0 {
		return res.Data, nil
	}
	return nil, fmt.Errorf(res.Msg)

}

func (g *GoCoder) GetImportPath(addPkgPath string) string {
	if g.IsDev {
		return filepath.Join(fmt.Sprintf("github.com/yunhanshu-net/function-go/soft/%s/%s/code/api", g.User, g.Name), addPkgPath)
	}
	return filepath.Join(fmt.Sprintf("git.yunhanshu.net/%s/%s/api", g.User, g.Name), addPkgPath)
}

func (g *GoCoder) GenMainGo(packages []PackageInfo) error {

	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return err
	}

	os.RemoveAll(g.MainFile)
	f, err := os.Create(g.MainFile)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		Packages []PackageInfo
	}{
		Packages: packages,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}
	return nil
}

func (g *GoCoder) buildProject(ctx context.Context) (info *coder.ApiChangeInfo, err error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if !g.IsDev {
		if !osx.FileExists(filepath.Join(g.CodePath, "go.mod")) {
			return nil, fmt.Errorf("CodePath %s is not a Go module root", g.CodePath)
		}
	}
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = g.CodePath
	if output, err := tidy.CombinedOutput(); err != nil {
		logger.Errorf(ctx, "g:%+v buildProject tidy:%+v\n", g, string(output))
		return nil, fmt.Errorf("go mod tidy failed: %v\n%s", err, string(output))
	}
	oldVersion := g.Version
	version := g.GetNextVersion()
	// 定义可注入变量的结构体
	type BuildConfig struct {
		Version string
		User    string
		Name    string
		Root    string
		Output  string // 输出文件名
	}
	buildConfig := BuildConfig{
		Version: version,
		User:    g.User,
		Name:    g.Name,
		Output:  g.ReleasesPath + "/" + g.GetNextBuildName(),
	}
	ldflags := fmt.Sprintf(
		`-X 'github.com/yunhanshu-net/function-go/env.Version=%s' `+
			`-X 'github.com/yunhanshu-net/function-go/env.User=%s' `+
			`-X 'github.com/yunhanshu-net/function-go/env.Name=%s' `+
			`-X 'github.com/yunhanshu-net/function-go/env.Root=%s'`,
		buildConfig.Version,
		buildConfig.User,
		buildConfig.Name,
		buildConfig.Root,
	)

	// 构造命令
	cmd := exec.Command(
		"go",
		"build",
		"-ldflags",
		ldflags,
		"-o",
		buildConfig.Output,
	)
	cmd.Dir = g.MainPath
	if output, e := cmd.CombinedOutput(); e != nil {
		logger.Errorf(ctx, "buildProject go:%s\n", string(output))
		return nil, fmt.Errorf("go build failed: %v\n%s", e, string(output))
	}
	//todo git add and commit
	cmd = exec.Command("ln", "-s", "releases/"+g.GetNextBuildName(), g.GetNextBuildName())
	cmd.Dir = g.BinPath
	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Errorf(ctx, "ln go:%+v\n", string(output))
		return nil, fmt.Errorf("ln failed: %v\n%s", err, string(output))
	}

	//生成api log
	err = g.refreshApiLogs(ctx)
	if err != nil {
		return nil, err
	}
	err = g.refreshVersion(ctx)
	if err != nil {
		return nil, err
	}

	add, del, updated, err := g.DiffApi(ctx, oldVersion, version)
	if err != nil {
		return nil, err
	}
	rsp := &coder.ApiChangeInfo{
		CurrentVersion: version,
		AddApi:         add,
		DelApi:         del,
		UpdateApi:      updated,
	}
	logger.Infof(ctx, "buildProject success %s", g.GetCurrentBuildName())
	return rsp, nil

}

func (g *GoCoder) refreshApiLogs(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	version := g.GetNextVersion()
	//生成api log
	cmd := exec.Command("./"+g.GetNextBuildName(), "apis")
	cmd.Dir = g.BinPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("getApiInfos failed: %v\n%s", err, string(output))
	}
	content := stringsx.ParserHtmlTagContent(string(output), "Response")
	if content == nil || len(content) == 0 {
		return fmt.Errorf("getApiInfos failed: %v\n%s", err, string(output))
	}
	var res []*api.Info

	err = json.Unmarshal([]byte(content[0]), &res)
	if err != nil {
		return fmt.Errorf("getApiInfos failed: %v\n%s", err, string(output))
	}
	logs := api.ApiLogs{Version: version, Apis: res}
	err = jsonx.SaveFile(filepath.Join(g.ApiLogsPath, g.GetNextVersion()+".json"), logs)
	if err != nil {
		return fmt.Errorf("getApiInfos failed: %v\n%s", err, string(output))
	}
	return nil
}

func (g *GoCoder) refreshVersion(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	g.Runner.Version = g.GetNextVersion()
	versionPath := filepath.Join(g.MetaDataPath, "version.txt")
	err := osx.UpsertFile(versionPath, g.Runner.Version)
	if err != nil {
		return err
	}
	return nil
}

func (g *GoCoder) rollbackVersion(ctx context.Context, version string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	g.Runner.Version = g.GetNextVersion()
	versionPath := filepath.Join(g.MetaDataPath, "version.txt")
	err := osx.UpsertFile(versionPath, version)
	if err != nil {
		return err
	}
	return nil
}

func (g *GoCoder) CreateProject(ctx context.Context) (*coder.CreateProjectResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	logger.Infof(ctx, "CreateNode project:%+v", g.Runner)
	err := g.mkdirAll(ctx)
	if err != nil {
		return nil, err
	}
	err = g.GenMainGo(nil)
	if err != nil {
		return nil, err
	}
	err = g.refreshVersion(ctx)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "CreateNode project success:%s", g.GetCurrentBuildName())
	git, err := InitGit(g)
	if err != nil {
		return nil, err
	}
	err = git.AddAll()
	if err != nil {
		return nil, err
	}
	msg := GitCommitMsg{Version: "v0", Msg: "create project"}
	_, err = git.CommitAll(msg.JSON())
	if err != nil {
		return nil, err
	}
	return &coder.CreateProjectResp{Version: msg.Version}, nil
}

func (g *GoCoder) AddBizPackage(ctx context.Context, bizPackage *coder.BizPackage) (*coder.BizPackageResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	absPkgPath := bizPackage.GetPackageAbsPath(g.ApiPath)
	if osx.DirExists(absPkgPath) { //先判断Package是否存在
		return nil, status.ErrorCodeApiFileExist.WithMessage(absPkgPath)
	}
	//不存在才可以创建
	err := os.MkdirAll(absPkgPath, 0755)
	if err != nil {
		return nil, err
	}
	err = osx.UpsertFile(filepath.Join(absPkgPath, "keep_.go"), fmt.Sprintf("package %s", bizPackage.GetPackageName()))
	if err != nil {
		return nil, err
	}

	packageInfos := []PackageInfo{{ImportPath: g.GetImportPath(bizPackage.GetSubPackagePath())}}

	imports, err := ParseImports(g.MainFile)
	if err != nil {
		return nil, err
	}
	imports = append(imports, packageInfos...)
	imports = slicesx.Filter(imports, func(t PackageInfo) string {
		return t.ImportPath
	})

	err = g.GenMainGo(imports)
	if err != nil {
		return nil, err
	}
	msg := GitCommitMsg{Version: g.Version, Msg: bizPackage.Desc}
	git, err := InitGit(g)
	if err != nil {
		return nil, err
	}
	err = git.AddAll()
	if err != nil {
		return nil, err
	}
	hash, err := git.CommitAll(msg.JSON())
	if err != nil {
		return nil, err
	}
	return &coder.BizPackageResp{Version: g.Version, Hash: hash}, nil
}

func (g *GoCoder) addApi(ctx context.Context, api *coder.CodeApi) (path string, err error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	f := filepath.Join(g.ApiPath, api.GetSubPackagePath(), api.EnName+".go")
	err = osx.UpsertFile(f, api.Code)
	if err != nil {
		return f, err
	}
	return f, nil
}

// apiEqual 比较两个API是否相等
func apiEqual(a, b *api.Info) bool {
	// 使用reflect.DeepEqual进行深度比较
	return reflect.DeepEqual(a, b)
}

func (g *GoCoder) DiffApi(ctx context.Context, old string, new string) (add []*api.Info, del []*api.Info, updated []*api.Info, err error) {
	if ctx.Err() != nil {
		return nil, nil, nil, ctx.Err()
	}
	newApiInfos := &api.ApiLogs{}
	oldApiInfos := &api.ApiLogs{}
	old = filepath.Join(g.ApiLogsPath, old+".json")
	new = filepath.Join(g.ApiLogsPath, new+".json")
	err = jsonx.UnmarshalFromFile(old, oldApiInfos)
	if err != nil {
		logger.Errorf(ctx, "jsonx.UnmarshalFromFile(old, oldApiInfos) version:%s error: %v", old, err)
		err = nil
	}
	err = jsonx.UnmarshalFromFile(new, newApiInfos)
	if err != nil {
		return
	}
	// 创建旧API的映射，用于快速查找
	lastApiMap := make(map[string]*api.Info)
	for _, lastApi := range oldApiInfos.Apis {
		key := fmt.Sprintf("%s:%s", lastApi.Method, lastApi.Router)
		lastApiMap[key] = lastApi
	}

	// 遍历当前API，查找新增和更新的API
	for _, currentApi := range newApiInfos.Apis {
		key := fmt.Sprintf("%s:%s", currentApi.Method, currentApi.Router)
		// 检查API是否在上一版本中存在
		if lastApi, exists := lastApiMap[key]; exists {
			// 检查API是否有更新
			if !apiEqual(currentApi, lastApi) {
				updated = append(updated, currentApi)
			}
			// 标记已处理过的API
			delete(lastApiMap, key)
		} else {
			// 新增的API
			add = append(add, currentApi)
		}
	}
	// 剩余未处理的旧API即为已删除的API
	for _, api := range lastApiMap {
		del = append(del, api)
	}
	return
}

func (g *GoCoder) AddApis(ctx context.Context, req *coder.AddApisReq) (resp *coder.AddApisResp, err error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	resp = new(coder.AddApisResp)
	var addFiles []string
	for _, codeApi := range req.CodeApis {
		file, err := g.addApi(ctx, codeApi)
		if err != nil {
			return nil, err
		}
		addFiles = append(addFiles, file)
	}
	oldVersion := g.Version
	res, err := g.buildProject(ctx)
	if err != nil {
		for _, file := range addFiles {
			if file != "" {
				os.Remove(file)
			}
		}
		return nil, err
	}
	add, del, updated, err := g.DiffApi(ctx, oldVersion, res.CurrentVersion)
	if err != nil {
		return nil, err
	}
	resp.Version = res.CurrentVersion
	resp.ApiChangeInfo = &coder.ApiChangeInfo{
		CurrentVersion: res.CurrentVersion,
		AddApi:         add,
		DelApi:         del,
		UpdateApi:      updated,
	}
	//此时需要回调
	for _, info := range resp.ApiChangeInfo.AddApi {
		if info.CreateTables != nil {
			var req0 usercall.Request
			req0.Method = info.Method
			req0.Router = info.Router
			req0.Type = usercallConst.CallbackTypeOnCreateTables
			call, err1 := g.UserCall(ctx, &req0)
			if err1 != nil {
				logger.Errorf(ctx, "GoCoder.UserCall(%+v) CallbackTypeOnCreateTables err: %v", req, err1)
				continue
			}
			logger.Infof(ctx, "GoCoder.UserCall(%+v):CallbackTypeOnCreateTables success resp:%v", req, call)
		}

		if !slicesx.ContainsString(info.Callbacks, usercallConst.UserCallTypeOnApiCreated) {
			logger.Infof(ctx, "api no callback:%s ", usercallConst.UserCallTypeOnApiCreated)
			continue
		}
		var req1 usercall.Request
		req1.Method = info.Method
		req1.Router = info.Router
		req1.Type = usercallConst.UserCallTypeOnApiCreated
		call, err1 := g.UserCall(ctx, &req1)
		if err1 != nil {
			logger.Errorf(ctx, "GoCoder.UserCall(%+v) UserCallTypeOnApiCreated err: %v", req, err1)
			continue
		}
		logger.Infof(ctx, "GoCoder.UserCall(%+v): success UserCallTypeOnApiCreated resp:%v", req, call)
	}
	//此时发生了变更，需要重新编译，另外需要提交一下代码，保证可以及时回滚，
	msg := GitCommitMsg{Version: res.CurrentVersion, Msg: req.Msg}
	git, err := InitGit(g)
	if err != nil {
		return nil, err
	}
	err = git.AddAll()
	if err != nil {
		return nil, err
	}
	hash, err := git.CommitAll(msg.JSON())
	if err != nil {
		return nil, err
	}
	resp.Hash = hash
	return resp, nil
}

func (g *GoCoder) DeleteProject(ctx context.Context, req *coder.DeleteProjectReq) (*coder.DeleteProjectResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	go func() {
		err := os.RemoveAll(g.SavePath)
		if err != nil {
			logger.Errorf(ctx, "os.RemoveAll(g.SavePath) error: %v path:%s", err, g.SavePath)
		}
	}()
	return &coder.DeleteProjectResp{}, nil
}

func (g *GoCoder) DeleteApis(ctx context.Context, req *coder.DeleteAPIsReq) (*coder.DeleteAPIsResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	for _, codeApi := range req.CodeApis {
		path := filepath.Join(g.ApiPath, codeApi.GetSubPackagePath(), codeApi.EnName+".go")
		err := os.Remove(path)
		if err != nil {
			return nil, err
		}
	}

	info, err := g.buildProject(ctx)
	if err != nil {
		return nil, err
	}

	//此时发生了变更，需要重新编译，另外需要提交一下代码，保证可以及时回滚，
	msg := GitCommitMsg{Version: info.CurrentVersion, Msg: req.Msg}
	git, err := InitGit(g)
	if err != nil {
		return nil, err
	}
	err = git.AddAll()
	if err != nil {
		return nil, err
	}
	hash, err := git.CommitAll(msg.JSON())
	if err != nil {
		return nil, err
	}

	return &coder.DeleteAPIsResp{Hash: hash, DelApis: info.DelApi, Version: info.CurrentVersion}, nil
}
