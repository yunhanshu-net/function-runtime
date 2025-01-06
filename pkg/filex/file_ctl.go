package filex

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
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
