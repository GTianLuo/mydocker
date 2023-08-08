package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	_ "my_docker/mydocker/nsentry"
	"os"
	"os/exec"
	"strings"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func ExecContainer(containerName, command string) {
	// 查找pid
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, err.Error())
		return
	}
	log.Infof("container pid %s ", pid)
	log.Infof("command %s ", command)

	//构建命令
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 设置环境变量
	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, command)
	env := getProcessEnv(pid)
	cmd.Env = append(os.Environ(), env...)
	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerName, err)
	}
}

// exec开启的进程会继承父进程的环境变量，
// 这里的父进程也就是主机进程，并非后台运行的
// 容器进程，所以这里需要从/proc/pid/environ 中读取容器进程的env
func getProcessEnv(pid string) []string {
	environFileUrl := fmt.Sprintf("/proc/%s/environ", pid)
	envBytes, err := ioutil.ReadFile(environFileUrl)
	if err != nil {
		log.Errorf("Read file %s error %v", environFileUrl, err)
	}
	env := strings.Split(string(envBytes), "\u0000")
	return env
}
