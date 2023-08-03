package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	_ "my_docker/mydocker/nsentry"
	"os"
	"os/exec"
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

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, command)
	cmd.Dir = "/home/gtl/test-overlayfs"
	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerName, err)
	}
}
