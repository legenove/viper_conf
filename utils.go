package viper_conf

import (
	"bytes"
	"os"
)

func createDir(path string) error {
	if !pathExists(path) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func concatenateStrings(s ...string) string {
	if len(s) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	for _, i := range s {
		buffer.WriteString(i)
	}
	return buffer.String()
}
