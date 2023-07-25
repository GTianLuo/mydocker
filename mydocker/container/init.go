package container

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 获取创建新进程的命令
// 该命令在执行时调用当前的可执行程序,这里通过参数设置调用init方法
func NewParentProcess(tty bool, interactive bool, command string) *exec.Cmd {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	}
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uint32(1),
		Gid: uint32(1),
	}
	if tty && interactive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

func RunContainerInitProcess(command string, args []string) error {
	/*
			/proc是虚拟的文件系统，并不是真实存在的。所以为了防止程序的恶意攻击，需要为该文件系统添加一些标志位
			 1. MS_NOEXEC 标志位通常用于挂载不可执行文件的文件系统，以防止恶意程序在其中执行任何代码。
		     2. MS_NOSUID 标志位通常用于挂载不含SUID或SGID程序的文件系统，以防止恶意程序以特权用户身份运行。
		     3. MS_NODEV 标志位通常用于挂载不含设备文件的文件系统，以防止恶意程序访问设备文件并进行攻击。
	*/
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		log.Errorf("Failed to mount /proc filesystem:" + err.Error())
		return err
	}
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}