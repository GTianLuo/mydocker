package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"my_docker/mydocker/cgroups/subsystems"
	"my_docker/mydocker/container"
	"strings"
)

func init() {
	app.AddCommand(
		runCommand,
		initCommand,
	)
	// 添加 -i 和 -t 参数
	runCommand.Flags().BoolP("interactive", "i", false, interactiveUsage)
	runCommand.Flags().BoolP("tty", "t", false, ttyUsage)
	runCommand.Flags().StringP("memory", "m", "-1m", memoryUsage)
}

var runCommand = &cobra.Command{
	Use:   "run",
	Short: runCommandShort,
	Long:  runCommandLong,

	RunE: func(cmd *cobra.Command, args []string) error {
		//判断参数是否包含command并获取
		if len(args) < 1 {
			return fmt.Errorf("Missing container command")
		}
		// 读取开关参数
		isTty, _ := cmd.Flags().GetBool("tty")
		isInteractive, _ := cmd.Flags().GetBool("interactive")

		// 读取资源限制参数
		memoryLimit, _ := cmd.Flags().GetString("memory")
		command := strings.Join(args, " ")
		Run(isTty, isInteractive, command, &subsystems.ResourceConfig{MemoryLimit: memoryLimit})
		return nil
	},
}

// 该命令用于程序内部的调用，是由新创建的容器进程调用，初始化新容器
var initCommand = &cobra.Command{
	Use:   "init",
	Short: initCommandShort,
	Long:  initCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Infof("command %s", args[0])
		return container.RunContainerInitProcess(args[0], nil)
	},
}
