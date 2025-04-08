package osx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	err = destination.Sync()
	if err != nil {
		return err
	}

	return nil
}

// CopyDirectory 复制目录，1 源目录存在的目录（文件），目标目录不存在，会copy文件到目标目录，2 目标目录存在，源目录不存在，会保留目标目录的文件
// srcDir 源目录
// dstDir 目标目录
func CopyDirectory(srcDir, dstDir string) error {
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			err = os.MkdirAll(dstPath, info.Mode())
			if err != nil && !os.IsExist(err) {
				return err
			}
		} else {
			err = copyFile(path, dstPath)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// FileExists 判断文件是否存在
func FileExists(filePath string) bool {
	// 使用 os.Stat 获取文件信息
	_, err := os.Stat(filePath)
	if err == nil {
		// 文件存在
		return true
	}
	// 检查错误是否是文件不存在的错误
	if os.IsNotExist(err) {
		// 文件不存在
		return false
	}
	// 其他错误（例如权限问题）
	fmt.Printf("Error checking file existence: %v\n", err)
	return false
}

// DirExists 判断目录是否存在
func DirExists(dirPath string) bool {
	// 使用 os.Stat 获取路径信息
	fileInfo, err := os.Stat(dirPath)
	if err == nil {
		// 如果路径存在且是目录
		return fileInfo.IsDir()
	}
	// 检查错误是否是路径不存在的错误
	if os.IsNotExist(err) {
		// 路径不存在
		return false
	}
	// 其他错误（例如权限问题）
	fmt.Printf("Error checking directory existence: %v\n", err)
	return false
}

// CheckDirectChildren 返回指定目录下的文件列表和目录列表
func CheckDirectChildren(baseDir string) (files []string, dirs []string, err error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, nil, err // 返回错误
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name()) // 添加到目录列表
		} else {
			files = append(files, entry.Name()) // 添加到文件列表
		}
	}

	return files, dirs, nil
}

// CountDirectories 统计指定路径下的目录数量
func CountDirectories(path string) (int, error) {
	entries, err := os.ReadDir(path) // 读取目录内容
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() { // 判断是否为目录
			count++
		}
	}
	return count, nil
}
