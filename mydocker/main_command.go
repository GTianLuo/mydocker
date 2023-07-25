package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"my_docker/mydocker/container"
)

func init() {
	app.AddCommand(
		runCommand,
		initCommand,
	)
	// 添加 -i 和 -t 参数
	runCommand.Flags().BoolP("interactive", "i", false, interactiveUsage)
	runCommand.Flags().BoolP("tty", "t", false, ttyUsage)
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
		isTty, err := cmd.Flags().GetBool("tty")
		if err != nil {
			log.Error("invalid flag:", err.Error())
		}
		isInteractive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			log.Error("invalid flag:", err.Error())
		}
		Run(isTty, isInteractive, args[0])
		return nil
	},
}

// 该命令用于程序内部的调用，是由新创建的容器进程调用，初始化新容器
var initCommand = &cobra.Command{
	Use:   "init",
	Short: initCommandShort,
	Long:  initCommandLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]
		log.Infof("command %s", command)
		return container.RunContainerInitProcess(command, nil)
	},
}
