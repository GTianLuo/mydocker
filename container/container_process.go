package container

import (
	"docker/common"
	"docker/common/pipe"
	"fmt"

	"os"
	"os/exec"
	"syscall"
)

// NewParentProcess 获取创建新进程的命令
// 该命令在执行时调用当前的可执行程序,这里通过参数设置调用init方法
func NewParentProcess(isTty, isInteractive, detach bool, containerName string, command string, volume []string, imageName string, env []string) (*exec.Cmd, *os.File, error) {
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...)
	if isTty && isInteractive {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else if detach {
		if err := redirectLog(cmd, containerName); err != nil {
			return nil, nil, fmt.Errorf("log redirect failed:%v", err)
		}
	}
	// 创建一个pipe用来传递command
	readPipe, writePipe, err := pipe.NewPipe()
	if err != nil {
		return nil, nil, err
	}
	// 设置cmd启动进程的属性
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// TODO user namespace
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
	}
	if err := NewWorkSpace(volume, imageName, containerName); err != nil {
		return nil, nil, err
	}
	// 设置起始目录
	cmd.Dir = fmt.Sprintf(MntUrl, containerName)
	// 设置进程额外打开的文件描述符
	cmd.ExtraFiles = []*os.File{readPipe}
	//设置进程的环境变量
	cmd.Env = append(os.Environ(), env...)
	return cmd, writePipe, nil
}

// RedirectLog 重定向日志
func redirectLog(cmd *exec.Cmd, containerName string) error {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := common.MkdirIfNotExist(dirUrl); err != nil {
		return fmt.Errorf("Create dir error:%v", err)
	}
	// 创建日志文件
	file, err := os.Create(dirUrl + LogFileName)
	if err != nil {
		return fmt.Errorf("Create log file error:%v", file)
	}
	cmd.Stdout = file
	return nil
}
