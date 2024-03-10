package utils

import (
	"os"
)

func PathInfo(path string) (bool, os.FileInfo, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil, err
	}

	return true, info, nil
}

func IsFile(path string) bool {
	exists, info, _ := PathInfo(path)
	if !exists {
		return false
	}

	return !info.IsDir()
}

func IsDirectory(path string) bool {
	exists, info, _ := PathInfo(path)
	if !exists {
		return false
	}

	return info.IsDir()
}
