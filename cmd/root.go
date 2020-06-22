package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-mail",
	Short: "go-mail",
	Long:  `go-mail`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() error {
	return rootCmd.Execute()
}
