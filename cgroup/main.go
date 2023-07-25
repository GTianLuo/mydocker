package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

func main() {

	// 创建的容器cgroup所在的层级结构
	const cgroupMount = "/sys/fs/cgroup/mydocker"
	//
	if os.Args[0] == "/proc/self/exe" {
		//容器进程
		fmt.Println("current pid : ", syscall.Getpid())
		//使用stress 命令模拟内存的负载
		cmd := exec.Command("sh", "-c", `stress --vm-bytes 200m --vm-keep -m 1`)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	// 改命令会动态链接到该程序自己，自己调用自己，然后程序会进入上面的if
	cmd := exec.Command("/proc/self/exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	} else {
		// 这里的pid启动的容器进程对应的外部空间的pid
		fmt.Printf("%v\n", cmd.Process.Pid)
		// 在指定的cgroup层级上创建层级
		os.Mkdir(path.Join(cgroupMount, "testmemorylimit"), 0755)
		//将容器进程加入这个层级
		ioutil.WriteFile(path.Join(cgroupMount, "testmemorylimit", "cgroup.procs"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
		// 限制cgroup进程的使用
		ioutil.WriteFile(path.Join(cgroupMount, "testmemorylimit", "memory.max"), []byte("300M"), 0644)
	}
	cmd.Process.Wait()
}
