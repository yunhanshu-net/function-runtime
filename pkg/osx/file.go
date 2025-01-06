package osx

import (
	"io"
	"os"
	"path/filepath"
)

// removeFile 删除文件
func removeFile(path string) error {
	return os.Remove(path)
}

// isDirEmpty 检查目录是否为空
func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// 读取目录中的文件
	_, err = f.Readdir(1)
	if err == io.EOF {
		// 目录为空
		return true, nil
	}
	return false, err
}

// removeEmptyParents 删除空的父目录
func removeEmptyParents(path string) error {
	parent := filepath.Dir(path)
	for parent != "/" && parent != "." {
		isEmpty, err := isDirEmpty(parent)
		if err != nil {
			return err
		}
		if isEmpty {
			if err := os.Remove(parent); err != nil {
				return err
			}
			parent = filepath.Dir(parent)
		} else {
			break
		}
	}
	return nil
}

// DeleteFileOrDir 删除文件或目录，并递归删除空的父目录
func DeleteFileOrDir(path string) error {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // 文件或目录不存在，不执行任何操作
	}

	// 检查是文件还是目录
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		// 如果是目录，先删除目录下的所有内容
		d, err := os.Open(path)
		if err != nil {
			return err
		}
		defer d.Close()

		objects, err := d.Readdir(-1)
		if err != nil {
			return err
		}

		for _, obj := range objects {
			err = DeleteFileOrDir(filepath.Join(path, obj.Name()))
			if err != nil {
				return err
			}
		}
	}

	// 删除文件或空目录
	err = removeFile(path)
	if err != nil {
		return err
	}

	// 删除空的父目录
	return removeEmptyParents(path)
}
