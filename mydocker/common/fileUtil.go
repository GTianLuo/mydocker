package common

import (
	"fmt"
	"os"
)

func PathExist(url string) (bool, error) {
	_, err := os.Stat(url)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func MkdirIfNotExist(url string) error {
	// 判断Path是否存在，不存在创建
	exist, err := PathExist(url)
	if err != nil {
		return fmt.Errorf("MkdirIfNotExist failed:%v", err)
	}
	if !exist {
		if err := os.MkdirAll(url, 0777); err != nil {
			return fmt.Errorf("MkdirIfNotExist failed:%v", err)
		}
	}
	return nil
}
