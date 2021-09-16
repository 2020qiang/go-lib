package file

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

// IsExist 判断所给路径文件或目录是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return IsExist(path) && s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return IsExist(path) && !IsDir(path)
}

// Sha256sum 等同于命令sha256sum <file>
func Sha256sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// FindTypeF 等同于命令 find <path> -type f，第二参数是返回数据是否拼接为绝对路径
func FindTypeF(path string, outAbs bool) ([]string, error) {

	var pwd string
	if outAbs {
		var err error
		pwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	var files []string
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if outAbs && !filepath.IsAbs(path) {
				files = append(files, filepath.Join(pwd, path))
			} else {
				files = append(files, path)
			}
			return nil
		})
	return files, err
}
