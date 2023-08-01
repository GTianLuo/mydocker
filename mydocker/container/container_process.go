package container

import (
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/common"
	"my_docker/mydocker/common/pipe"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// NewParentProcess 获取创建新进程的命令
// 该命令在执行时调用当前的可执行程序,这里通过参数设置调用init方法
func NewParentProcess(tty bool, interactive bool, command string) (*exec.Cmd, *os.File) {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	// 创建一个pipe用来传递command
	readPipe, writePipe, err := pipe.NewPipe()
	if err != nil {
		return nil, nil
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// TODO user namespace
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	}
	rootUrl := "/home/gtl/docker"
	mntUrl := "/home/gtl/docker/mnt"
	NewWorkSpace(rootUrl, mntUrl)
	cmd.Dir = mntUrl
	if tty && interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	return cmd, writePipe
}

// NewWorkSpace 创建容器运行的文件系统
func NewWorkSpace(rootUrl string, mntUrl string) {
	CreateReadOnlyLayer(rootUrl, path.Join(rootUrl, "busybox.tar"))
	CreateWriteLayer(rootUrl)
	CreateMountPoint(rootUrl, mntUrl)
}

// CreateReadOnlyLayer 创建只读的镜像层,将基础镜像busybox.tar 解压到busybox 目录下
func CreateReadOnlyLayer(rootUrl string, busyboxTarUrl string) {
	busyboxUrl := path.Join(rootUrl, "busybox")
	isExist, err := common.PathExist(busyboxUrl)
	if err != nil {
		log.Errorf("create read only layer failed: %v", err)
		return
	}
	if !isExist {
		// 不存在，创建busybox目录
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			log.Errorf("create busybox dir failed: %v", err)
			return
		}
		//并解压busybox.tar
		if _, err := common.Exec("tar", "-xvf", busyboxTarUrl, "-C", busyboxUrl); err != nil {
			log.Errorf("failed unzip busybox.tar: %v", err)
			return
		}
	}
}

// CreateWriteLayer 创建镜像的可写层
func CreateWriteLayer(rootUrl string) {
	writeLayer := path.Join(rootUrl, "writeLayer")
	if err := os.Mkdir(writeLayer, 0777); err != nil {
		log.Errorf("create writeLayer failed: %v", err)
		return
	}
}

// CreateMountPoint overlay文件系统的挂载
func CreateMountPoint(rootUrl string, mntUrl string) {

	lowerDir := path.Join(rootUrl, "busybox")
	workDir := path.Join(rootUrl, "work")
	upperDir := path.Join(rootUrl, "writeLayer")
	os.Chdir("/home/gtl/docker")
	//bytes, _ := common.Exec("ls")
	//fmt.Println(string(bytes))
	// 创建挂载目录
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		log.Errorf("create mnt dir failed: %v", err)
		return
	}
	// 创建overlay的work目录
	if err := os.Mkdir(workDir, 0777); err != nil {
		log.Errorf("create mnt dir failed: %v", err)
		return
	}
	// 挂载
	if bytes, err := common.Exec("mount", "-t", "overlay", "overlay", "-o", "lowerdir="+lowerDir+",upperdir="+upperDir+",workdir="+workDir, mntUrl); err != nil {
		log.Errorf(string(bytes))
		log.Errorf("overlay filesystem mount filed: %v", err)
		return
	}
}

// DeleteWorkSpace 删除容器的工作空间，删除可写层，取消挂载，删除work目录
func DeleteWorkSpace(rootUrl string, mntUrl string) {
	DeleteMountPoint(mntUrl)
	DeleteWriteLayer(rootUrl)
}

// DeleteMountPoint 解除文件系统的挂载
func DeleteMountPoint(mntUrl string) {
	// 取消挂载
	if _, err := common.Exec("umount", mntUrl); err != nil {
		log.Errorf("failed delete overlay mount point:%v ", err)
		return
	}
	// 删除目录
	if err := os.Remove(mntUrl); err != nil {
		log.Errorf("failed delete %v:%v", mntUrl, err)
	}
}

// DeleteWriteLayer 删除可写层
func DeleteWriteLayer(rootUrl string) {
	writeLayerDir := path.Join(rootUrl, "writeLayer")
	workDir := path.Join(rootUrl, "work")
	// 删除目录
	if err := os.RemoveAll(writeLayerDir); err != nil {
		log.Errorf("failed delete %v:%v", writeLayerDir, err)
	}
	// 删除目录
	if err := os.RemoveAll(workDir); err != nil {
		log.Errorf("failed delete %v:%v", workDir, err)
	}
}
