package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"my_docker/mydocker/cgroups/subsystems"
	"my_docker/mydocker/command"
	"my_docker/mydocker/common"
	"my_docker/mydocker/container"
	"os"
	"strings"
)

func init() {

	app.AddCommand(
		runCommand,
		initCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		rmCommand,
		commitCommand,
	)
	// 添加 -i 和 -t 参数
	runCommand.Flags().BoolP("interactive", "i", false, interactiveUsage)
	runCommand.Flags().BoolP("tty", "t", false, ttyUsage)
	//资源限制参数 -m
	runCommand.Flags().StringP("memory", "m", "max", memoryUsage)
	//数据卷映射参数 -v
	runCommand.Flags().StringSliceP("volume", "v", []string{}, volumeUsage)
	//--name 参数
	runCommand.Flags().StringP("name", "n", "", nameUsage)
	// -d 参数
	runCommand.Flags().BoolP("detach", "d", false, detachUsage)
	// -e 参数
	runCommand.Flags().StringSliceP("env", "e", []string{}, envUsage)
}

var runCommand = &cobra.Command{
	Use:   "run",
	Short: runCommandShort,
	Long:  runCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		//判断参数是否包含command并获取
		if len(args) < 2 {
			return fmt.Errorf("Missing container command")
		}
		// 获取容器名字
		imageName := args[0]
		command := strings.Join(args[1:], " ")
		// 读取开关参数
		isTty, _ := cmd.Flags().GetBool("tty")
		isInteractive, _ := cmd.Flags().GetBool("interactive")

		// 读取资源限制参数
		memoryLimit, _ := cmd.Flags().GetString("memory")

		// 获取数据卷映射参数
		volume, _ := cmd.Flags().GetStringSlice("volume")
		volumeParam, err := common.ParseVolumeParam(volume)
		if err != nil {
			return err
		}
		//获取name参数
		name, _ := cmd.Flags().GetString("name")
		// 获取detach
		detach, err := cmd.Flags().GetBool("detach")
		if detach && isTty || detach && isInteractive {
			return fmt.Errorf("ti and paramter can not both provided")
		}
		// 读取环境变量
		envs, _ := cmd.Flags().GetStringSlice("env")

		Run(isTty, isInteractive, detach, command, &subsystems.ResourceConfig{MemoryLimit: memoryLimit}, name, volumeParam, imageName, envs)
		return nil
	},
}

// 该命令用于程序内部的调用，是由新创建的容器进程调用，初始化新容器
var initCommand = &cobra.Command{
	Use:   "init",
	Short: initCommandShort,
	Long:  initCommandLong,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("command %s", args[0])
		container.RunContainerInitProcess()
	},
}

// 列出正在运行的容器
var listCommand = &cobra.Command{
	Use:   "ps",
	Short: psCommandShort,
	Long:  psCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		return container.ListContainers()
	},
}

// 查看容器日志
var logCommand = &cobra.Command{
	Use:   "logs",
	Short: logsCommandShort,
	Long:  logsCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取参数并校验
		if len(args) != 1 {
			return fmt.Errorf("\"docker logs\" requires exactly 1 argument")
		}
		containerName := args[0]
		container.LogContainerLog(containerName)
		return nil
	},
}

var execCommand = &cobra.Command{
	Use:   "exec [flags] CONTAINER COMMAND [ARG...]",
	Short: execCommandShort,
	Long:  execCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv(command.ENV_EXEC_PID) != "" {
			// callback
			log.Infof("pid callback pid: %d", os.Getpid())
			return nil
		}

		// 校验参数
		if len(args) < 2 {
			return fmt.Errorf("Messing container name or command")
		}
		containerName := args[0]
		cmdS := strings.Join(args[1:], " ")
		command.ExecContainer(containerName, cmdS)
		return nil
	},
}

var stopCommand = &cobra.Command{
	Use:   "stop",
	Short: stopCommandShort,
	Long:  stopCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Messing container name")
		}
		containerName := args[0]
		command.StopContainer(containerName)
		return nil
	},
}

var rmCommand = &cobra.Command{
	Use:   "rm",
	Short: rmCommandShort,
	Long:  rmCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Messing container name")
		}
		containerName := args[0]
		command.RmContainer(containerName)
		return nil
	},
}

var commitCommand = &cobra.Command{
	Use:   "commit",
	Short: commitShort,
	Long:  commitLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("Messing container name or image name")
		}
		command.CommitContainer(args[0], args[1])
		return nil
	},
}
