package container

import (
	"docker/common"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"syscall"
)

var (
	RootUrl     = "/var/lib/mydocker/overlay2"  // docker的root目录
	MntUrl      = RootUrl + "/mnt/%s"           // 容器的挂载目录
	WriteLayer  = RootUrl + "/writeLayer/%s"    // 容器可写层目录
	ImagesUrl   = "/var/lib/mydocker/images/%s" // 存放镜像tar压缩文件的目录
	ReadOnlyUrl = RootUrl + "/readOnly/%s"      // 只读成目录
	WorkUrl     = RootUrl + "/work/%s"          // 用于overlay fs 中转目录
)

// NewWorkSpace 创建容器运行的文件系统
func NewWorkSpace(volume []string, imageName, containerName string) error {
	if err := CreateReadOnlyLayer(imageName); err != nil {
		return err
	}
	if err := CreateWriteLayer(containerName); err != nil {
		return err
	}
	if err := CreateMountPoint(imageName, containerName); err != nil {
		return err
	}
	if err := MountVolume(containerName, volume); err != nil {
		return err
	}
	return nil
}

// CreateReadOnlyLayer 创建只读的镜像层,将基础镜像image.tar 解压到readOnly/image 目录下
func CreateReadOnlyLayer(imageName string) error {
	imageUrl := fmt.Sprintf(ImagesUrl, imageName) + ".tar"
	readOnlyUrl := fmt.Sprintf(ReadOnlyUrl, imageName)
	isExist, err := common.PathExist(ReadOnlyUrl)
	if err != nil {
		return fmt.Errorf("Fail to judge whether dir %s exists : %v", readOnlyUrl, err)
	}
	if !isExist {
		// 判断image的tar文件是否存在
		if !imageIsExist(imageName) {
			return fmt.Errorf("Unable to find image '%s' locally", imageName)
		}
		// image存在，创建readOnly目录
		if err := os.MkdirAll(readOnlyUrl, 0777); err != nil {
			return fmt.Errorf("create %s dir failed: %v", readOnlyUrl, err)
		}
		//并解压image.tar
		if _, err := common.Exec("tar", "-xvf", imageUrl, "-C", readOnlyUrl); err != nil {
			return fmt.Errorf("failed unzip busybox.tar: %v", err)
		}
	}
	return nil
}

// CreateWriteLayer 创建镜像的可写层
func CreateWriteLayer(containerName string) error {
	writeLayer := fmt.Sprintf(WriteLayer, containerName)
	if err := os.MkdirAll(writeLayer, 0777); err != nil {
		return fmt.Errorf("create writeLayer failed: %v", err)
	}
	return nil
}

// CreateMountPoint overlay文件系统的挂载，创建独立的root fs
func CreateMountPoint(imageName, containerName string) error {

	lowerDir := fmt.Sprintf(ReadOnlyUrl, imageName)
	workDir := fmt.Sprintf(WorkUrl, containerName)
	upperDir := fmt.Sprintf(WriteLayer, containerName)
	mntDir := fmt.Sprintf(MntUrl, containerName)
	// 创建挂载目录
	if err := os.MkdirAll(mntDir, 0777); err != nil {
		return fmt.Errorf("create mnt dir failed: %v", err)
	}
	// 创建overlay的work目录
	if err := os.MkdirAll(workDir, 0777); err != nil {
		return fmt.Errorf("create mnt dir failed: %v", err)
	}
	// 挂载
	if bytes, err := common.Exec("mount", "-t", "overlay", "overlay", "-o", "lowerdir="+lowerDir+",upperdir="+upperDir+",workdir="+workDir, mntDir); err != nil {
		return fmt.Errorf("overlay filesystem mount filed: %v:%v ", err, string(bytes))
	}
	return nil
}

// MountVolume 进行数据卷的挂载
func MountVolume(containerName string, volume []string) error {

	mntDir := fmt.Sprintf(MntUrl, containerName)
	// 判断参数的合法性
	if len(volume) == 0 {
		// 用户未设置数据卷挂载
		return nil
	}
	i := 0
	for i < len(volume) {
		if err := mountVolume(mntDir, volume[i], volume[i+1]); err != nil {
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
func DeleteWorkSpace(imageName string, containerName string, volume []string) {
	if err := DeleteVolumeMount(imageName, volume); err != nil {
		log.Error(err)
	}
	DeleteMountPoint(containerName)
	DeleteWriteLayer(containerName)
}

// DeleteVolumeMount 删除数据卷挂载
func DeleteVolumeMount(containerName string, volume []string) error {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	i := 0
	for i < len(volume) {
		if bytes, err := common.Exec("umount", path.Join(mntUrl, volume[i+1])); err != nil {
			return fmt.Errorf("delete volume mount failed:%v", string(bytes))
		}
		i = i + 2
	}
	return nil
}

// DeleteMountPoint 解除文件系统的挂载
func DeleteMountPoint(containerName string) {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	// 取消挂载
	if _, err := common.Exec("umount", mntUrl); err != nil {
		log.Errorf("failed umount overlay point:%v ", err)
		return
	}
	// 删除目录
	if err := os.Remove(mntUrl); err != nil {
		log.Errorf("failed delete %v:%v", mntUrl, err)
	}
}

// DeleteWriteLayer 删除可写层
func DeleteWriteLayer(containerName string) {
	writeLayerDir := fmt.Sprintf(WriteLayer, containerName)
	workDir := fmt.Sprintf(WorkUrl, containerName)
	// 删除目录
	if err := os.RemoveAll(writeLayerDir); err != nil {
		log.Errorf("failed delete %v:%v", writeLayerDir, err)
	}
	// 删除目录
	if err := os.RemoveAll(workDir); err != nil {
		log.Errorf("failed delete %v:%v", workDir, err)
	}
}

// 判断容器是否存在
func imageIsExist(imageName string) bool {
	imageUrl := fmt.Sprintf(ImagesUrl, imageName) + ".tar"
	exist, _ := common.PathExist(imageUrl)
	return exist
}
