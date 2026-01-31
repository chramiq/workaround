package main

import (
	"os"

	"github.com/chramiq/workaround/internal/cli"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Delete all deployed workers",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cli.Cleanup(cfg); err != nil {
			ui.Error("Cleanup failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
