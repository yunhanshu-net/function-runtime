package coder

import (
	"bufio"
	"fmt"
	"github.com/yunhanshu-net/runcher/pkg/slicesx"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type GolangProjectManager struct {
	Path string
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

func NewGolangProjectManager(path string) *GolangProjectManager {
	return &GolangProjectManager{path}
}

func (g *Golang) GetMainFile() string {
	return strings.TrimSuffix(g.getCurrentVersionPath(), "/") + "/main.go"
}

func (g *Golang) getMainTemplate() string {
	return fmt.Sprintf(`package main

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
}`)
}

func (g *Golang) CreateMain(addPackages []PackageInfo) error {
	return GenMainGo(addPackages, g.GetMainFile())
}

type PackageInfo struct {
	Alias      string // 包别名
	ImportPath string // 完整导入路径
}

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

func (g *Golang) AddPackages(addPackages []PackageInfo) error {
	imports, err := g.parseImports(g.GetMainFile())
	if err != nil {
		return err
	}
	imports = append(imports, addPackages...)
	imports = slicesx.Filter(imports, func(t PackageInfo) string {
		return t.ImportPath
	})
	return g.CreateMain(imports)
}

// 从文件中提取import信息（排除runner）
func (g *Golang) parseImports(filename string) ([]PackageInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//results := make(map[string]string)
	var packages []PackageInfo

	inImportBlock := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		switch {
		case strings.HasPrefix(line, "import ("):
			inImportBlock = true

			continue
		case strings.HasPrefix(line, "import "):
			// 处理单行import
			parts := strings.SplitN(line[len("import "):], " ", 2)
			packages = g.processImportLine(parts)
			inImportBlock = true
		case inImportBlock && line == ")":
			inImportBlock = false

		case inImportBlock && line != "":
			// 处理分组import中的行
			packages = append(packages, g.processImportLine([]string{line})...)
		}
	}

	packages = slicesx.Filter(packages, func(p PackageInfo) string {
		return p.ImportPath
	})
	return packages, scanner.Err()
}

// 处理单行import内容
func (g *Golang) processImportLine(parts []string) (packages []PackageInfo) {
	// 使用正则表达式解析
	re := regexp.MustCompile(`^(\w+)?\s*(".*")$`)
	matches := re.FindStringSubmatch(strings.Join(parts, " "))

	if len(matches) == 3 {
		alias := strings.TrimSpace(matches[1])
		path := strings.Trim(matches[2], `"`)

		// 排除runner包
		if path == "github.com/yunhanshu-net/sdk-go/runner" {
			return
		}

		// 如果没有别名，使用路径最后部分
		if alias == "" {
			pathParts := strings.Split(path, "/")
			alias = pathParts[len(pathParts)-1]
		}

		//results[alias] = path
		packages = append(packages, PackageInfo{
			Alias:      alias,
			ImportPath: path,
		})
	}
	return packages
}
