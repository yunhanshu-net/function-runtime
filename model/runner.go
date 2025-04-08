package model

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/pkg/osx"
	"strconv"
	"strings"
)

type Runner struct {
	Kind     string `json:"kind"`     //类型，可执行程序，so文件等等
	Language string `json:"language"` //编程语言
	Name     string `json:"name"`     //应用名称（英文标识）
	Version  string `json:"version"`  //应用版本
	User     string `json:"user"`     //所属租户
	//WorkPath        string `json:"work_path"`
	//Command         string `json:"command"`
	//RequestJsonPath string `json:"request_json_path"`

	//StoreRoot string `json:"store_root"` //oss 存储的跟路径

	//ToolType  string `json:"tool_type"`  //工具类型

	//OssPath   string `json:"oss_path"`   //文件地址

}

func (r *Runner) GetLatestVersion() (string, error) {
	directories, err := osx.CountDirectories(conf.GetRunnerRoot() + "/" + r.User + "/" + r.Name + "/" + "version")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("v%v", directories-1), nil
}
func (r *Runner) GetUnixFileName() string {
	return fmt.Sprintf("%s_%s_%s.sock", r.User, r.Name, r.Version)
}

func (r *Runner) GetUnixPathFile() string {
	return fmt.Sprintf("%s/%s/%s/bin/%s", conf.GetRunnerRoot(), r.User, r.Name, r.GetUnixFileName())
}
func (r *Runner) GetBinPath() string {
	return fmt.Sprintf("%s/%s/%s/bin", conf.GetRunnerRoot(), r.User, r.Name)
}
func (r *Runner) GetRequestPath() string {
	return fmt.Sprintf("%s/.request", r.GetBinPath())
}

func (r *Runner) GetBuildRunnerName() string {
	return fmt.Sprintf("%s_%s_%s", r.User, r.Name, r.GetNextVersion())
}

func (r *Runner) GetBuildRunnerCurrentVersionName() string {
	return fmt.Sprintf("%s_%s_%s", r.User, r.Name, r.Version)
}
func (r *Runner) GetBuildRunnerNextVersionName() string {
	return fmt.Sprintf("%s_%s_%s", r.User, r.Name, r.GetNextVersion())
}
func (r *Runner) GetBuildPath(root string) string {
	return fmt.Sprintf("%s/%s/%s/bin", root, r.User, r.Name)
}

func (r *Runner) GetVersionNum() (int, error) {
	replace := strings.ReplaceAll(r.Version, "v", "")
	version, err := strconv.Atoi(replace)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return version, nil
}

func (r *Runner) GetNextVersion() string {
	num, err := r.GetVersionNum()
	if err != nil {
		fmt.Println("GetVersionNum err:" + err.Error())
	}
	return fmt.Sprintf("v%d", num+1)
}

func (r *Runner) GetInstallPath(rootPath string) string {
	return fmt.Sprintf("%s/%s/%s/version/%s", strings.TrimSuffix(rootPath, "/"), r.User, r.Name, r.Version)
}

type Path struct {
	RootPath              string //根目录
	CurrentVersionPath    string //当前版本目录
	NextVersionPath       string //下一个版本目录
	CurrentVersionBakPath string //当前版本备份目录
	CurrentVersionErrPath string //当前版本失败目录
	NextVersionBakPath    string //下一个版本备份目录
}

func (r *Runner) GetPaths(rootPath string) Path {
	return Path{
		RootPath:              rootPath,
		CurrentVersionPath:    fmt.Sprintf("%s/%s/%s/version/%s", strings.TrimSuffix(rootPath, "/"), r.User, r.Name, r.Version),
		NextVersionPath:       fmt.Sprintf("%s/%s/%s/version/%s", strings.TrimSuffix(rootPath, "/"), r.User, r.Name, r.GetNextVersion()),
		CurrentVersionErrPath: fmt.Sprintf("%s/%s/%s/version/%s_err", strings.TrimSuffix(rootPath, "/"), r.User, r.Name, r.Version),
		CurrentVersionBakPath: fmt.Sprintf("%s/%s/%s/version/%s_bak", strings.TrimSuffix(rootPath, "/"), r.User, r.Name, r.Version),
	}
}

func (r *Runner) GetToolPath(rootPath string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(rootPath, "/"), r.User, r.Name)
}

func (r *Runner) GetNextVersionInstallPath(rootPath string) (string, error) {
	nextVersion := r.GetNextVersion()
	return fmt.Sprintf("%s/%s/%s/version/%s", strings.TrimSuffix(rootPath, "/"), r.User, r.Name, nextVersion), nil
}

func (r *Runner) Check() error {

	return nil
}

type UpdateVersion struct {
	RunnerConf *Runner `json:"runner_conf"`
	OldVersion string  `json:"old_version"`
	//NewVersion        string  `json:"new_version"`
	//NewVersionOssPath string  `json:"new_version_oss_path"`
}
