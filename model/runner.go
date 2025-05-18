package model

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/pkg/osx"
	"path/filepath"
	"strconv"
	"strings"
)

type Runner struct {
	Kind     string `json:"kind"`     //类型，可执行程序，so文件等等
	Language string `json:"language"` //编程语言
	Name     string `json:"name"`     //应用名称（英文标识）
	Version  string `json:"version"`  //应用版本
	User     string `json:"user"`     //所属租户
	root     string //
}

func NewRunner(user string, name string, root string, version ...string) (*Runner, error) {
	if user == "" {
		return nil, fmt.Errorf("user is empty")
	}
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	r := Runner{
		User: user,
		Name: name,
		root: root,
	}

	v := ""
	if len(version) > 0 {
		v = version[0]
		b := isVersion(v)
		r.Version = v
		if !b {
			return nil, fmt.Errorf("is failed version")
		}
	} else {
		vs, err := r.GetCurrentVersion()
		if err != nil {
			return nil, err
		}
		r.Version = vs
	}
	return &r, nil
}
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

func (r *Runner) GetCurrentVersion() (string, error) {
	v := osx.ReadToString(filepath.Join(r.root, r.User, r.Name, "workplace", "metadata", "version.txt"))
	if v != "" {
		return v, nil
	}
	return "v0", nil
}

func (r *Runner) GetBinPath() string {
	return fmt.Sprintf("%s/%s/%s/workplace/bin", r.root, r.User, r.Name)
}
func (r *Runner) GetRequestPath() string {
	return fmt.Sprintf("%s/.request", r.GetBinPath())
}

func (r *Runner) GetBuildRunnerCurrentVersionName() string {
	return fmt.Sprintf("%s_%s_%s", r.User, r.Name, r.Version)
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
