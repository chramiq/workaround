package main

import (
	"os"

	"github.com/chramiq/workaround/internal/cli"
	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var (
	useHTTP    bool
	newCircuit bool
	unsafeHTTP bool
	randomUA   bool
)

var execCmd = &cobra.Command{
	Use:   "exec [flags] -- [command]",
	Short: "Run a tool through the proxy",
	Long: `Starts the local proxy and executes the provided command with HTTP_PROXY environment variables set.
Flags for the 'exec' command must be placed BEFORE the command you want to run.
Example: workaround exec --new-circuit -- curl https://example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		ensureUserAgents()

		// args contains the command and its arguments
		if err := cli.RunWrapper(cfg, args, useHTTP, unsafeHTTP, randomUA, newCircuit); err != nil {
			ui.Error("Exec failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Stop parsing flags after the first positional argument
	execCmd.Flags().SetInterspersed(false)

	execCmd.Flags().BoolVar(&useHTTP, "http", false, "Force the proxy to use HTTP (don't upgrade to HTTPS)")
	execCmd.Flags().BoolVarP(&unsafeHTTP, "unsafe-http", "u", false, "Disable the auto-downgrade safety check")
	execCmd.Flags().BoolVarP(&randomUA, "random-useragent", "r", false, "Enable User-Agent randomization (default: transparent)")
	execCmd.Flags().BoolVarP(&newCircuit, "new-circuit", "c", false, "Force a new Tor exit node for this session")

	rootCmd.AddCommand(execCmd)
}

func ensureUserAgents() {
	uaPath, _ := config.GetUserAgentsPath()
	if _, err := os.Stat(uaPath); os.IsNotExist(err) {
		if err := os.WriteFile(uaPath, config.DefaultUserAgents, 0644); err == nil {
			ui.Success("Default User-Agents restored at: %s", uaPath)
		}
	}
}
