package main

import (
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/container"
	"os"
)

// Run 启动容器
func Run(isTty bool, isInteractive bool, command string) {
	cmd := container.NewParentProcess(isTty, isInteractive, command)
	if err := cmd.Start(); err != nil {
		log.Error(err)
	}
	cmd.Wait()
	os.Exit(-1)
}
