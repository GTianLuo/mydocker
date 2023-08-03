package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/container"
	"os"
	"strconv"
	"syscall"
)

func StopContainer(containerName string) {
	// 获取容器信息
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Error(err)
		return
	}
	if containerInfo.Status == container.STOP {
		// 已经关闭
		return
	}

	// 通过信号关闭进程
	pidInt, _ := strconv.Atoi(containerInfo.Pid)
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error %v", containerName, err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	// 写回文件
	if err := writeContainerInfo(containerInfo); err != nil {
		log.Error(err)
		return
	}
	fmt.Fprintln(os.Stdout, containerInfo.Name+"\n")
}
