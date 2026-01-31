package main

import (
	"fmt"
	"os"

	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "workaround",
	Short: "workaround - proxify through cloudflare's workers",
	Long: `Workaround is a CLI tool designed to route local network traffic through 
Cloudflare Workers, providing IP rotation and traffic washing.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize UI
		configDir, _ := config.GetConfigDir()
		
		if err := ui.Init(configDir, verbose); err != nil {
			fmt.Printf("Warning: Could not create log file: %v\n", err)
		}
		
		// Skip config loading for 'help' command
		if cmd.Name() == "help" {
			return
		}

		// Try to load configuration
		var err error
		cfg, err = config.Load()
		
		// If config is nil (file missing), initialize it
		if cfg == nil && err == nil {
			ui.SetVerbose(true) // Force verbose to tell user we created config
			ui.Info("Configuration not found. Initializing...")
			
			_, initErr := config.Initialize()
			if initErr != nil {
				ui.Error("Failed to initialize config: %v", initErr)
				os.Exit(1)
			}
			
			ui.Success("Default configuration created at: %s", configDir)
			ui.Info("Please edit 'config.json' with your Cloudflare credentials.")
			os.Exit(0) // Exit so user can edit config
		}

		if err != nil {
			ui.Error("Error loading config: %v", err)
			os.Exit(1)
		}

		// Check for default config
		if len(cfg.Accounts) > 0 && cfg.Accounts[0].AccountID == "YOUR_ACCOUNT_ID_HERE" {
			ui.SetVerbose(true) // Ensure they see this error
			ui.Error("Default configuration detected.")
			ui.Info("Please edit %s/config.json with real API keys.", configDir)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
