package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var app = &cobra.Command{
	Use:   "mydocker",
	Short: appShort,
	Long:  appLong,
}

func main() {
	app.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// mydocker 运行配置初始化
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}
	app.AddCommand(runCommand)
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}

}
