package command

import (
	"docker/network"
	log "github.com/sirupsen/logrus"
)

func NetworkCreate(driver, subnet, name string) {
	// 初始化网络模块
	if err := network.Init(); err != nil {
		log.Error("network create failed:", err.Error())
		return
	}
	if err := network.CreateNetwork(driver, subnet, name); err != nil {
		log.Error("network create failed:", err.Error())
		return
	}
}
