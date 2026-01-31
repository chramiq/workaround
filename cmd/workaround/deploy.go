package main

import (
	"os"

	"github.com/chramiq/workaround/internal/cli"
	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy workers based on config",
	Run: func(cmd *cobra.Command, args []string) {
		ensureWorkerScript()
		if err := cli.Deploy(cfg); err != nil {
			ui.Error("Deployment failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func ensureWorkerScript() {
	scriptPath, _ := config.GetScriptPath()
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		if err := os.WriteFile(scriptPath, config.DefaultWorkerScript, 0644); err == nil {
			ui.Success("Default worker script restored at: %s", scriptPath)
		}
	}
}
