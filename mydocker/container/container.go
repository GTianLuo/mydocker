package container

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"my_docker/mydocker/common"
	"os"
	"text/tabwriter"
	"time"
)

type ContainerInfo struct {
	Pid        string `json:"pid"`        // 容器init进程在宿主机上的pid
	Id         string `json:"id"`         // 容器的唯一id
	Name       string `json:"name"`       // 容器名
	Command    string `json:"command"`    // init进程运行的命令
	CreateTime string `json:"createTime"` //容器的创建时间
	Status     string `json:"status"`     // 容器状态
}

var (
	Running             string = "running"
	STOP                string = "stopped"
	Exit                string = "exist"
	DefaultInfoLocation string = "/var/run/mydocker/%s/"
	ConfigName          string = "config.json"
)

func RecordContainerInfo(pid string, cmd string, containerName string) (string, error) {
	// 获取容器id
	id := common.GetRandomID()
	// 获取创建时间
	createTime := time.Now().Format("2006-01-02 15:04:05")
	if containerName == "" {
		// 未指定容器名字
		containerName = id
	}
	containerInfo := &ContainerInfo{
		Pid:        pid,
		Id:         id,
		Name:       containerName,
		Command:    cmd,
		CreateTime: createTime,
		Status:     Running,
	}
	// 序列化信息
	jsonBytes, err := json.Marshal(containerInfo)
	//容器信息的完整路径
	configUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err = common.MkdirIfNotExist(configUrl); err != nil {
		return "", err
	}
	fileName := configUrl + "config.json"
	file, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("Create file %s error:%s", fileName)
	}
	if _, err := file.Write(jsonBytes); err != nil {
		return "", fmt.Errorf("File write error:%v", err)
	}
	return containerName, nil
}

func DeleteContainerInfo(containerName string) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("delete container failed %v", err)
	}
}

func ListContainers() error {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, "")
	dirUrl = dirUrl[0 : len(dirUrl)-1]
	files, err := ioutil.ReadDir(dirUrl)
	if err != nil {
		return fmt.Errorf("Read dir %s error %v", dirUrl, err)
	}
	var containers []*ContainerInfo
	for _, file := range files {
		containerInfo, err := getContainerInfo(file)
		if err != nil {
			return err
		}
		containers = append(containers, containerInfo)
	}
	// tabwriter.NewWriter 在控制台打印容器信息,该库可以打印对齐表格
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreateTime)
	}
	// 刷新标准输出流缓存区，打印出信息
	if err := w.Flush(); err != nil {
		return fmt.Errorf("Flush error %v", err)
	}
	return nil
}

// 读取config file，并进行行序列化
func getContainerInfo(file os.FileInfo) (*ContainerInfo, error) {
	fileUrl := fmt.Sprintf(DefaultInfoLocation, file.Name()) + ConfigName
	jsonBytes, err := ioutil.ReadFile(fileUrl)
	if err != nil {
		return nil, err
	}
	containerInfo := &ContainerInfo{}
	if err := json.Unmarshal(jsonBytes, containerInfo); err != nil {
		return nil, fmt.Errorf("Json unmarshal error %v", err)
	}
	return containerInfo, nil
}
