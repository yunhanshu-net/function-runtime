package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/compress"
	"github.com/yunhanshu-net/runcher/pkg/osx"
	"github.com/yunhanshu-net/runcher/pkg/slicesx"
	"github.com/yunhanshu-net/runcher/pkg/store"
	"github.com/yunhanshu-net/runcher/pkg/stringsx"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func NewCmd(runner *model.Runner) *Cmd {
	//dir, _ := os.UserHomeDir()
	dir := "./soft_cmd"
	fullName := runner.AppCode

	//这里应该判断本机系统类型
	//if runner.ToolType == "windows" {
	//	fullName += ".exe"
	//}
	if runtime.GOOS == "windows" {
		fullName += ".exe"
	}

	return &Cmd{
		InstallInfo: response.InstallInfo{
			TempPath:     filepath.Join(os.TempDir(), runner.ToolType),
			RootPath:     dir,
			Name:         runner.AppCode,
			FullName:     fullName,
			User:         runner.TenantUser,
			Version:      runner.Version,
			DownloadPath: runner.OssPath,
		},
	}
}

func (c *Cmd) GetAppName() string {
	return c.FullName
}

type Cmd struct {
	response.InstallInfo
	NatsClient *nats.Conn
}

func (c *Cmd) RollbackVersion(r *request.RollbackVersion, fileStore store.FileStore) (*response.RollbackVersion, error) {
	_, err := c.UpdateVersion(&model.UpdateVersion{RunnerConf: r.RunnerConf, OldVersion: r.OldVersion}, fileStore)
	if err != nil {
		return nil, err
	}
	return &response.RollbackVersion{}, nil
}

// DeCompressPath 解压临时目录
func (c *Cmd) DeCompressPath() string {
	return filepath.Join(c.TempPath, c.User, c.Name)
}

// GetInstallPath  安装目录
func (c *Cmd) GetInstallPath() string {
	abs, err := filepath.Abs(fmt.Sprintf("%s/%s/%s", c.RootPath, c.User, c.Name))
	if err != nil {
		panic(err)
		return abs
	}
	return abs
}

// Chmod mac和linux需要授予执行权限
func (c *Cmd) Chmod() error {
	if runtime.GOOS != "windows" {
		cmdPath := filepath.Join(c.GetInstallPath(), c.Name)
		cmd := exec.Command("chmod", "+x", cmdPath)
		// 执行命令
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Chmod cmd.Run() failed with path:%s err:%s\n", cmdPath, err)
		}
	}
	return nil
}

func (c *Cmd) Install(fileStore store.FileStore) (*response.InstallInfo, error) {
	//absPath := c.RootPath

	file, err := fileStore.GetFile(c.DownloadPath)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(file.FileLocalPath)

	DeCompressPath := c.DeCompressPath()
	defer os.RemoveAll(DeCompressPath)
	err = compress.DeCompress(file.FileLocalPath, DeCompressPath)
	if err != nil {
		return nil, err
	}
	files, dirs, err := osx.CheckDirectChildren(DeCompressPath)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 && len(dirs) == 0 {
		return nil, fmt.Errorf("cmd Install no files found in %s", DeCompressPath)
	}
	dirs = slicesx.Select(dirs, func(t string) bool {
		return !strings.HasPrefix(t, "_")
	})
	compressHome := ""
	if len(files) == 0 && len(dirs) == 1 {
		if dirs[0] == c.Name { //说明解压时候解压路径多出一级和应用名称相同的目录
			compressHome = filepath.Join(DeCompressPath, c.Name)
		}
	} else {
		compressHome = DeCompressPath
	}
	InstallPath := c.GetInstallPath()
	err = osx.CopyDirectory(compressHome, InstallPath)
	if err != nil {
		return nil, err
	}
	//c.InstallInfo.InstallPath = InstallPath

	err = c.Chmod()
	if err != nil {
		return nil, err
	}
	return &c.InstallInfo, nil

}

func (c *Cmd) UnInstall() (*response.UnInstallInfo, error) {
	return nil, nil
}

func (c *Cmd) UpdateVersion(up *model.UpdateVersion, fileStore store.FileStore) (*response.UpdateVersion, error) {
	//ps: OssPath=  tool/beiluo/1725442391820/helloworld.zip
	src := filepath.Join(c.RootPath, c.User)
	appDirName := c.Name                                               //目录名称：helloworld
	backPath := filepath.Join(src, ".back", appDirName, up.OldVersion) // ps: ./soft_cmd/beiluo/.back/helloworld/v1.0
	currentSoftSrc := c.GetInstallPath()                               //ps: ./soft_cmd/beiluo/helloworld
	fileInfo, err := fileStore.GetFile(up.RunnerConf.Version)
	if err != nil {
		return nil, err
	}
	defer fileInfo.RemoveLocalFile()

	err = copy.Copy(currentSoftSrc, backPath) //更换版本前把旧版本备份一份，存档
	if err != nil {
		return nil, err
	}
	copyTempDir := currentSoftSrc + "/.temp/" + appDirName
	defer os.RemoveAll(copyTempDir)
	err = compress.DeCompress(fileInfo.FileLocalPath, copyTempDir)
	if err != nil {
		return nil, err
	}
Back:
	count := 0

	if count >= 3 {
		return nil, fmt.Errorf("cmd UpdateVersion path faild")
	}
	files, dirs, err := osx.CheckDirectChildren(copyTempDir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 && len(dirs) == 0 {
		return nil, fmt.Errorf("cmd UpdateVersion no files found in %s", copyTempDir)
	}
	if len(files) == 0 && len(dirs) == 1 {
		if dirs[0] == c.Name {
			copyTempDir = filepath.Join(copyTempDir, c.Name)
			count++
			goto Back
		}
	}
	err = osx.CopyDirectory(copyTempDir, currentSoftSrc) //复制目录
	if err != nil {
		return nil, err
	}
	err = c.Chmod()
	if err != nil {
		return nil, err
	}
	//复制目录
	return &response.UpdateVersion{}, nil
}

func (c *Cmd) Request(req *request.Request) (*response.Response, error) {
	var (
		cmdStr    string
		err       error
		outString string
		res       response.Response
	)
	defer func() {
		if err != nil {
			logrus.Errorf("Cmd call err:%s exec:%s ", err, cmdStr)
		} else {
			logrus.Infof("Cmd call exec:%s ", cmdStr)
		}
	}()
	now := time.Now()
	installPath := c.GetInstallPath()
	appName := c.GetAppName()
	softPath := fmt.Sprintf("%s/%s", installPath, appName)
	softPath = strings.ReplaceAll(softPath, "\\", "/")
	req.SoftInfo.RequestJsonPath = strings.ReplaceAll(req.SoftInfo.RequestJsonPath, "\\", "/")
	var cmd *exec.Cmd
	split := strings.Split(req.SoftInfo.RequestJsonPath, "/")
	reqName := split[len(split)-1]
	switch runtime.GOOS {
	case "windows":
		cc := fmt.Sprintf("cd /D %s && %s %s .request/%s",
			installPath, appName, req.SoftInfo.Command, reqName)
		cmd = exec.Command("cmd.exe", "/C", cc)
	case "linux", "darwin":
		//cc := fmt.Sprintf("cd  %s && %s %s %s",
		//	installPath, softPath, req.Command, req.RequestJsonPath)

		cc := fmt.Sprintf("cd %s && ./%s %s .request/%s",
			installPath, appName, req.SoftInfo.Command, reqName)
		// Linux和macOS可以直接使用 && 连接命令
		cmd = exec.Command("sh", "-c", cc)
	default:
		cmd = exec.Command(softPath, req.SoftInfo.Command, req.SoftInfo.RequestJsonPath)
		fmt.Printf("cmd call err:%s exec:%s ", err, cmdStr)
	}

	//cmd := exec.Command(softPath, req.Command, req.RequestJsonPath)
	//cmd.Dir = installPath
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		logrus.Errorf("cmd run err:%s", err.Error())
		return nil, err
	}
	cmdStr = fmt.Sprintf("%s %s %s", softPath, req.SoftInfo.Command, req.SoftInfo.RequestJsonPath)
	logrus.Infof("Cmd run %s", cmdStr)
	outString = out.String()
	if outString == "" {
		//todo
		return nil, fmt.Errorf("out.String() ==== nil cmd程序输出的结果为空，请检测程序是否正确")
	}
	mem := stringsx.ParserHtmlTagContent(outString, "mem_use_B")
	if len(mem) <= 0 {
		//todo 请使用sdk开发软件
		return nil, fmt.Errorf("soft call err 未获取到内存占用信息，请使用sdk开发软件")
	}
	i, err := strconv.ParseInt(mem[0], 10, 64)
	if err != nil {
		return nil, err
	}

	resList := stringsx.ParserHtmlTagContent(outString, "Response")
	if len(resList) == 0 {
		//todo 请使用sdk开发软件
		return nil, fmt.Errorf("soft call err 请使用sdk开发软件")
	}
	err = json.Unmarshal([]byte(resList[0]), &res)
	if err != nil {
		return nil, err
	}
	since := time.Since(now)
	res.MetaData["cost"] = since
	res.MetaData["mem_b"] = int(i)
	err = json.Unmarshal([]byte(outString), &res.SoftResponse)
	if err != nil {
		return nil, err
	}
	//p.printSoftLogs(s, since)
	return &res, nil
}
