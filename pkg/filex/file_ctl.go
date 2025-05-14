package filex

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func FileCopy(src, dist string) error {
	os.Remove(dist)
	dir := filepath.Dir(dist)
	os.MkdirAll(dir, os.ModePerm)
	input, err := os.Open(src) // 要复制的源文件
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dist) // 复制到的目标文件
	if err != nil {
		return err
	}
	defer output.Close()

	// 复制文件内容
	_, err = io.Copy(output, input)
	if err != nil {
		return err
	}
	return nil
}

func GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hashType := "sha256"

	var hashValue []byte
	switch hashType {
	case "md5":
		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}
		hashValue = hash.Sum(nil)
	case "sha1":
		hash := sha1.New()
		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}
		hashValue = hash.Sum(nil)
	case "sha256":
		hash := sha256.New()
		if _, err := io.Copy(hash, file); err != nil {
			return "", err
		}
		hashValue = hash.Sum(nil)
	default:
		return "", fmt.Errorf("unsupported hash type")
	}
	return fmt.Sprintf("%x", hashValue), nil
}

// MustCreateFileAndWriteContent 创建文件并写入内容，如果出错则panic
func MustCreateFileAndWriteContent(path string, content string) {
	err := CreateFileAndWriteContent(path, content)
	if err != nil {
		panic(fmt.Sprintf("创建文件失败: %v", err))
	}
}

// CreateFileAndWriteContent 创建文件并写入内容
func CreateFileAndWriteContent(path string, content string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入内容
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入内容失败: %v", err)
	}

	return nil
}

// MustReadFile 读取文件内容，如果出错则panic
func MustReadFile(path string) string {
	content, err := ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("读取文件失败: %v", err))
	}
	return content
}

// ReadFile 读取文件内容
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}
	return string(content), nil
}

// MustReadFileLines 读取文件的所有行，如果出错则panic
func MustReadFileLines(path string) []string {
	lines, err := ReadFileLines(path)
	if err != nil {
		panic(fmt.Sprintf("读取文件行失败: %v", err))
	}
	return lines
}

// ReadFileLines 读取文件的所有行
func ReadFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件行失败: %v", err)
	}

	return lines, nil
}

// MustWriteJSON 将对象写入JSON文件，如果出错则panic
func MustWriteJSON(path string, v interface{}) {
	err := WriteJSON(path, v)
	if err != nil {
		panic(fmt.Sprintf("写入JSON文件失败: %v", err))
	}
}

// WriteJSON 将对象写入JSON文件
func WriteJSON(path string, v interface{}) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 创建JSON编码器
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	// 编码并写入
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("编码JSON失败: %v", err)
	}

	return nil
}

// MustReadJSON 从JSON文件读取对象，如果出错则panic
func MustReadJSON(path string, v interface{}) {
	err := ReadJSON(path, v)
	if err != nil {
		panic(fmt.Sprintf("读取JSON文件失败: %v", err))
	}
}

// ReadJSON 从JSON文件读取对象
func ReadJSON(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("解码JSON失败: %v", err)
	}

	return nil
}

// MustCopyFile 复制文件，如果出错则panic
func MustCopyFile(src, dst string) {
	err := CopyFile(src, dst)
	if err != nil {
		panic(fmt.Sprintf("复制文件失败: %v", err))
	}
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	// 打开源文件
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer source.Close()

	// 确保目标目录存在
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 创建目标文件
	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destination.Close()

	// 复制内容
	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	return nil
}

// MustMoveFile 移动文件，如果出错则panic
func MustMoveFile(src, dst string) {
	err := MoveFile(src, dst)
	if err != nil {
		panic(fmt.Sprintf("移动文件失败: %v", err))
	}
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	// 确保目标目录存在
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 尝试直接重命名
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// 如果重命名失败，尝试复制后删除
	if err := CopyFile(src, dst); err != nil {
		return fmt.Errorf("复制文件失败: %v", err)
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("删除源文件失败: %v", err)
	}

	return nil
}

// MustDeleteFile 删除文件，如果出错则panic
func MustDeleteFile(path string) {
	err := DeleteFile(path)
	if err != nil {
		panic(fmt.Sprintf("删除文件失败: %v", err))
	}
}

// DeleteFile 删除文件
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	return nil
}

// MustExists 检查文件是否存在，如果出错则panic
func MustExists(path string) bool {
	exists, err := Exists(path)
	if err != nil {
		panic(fmt.Sprintf("检查文件是否存在失败: %v", err))
	}
	return exists
}

// Exists 检查文件是否存在
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("检查文件是否存在失败: %v", err)
}

// MustIsDir 检查路径是否为目录，如果出错则panic
func MustIsDir(path string) bool {
	isDir, err := IsDir(path)
	if err != nil {
		panic(fmt.Sprintf("检查是否为目录失败: %v", err))
	}
	return isDir
}

// IsDir 检查路径是否为目录
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("获取文件信息失败: %v", err)
	}
	return info.IsDir(), nil
}
