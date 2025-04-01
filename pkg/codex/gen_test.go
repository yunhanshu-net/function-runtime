package codex

import (
	"testing"
)

func TestGEN(t *testing.T) {

	manager := NewGolangProjectManager("/Users/beiluo/Documents/code/runcher/pkg/codex/test")
	//err := manager.CreateMain([]PackageInfo{
	//	{Alias: "excelx1", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx1"},
	//	{Alias: "excelx2", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx2"},
	//	{Alias: "excelx3", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx3"},
	//})

	err := manager.CreateMain(nil)
	if err != nil {
		panic(err)
	}

}

func TestAdd(t *testing.T) {
	manager := NewGolangProjectManager("/Users/beiluo/Documents/code/runcher/pkg/codex/test")
	err := manager.AddPackages([]PackageInfo{
		{Alias: "excelx4", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx4"},
		{Alias: "excelx5", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx5"},
		{Alias: "excelx6", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx6"},
		{Alias: "excelx7", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx7"},
		{Alias: "excelx8", ImportPath: "git.yunhanshu.net/beiluo/api/api/excelx8"},
	})
	if err != nil {
		panic(err)
	}
}
