package main

import (
	"docker/command"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	command.App.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return initDocker()
	}
	if err := command.App.Execute(); err != nil {
		os.Exit(1)
	}
}
