package main

import (
	"os"

	"github.com/chramiq/workaround/internal/cli"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List active workers",
	Run: func(cmd *cobra.Command, args []string) {
		// Status command should always show output
		ui.SetVerbose(true)
		if err := cli.List(cfg); err != nil {
			ui.Error("List failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
