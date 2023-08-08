package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/common"
	"my_docker/mydocker/container"
	"os"
)

// CommitContainer 将容器提交为一个image
func CommitContainer(containerName, imageName string) {
	mntUrl := fmt.Sprintf(container.MntUrl, containerName)
	// 判断容器是否存在
	if exist, _ := common.PathExist(mntUrl); !exist {
		fmt.Fprintf(os.Stdout, "No such container:%s\n", containerName)
		return
	}
	imagesUrl := fmt.Sprintf(container.ImagesUrl, imageName) + ".tar"
	if errorBytes, err := common.Exec("tar", "-czf", imagesUrl, "-C", mntUrl, "."); err != nil {
		log.Errorf("Tar folder %s error %v", mntUrl, string(errorBytes))
		return
	}
}
