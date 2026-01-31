package main

import (
	"os"

	"github.com/chramiq/workaround/internal/cli"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var creditsCmd = &cobra.Command{
	Use:   "credits",
	Short: "Check daily request usage",
	Long:  "Displays the number of requests made today (UTC) against the Cloudflare Workers free tier limit (100k/day).",
	Run: func(cmd *cobra.Command, args []string) {
		ui.SetVerbose(true)
		if err := cli.CheckCredits(cfg); err != nil {
			ui.Error("Check credits failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(creditsCmd)
}
