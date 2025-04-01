package codex

import (
	"github.com/yunhanshu-net/runcher/pkg/slicesx"
	"strings"
)

type GolangProjectManager struct {
	Path string
}

func NewGolangProjectManager(path string) *GolangProjectManager {
	return &GolangProjectManager{path}
}

func (g *GolangProjectManager) GetMainFile() string {
	return strings.TrimSuffix(g.Path, "/") + "/main.go"
}

func (g *GolangProjectManager) CreateMain(addPackages []PackageInfo) error {
	return GenMainGo(addPackages, g.GetMainFile())
}

func (g *GolangProjectManager) AddPackages(addPackages []PackageInfo) error {
	imports, err := parseImports(g.GetMainFile())
	if err != nil {
		return err
	}
	imports = append(imports, addPackages...)
	imports = slicesx.Filter(imports, func(t PackageInfo) string {
		return t.ImportPath
	})
	return g.CreateMain(imports)
}
