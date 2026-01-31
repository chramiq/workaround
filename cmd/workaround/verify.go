package main

import (
	"os"

	"github.com/chramiq/workaround/internal/cli"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Test connectivity of active workers",
	Run: func(cmd *cobra.Command, args []string) {
		// Verify command should always show output
		ui.SetVerbose(true)
		if err := cli.Test(cfg); err != nil {
			ui.Error("Test failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}
