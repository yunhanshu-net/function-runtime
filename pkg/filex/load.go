package filex

import (
	"fmt"
	"os"
)

func LoadStringFromFile(path string) string {
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("filex LoadStringFromFile err:%s\n", err)
		return ""
	}
	return string(file)
}
