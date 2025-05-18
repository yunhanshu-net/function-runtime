package coder

const mainTemplate = `package main

import (
	"github.com/yunhanshu-net/sdk-go/runner"{{if .Packages}}
	{{- range .Packages}}
	_ "{{.ImportPath}}"{{end}}{{end}}
)


func main() {
	err := runner.Run()
	if err != nil {
		panic(err)
	}
}
`

type PackageInfo struct {
	Alias      string // 包别名
	ImportPath string // 完整导入路径
}
