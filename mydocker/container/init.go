package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"my_docker/mydocker/common/pipe"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {

	// 从pipe文件中读取command
	readPipe := os.NewFile(uintptr(3), "pipe")
	cmdBytes, err := pipe.ReadPipe(readPipe)
	if err != nil {
		return err
	}
	_ = pipe.ClosePipe(readPipe)
	cmd := string(cmdBytes)

	setUpMount()

	// 在环境变量中找到可执行文件的完整路径
	file := strings.Split(cmd, " ")[0]
	path, err := exec.LookPath(file)
	if err != nil {
		return fmt.Errorf("not find %s in PATH:%s", file, err.Error())
	}
	argv := []string{cmd}
	if err := syscall.Exec(path, argv, os.Environ()); err != nil {
		return err
	}
	return nil
}

func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v", err)
		return
	}
	log.Infof("Current location is %s", pwd)
	if err := pivotRoot(pwd); err != nil {
		log.Infof("pivot_root fail: %v", err)
		return
	}
	/*
			/proc是虚拟的文件系统，并不是真实存在的。所以为了防止程序的恶意攻击，需要为该文件系统添加一些标志位
			 1. MS_NOEXEC 标志位通常用于挂载不可执行文件的文件系统，以防止恶意程序在其中执行任何代码。
		     2. MS_NOSUID 标志位通常用于挂载不含SUID或SGID程序的文件系统，以防止恶意程序以特权用户身份运行。
		     3. MS_NODEV 标志位通常用于挂载不含设备文件的文件系统，以防止恶意程序访问设备文件并进行攻击。
	*/
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("Failed to mount /proc filesystem:" + err.Error())
		return
	}
	if err := syscall.Mount(
		"tmpfs",
		"/dev",
		"tmpfs",
		syscall.MS_NOSUID|syscall.MS_STRICTATIME,
		"mode=755",
	); err != nil {
		log.Errorf("Failed to mount /dev filesystem:", err.Error())
	}
}

func pivotRoot(newroot string) error {

	// 将namespace下的所有挂载点改为私有挂载点
	if err := syscall.Mount(
		"",
		"/",
		"",
		syscall.MS_PRIVATE|syscall.MS_REC,
		"",
	); err != nil {
		return fmt.Errorf("mount / private failed: %v", err)
	}
	putold := filepath.Join(newroot, "/.pivot_root")
	// 创建 rootfs/.pivot_root 存储old_root
	if err := os.Mkdir(putold, 0777); err != nil {
		return err
	}
	// 将newroot变为挂载点
	if err := syscall.Mount(
		newroot,
		newroot,
		"",
		syscall.MS_SLAVE|syscall.MS_BIND|syscall.MS_REC,
		"",
	); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	// pivot_root
	if err := syscall.PivotRoot(newroot, putold); err != nil {
		return fmt.Errorf("pivot_root :%v", err)
	}

	// 修改当前的工作目录
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}
	putold = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(putold, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount pivot_root dir %v", err)
	}
	return os.Remove(putold)
}
