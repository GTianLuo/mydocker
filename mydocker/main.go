package main

import (
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
		return initDocker()
	}
	app.AddCommand(runCommand)
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}

}
