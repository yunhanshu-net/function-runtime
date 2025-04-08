package codex

import (
	"os"
	"text/template"
)

type PackageInfo struct {
	Alias      string // 包别名
	ImportPath string // 完整导入路径
}

const mainTemplate = `package main

import (
	"github.com/yunhanshu-net/sdk-go/runner"{{if .Packages}}
	{{- range .Packages}}
	{{.Alias}} "{{.ImportPath}}"{{end}}{{end}}
)

func InitPackages() {
	{{- range .Packages}}
	{{.Alias}}.Init(){{end}}
}

func main() {
	InitPackages()
	err := runner.Run()
	if err != nil {
		panic(err)
	}
}
`

func GenMainGo(packages []PackageInfo, filePath string) error {

	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		panic(err)
	}

	os.RemoveAll(filePath)
	f, err := os.Create(filePath)
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
