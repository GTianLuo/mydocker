package command

import "github.com/spf13/cobra"

var App = &cobra.Command{
	Use:   "mydocker",
	Short: appShort,
	Long:  appLong,
}
