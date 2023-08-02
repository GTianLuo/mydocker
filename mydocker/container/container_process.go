package container

import (
	"fmt"
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
func NewParentProcess(tty bool, interactive bool, command string, volume []string) (*exec.Cmd, *os.File, error) {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	// 创建一个pipe用来传递command
	readPipe, writePipe, err := pipe.NewPipe()
	if err != nil {
		return nil, nil, err
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// TODO user namespace
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	}
	rootUrl := "/home/gtl/docker"
	mntUrl := "/home/gtl/docker/mnt"
	if err := NewWorkSpace(rootUrl, mntUrl, volume); err != nil {
		return nil, nil, err
	}
	cmd.Dir = mntUrl
	if tty && interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	return cmd, writePipe, nil
}

// NewWorkSpace 创建容器运行的文件系统
func NewWorkSpace(rootUrl string, mntUrl string, volume []string) error {
	if err := CreateReadOnlyLayer(rootUrl, path.Join(rootUrl, "busybox.tar")); err != nil {
		return err
	}
	if err := CreateWriteLayer(rootUrl); err != nil {
		return err
	}
	if err := CreateMountPoint(rootUrl, mntUrl); err != nil {
		return err
	}
	if err := MountVolume(mntUrl, volume); err != nil {
		return err
	}
	return nil
}

// CreateReadOnlyLayer 创建只读的镜像层,将基础镜像busybox.tar 解压到busybox 目录下
func CreateReadOnlyLayer(rootUrl string, busyboxTarUrl string) error {
	busyboxUrl := path.Join(rootUrl, "busybox")
	isExist, err := common.PathExist(busyboxUrl)
	if err != nil {
		return fmt.Errorf("create read only layer failed: %v", err)
	}
	if !isExist {
		// 不存在，创建busybox目录
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			return fmt.Errorf("create busybox dir failed: %v", err)
		}
		//并解压busybox.tar
		if _, err := common.Exec("tar", "-xvf", busyboxTarUrl, "-C", busyboxUrl); err != nil {
			return fmt.Errorf("failed unzip busybox.tar: %v", err)
		}
	}
	return nil
}

// CreateWriteLayer 创建镜像的可写层
func CreateWriteLayer(rootUrl string) error {
	writeLayer := path.Join(rootUrl, "writeLayer")
	if err := os.Mkdir(writeLayer, 0777); err != nil {
		return fmt.Errorf("create writeLayer failed: %v", err)
	}
	return nil
}

// CreateMountPoint overlay文件系统的挂载
func CreateMountPoint(rootUrl string, mntUrl string) error {

	lowerDir := path.Join(rootUrl, "busybox")
	workDir := path.Join(rootUrl, "work")
	upperDir := path.Join(rootUrl, "writeLayer")
	// 创建挂载目录
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		return fmt.Errorf("create mnt dir failed: %v", err)
	}
	// 创建overlay的work目录
	if err := os.Mkdir(workDir, 0777); err != nil {
		return fmt.Errorf("create mnt dir failed: %v", err)
	}
	// 挂载
	if bytes, err := common.Exec("mount", "-t", "overlay", "overlay", "-o", "lowerdir="+lowerDir+",upperdir="+upperDir+",workdir="+workDir, mntUrl); err != nil {
		return fmt.Errorf("overlay filesystem mount filed: %v:%v ", err, string(bytes))
	}
	return nil
}

// MountVolume 进行数据卷的挂载
func MountVolume(mntUrl string, volume []string) error {

	//if len(volume) == 0 {
	//	fmt.Println("未设置数据卷挂载")
	//} else {
	//	fmt.Println(len(volume), " ", volume)
	//}
	//return fmt.Errorf("error")
	// 判断参数的合法性
	if len(volume) == 0 {
		// 用户未设置数据卷挂载
		return nil
	}
	i := 0
	for i < len(volume) {
		if err := mountVolume(mntUrl, volume[i], volume[i+1]); err != nil {
			return err
		}
		i = i + 2
	}
	return nil
}

func mountVolume(mntUrl string, srcPath string, destPath string) error {

	// 判断srcPath 是否存在，不存在创建
	if err := common.MkdirIfNotExist(srcPath); err != nil {
		return fmt.Errorf("mount volume failed:%v", err)
	}
	// 判断destUrl是否存在，不存在创建
	if err := common.MkdirIfNotExist(path.Join(mntUrl, destPath)); err != nil {
		return fmt.Errorf("mount volume failed:%v", err)
	}
	// bind mount
	if err := syscall.Mount(srcPath, path.Join(mntUrl, destPath), "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount volume failed:%v", err)
	}
	return nil
}

// DeleteWorkSpace 删除容器的工作空间，删除可写层，取消挂载，删除work目录
func DeleteWorkSpace(rootUrl string, mntUrl string, volume []string) {
	if err := DeleteVolumeMount(mntUrl, volume); err != nil {
		log.Error(err)
	}
	DeleteMountPoint(mntUrl)
	DeleteWriteLayer(rootUrl)
}

// DeleteVolumeMount 删除数据卷挂载
func DeleteVolumeMount(mntUtl string, volume []string) error {
	i := 0
	for i < len(volume) {
		if bytes, err := common.Exec("umount", path.Join(mntUtl, volume[i+1])); err != nil {
			return fmt.Errorf("delete volume mount failed:%v", string(bytes))
		}
		i = i + 2
	}
	return nil
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
