package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"my_docker/mydocker/common"
	"my_docker/mydocker/container"
)

func getContainerPidByName(containerName string) (string, error) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	containerConfigPath := fmt.Sprintf(container.DefaultInfoLocation, containerName) + container.ConfigFileName
	if exist, _ := common.PathExist(containerConfigPath); !exist {
		return nil, fmt.Errorf(" No such container: %s", containerName)
	}
	contentBytes, err := ioutil.ReadFile(containerConfigPath)
	if err != nil {
		return nil, fmt.Errorf("Read config file error:%v", err)
	}
	containerInfo := &container.ContainerInfo{}
	_ = json.Unmarshal(contentBytes, containerInfo)
	return containerInfo, nil
}

func writeContainerInfo(info *container.ContainerInfo) error {
	// 序列化
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("Marshal containerInfo failed error:%v", err)
	}
	configFileUrl := fmt.Sprintf(container.DefaultInfoLocation, info.Name) + container.ConfigFileName
	if err := ioutil.WriteFile(configFileUrl, infoBytes, 0622); err != nil {
		return fmt.Errorf("Write file %s error", configFileUrl)
	}
	return nil
}
