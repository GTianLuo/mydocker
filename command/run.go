package command

import (
	"docker/cgroups"
	"docker/cgroups/subsystems"
	"docker/common"
	"docker/common/pipe"
	"docker/container"
	"docker/network"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

// Run 启动容器
func Run(isTty bool, isInteractive bool, detach bool, command string, res *subsystems.ResourceConfig, containerName string, volume []string, imageName string, env []string, nw string) {

	// 获取容器id
	cid := common.GetRandomID()
	if containerName == "" {
		// 未指定容器名字
		containerName = cid
	}
	// 获取容器创建初始化的command
	cmd, writePipe, err := container.NewParentProcess(isTty, isInteractive, detach, containerName, command, volume, imageName, env)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, err.Error())
		return
	}
	// 启动容器进程
	if err := cmd.Start(); err != nil {
		_, _ = fmt.Fprintf(os.Stdout, err.Error())
		return
	}
	//defer func() {
	//	if !detach {
	//		container.DeleteWorkSpace(imageName, containerName, volume)
	//	}
	//}()

	info, err := container.RecordContainerInfo(cid, strconv.Itoa(cmd.Process.Pid), command, containerName, volume)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, err.Error())
		return
	}
	// 使用container-pid作为cgroup名字
	cgroupManager := cgroups.NewCgroupManager("container-"+strconv.Itoa(os.Getpid()), res)
	//defer func() {
	//	if !detach {
	//		if err := cgroupManager.Destroy(); err != nil {
	//			_, _ = fmt.Fprintf(os.Stdout, err.Error())
	//		}
	//	}
	//}()

	if err := cgroupManager.Apply(cmd.Process.Pid); err != nil {
		_, _ = fmt.Fprintf(os.Stdout, err.Error())
		return
	}
	if err := cgroupManager.Set(); err != nil {
		_, _ = fmt.Fprintf(os.Stdout, err.Error())
		return
	}
	// 传递参数
	if err := pipe.WritePipe(writePipe, command); err != nil {
		log.Error(err)
	}
	_ = pipe.ClosePipe(writePipe)

	// 加载网络配置
	if err := network.Init(); err != nil {
		log.Error(err)
	}
	// 创建网络端点
	if err := network.Connect(nw, info); err != nil {
		log.Error(err)
	}

	if isTty && isInteractive {
		_ = cmd.Wait()
		//container.DeleteContainerInfo(info.Name)
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, info.Name)
	return
}
