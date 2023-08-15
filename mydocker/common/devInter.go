package common

import (
	"net"
	"strings"
)

// DevInterIsExist 判断网络设备接口是否存在
func DevInterIsExist(devName string) (bool, error) {
	_, err := net.InterfaceByName(devName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return false, err
	}
	return true, nil
}
