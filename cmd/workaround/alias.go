package main

import (
	"os"

	"github.com/chramiq/workaround/internal/cli"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Add workaround run -- alias (wa) to your shell",
	Run: func(cmd *cobra.Command, args []string) {
		ui.SetVerbose(true)
		if err := cli.AddAlias(); err != nil {
			ui.Error("Failed to add alias: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(aliasCmd)
}
