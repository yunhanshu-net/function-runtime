package coder

import (
	"bufio"
	"github.com/yunhanshu-net/runcher/pkg/slicesx"
	"os"
	"regexp"
	"strings"
)

// ParseImports 从文件中提取import信息（排除runner）
func ParseImports(filename string) ([]PackageInfo, error) {
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
			packages = processImportLine(parts)
			inImportBlock = true
		case inImportBlock && line == ")":
			inImportBlock = false

		case inImportBlock && line != "":
			// 处理分组import中的行
			packages = append(packages, processImportLine([]string{line})...)
		}
	}

	packages = slicesx.Filter(packages, func(p PackageInfo) string {
		return p.ImportPath
	})
	return packages, scanner.Err()
}

// 处理单行import内容
func processImportLine(parts []string) (packages []PackageInfo) {
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
