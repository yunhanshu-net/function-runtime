package model

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/conf"
	"github.com/yunhanshu-net/runcher/model/dto/api"
	"github.com/yunhanshu-net/runcher/pkg/jsonx"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func isVersion(v string) bool {
	if v == "" {
		return false
	}
	if v[0] != 'v' {
		return false
	}
	s := v[1:]
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return false
	}
	return i >= 0
}

func addVersion(v string, inc int64) (string, error) {
	if v == "" {
		return "", fmt.Errorf("v is empty")
	}
	if v[0] != 'v' {
		return "", fmt.Errorf("v 不符合规范")
	}
	s := v[1:]
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", fmt.Errorf("数字部分不符合规范")
	}
	r := i + inc
	if r >= 0 {
		return fmt.Sprintf("v%v", r), nil
	}
	return "", fmt.Errorf("inc 输入错误")
}

type Runner struct {
	Kind     string `json:"kind"`     //类型，可执行程序，so文件等等
	Language string `json:"language"` //编程语言
	Name     string `json:"name"`     //应用名称（英文标识）
	Version  string `json:"version"`  //应用版本
	User     string `json:"user"`     //所属租户
}

func (r *Runner) GetOldVersion() (*Runner, error) {
	version, err := addVersion(r.Version, -1)
	if err != nil {
		return nil, err
	}
	old := *r
	old.Version = version
	return &old, nil
}

func NewRunner(user string, name string, version ...string) (*Runner, error) {
	if user == "" {
		return nil, fmt.Errorf("user is empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	r := Runner{
		User: user,
		Name: name,
	}
	v := ""
	if len(version) > 0 {
		v = version[0]
		b := isVersion(v)
		if !b {
			return nil, fmt.Errorf("is failed version")
		}
	} else {
		vs, err := r.GetLatestVersion()
		if err != nil {
			return nil, err
		}
		r.Version = vs
	}
	return &r, nil
}

func (r *Runner) GetRequestSubject() string {
	builder := strings.Builder{}
	builder.WriteString("runner.")
	builder.WriteString(r.User)
	builder.WriteString(".")
	builder.WriteString(r.Name)
	builder.WriteString(".")
	builder.WriteString(r.Version)
	builder.WriteString(".run")
	return builder.String()
}

func (r *Runner) GetLatestVersion() (string, error) {
	versions, err := r.GetLatestVersions(1)
	if err != nil {
		return "", err
	}
	if len(versions) > 0 {
		return versions[0], nil
	}
	return "v0", nil
}

func (r *Runner) GetLatestVersionsDebug(count int) ([]string, error) {
	path := conf.GetRunnerRoot() + "/" + r.User + "/" + r.Name + "/debug"

	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("目录读取失败：%s", err.Error())
	}

	var versions []int

	// 构建前缀：user_name_v
	prefix := r.User + "_" + r.Name + "_v"

	for _, entry := range entries {
		// 只处理文件，忽略目录
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// 检查文件名是否以 user_name_v 开头
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		// 提取版本号部分
		numStr := name[len(prefix):]
		if numStr == "" {
			continue
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}

		versions = append(versions, num)
	}

	// 按版本号从大到小排序
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	// 截取指定数量
	if count < 0 {
		count = 0
	}
	if len(versions) > count {
		versions = versions[:count]
	}

	// 转换为字符串形式
	result := make([]string, len(versions))
	for i, v := range versions {
		result[i] = fmt.Sprintf("v%d", v)
	}

	return result, nil
}

// GetLatestVersions 返回指定目录下最新的若干个版本目录（如 v0, v1, v2...）
func (r *Runner) GetLatestVersions(count int) ([]string, error) {
	if conf.IsDev() {
		return r.GetLatestVersionsDebug(count)
	}
	path := conf.GetRunnerRoot() + "/" + r.User + "/" + r.Name + "/" + "version"
	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("目录读取失败：%s", err.Error())
	}

	// 收集所有合法的版本号
	var versions []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 2 || name[0] != 'v' {
			continue
		}
		numStr := name[1:]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		versions = append(versions, num)
	}

	// 按版本号从大到小排序
	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})

	// 截取指定数量
	if count < 0 {
		count = 0
	}
	if len(versions) > count {
		versions = versions[:count]
	}

	// 转换为字符串形式
	result := make([]string, len(versions))
	for i, v := range versions {
		result[i] = fmt.Sprintf("v%d", v)
	}

	return result, nil
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
	root := strings.TrimSuffix(rootPath, "/")
	if conf.IsDev() {
		return fmt.Sprintf("%s/%s/%s/debug", root, r.User, r.Name)
	}
	return fmt.Sprintf("%s/%s/%s/version/%s", root, r.User, r.Name, r.Version)
}

type RunnerPath struct {
	RootPath              string //根目录
	RunnerRoot            string
	CurrentVersionPath    string //当前版本目录
	NextVersionPath       string //下一个版本目录
	CurrentVersionBakPath string //当前版本备份目录
	CurrentVersionErrPath string //当前版本失败目录
	NextVersionBakPath    string //下一个版本备份目录
}

func (r *Runner) GetPaths(rootPath string) RunnerPath {
	return RunnerPath{
		RootPath:              rootPath,
		RunnerRoot:            fmt.Sprintf("%s/%s/%s", rootPath, r.User, r.Name),
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

func (r *Runner) GetApiPath() string {
	return fmt.Sprintf("%s/%s/%s/bin/api-logs", conf.GetRunnerRoot(), r.User, r.Name)
}

func (r *Runner) DiffApi(old string, new string) (add []*api.Info, del []*api.Info, updated []*api.Info, err error) {
	newApiInfos := &api.ApiLogs{}
	oldApiInfos := &api.ApiLogs{}
	if old != "" {
		err = jsonx.UnmarshalFromFile(old, oldApiInfos)
		if err != nil {
			return
		}
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

// apiEqual 比较两个API是否相等
func apiEqual(a, b *api.Info) bool {
	// 使用reflect.DeepEqual进行深度比较
	return reflect.DeepEqual(a, b)
}
