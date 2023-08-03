package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/container"
	"os"
)

func RmContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Error(err)
		return
	}
	if containerInfo.Status != container.STOP {
		log.Errorf("You cannot remove a running container %v", containerName)
		return
	}
	if err := os.RemoveAll(fmt.Sprintf(container.DefaultInfoLocation, containerName)); err != nil {
		log.Errorf("Remove container info dir error:%v", err)
		return
	}
	fmt.Fprintf(os.Stdout, containerName+"\n")
}
