package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chramiq/workaround/internal/ui"
)

func AddAlias() error {
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	shell := os.Getenv("SHELL")
	home, _ := os.UserHomeDir()
	configFile := ""
	aliasCmd := fmt.Sprintf("\nalias wa='%s exec'\n", binPath)

	if strings.Contains(shell, "zsh") {
		configFile = filepath.Join(home, ".zshrc")
	} else if strings.Contains(shell, "bash") {
		if _, err := os.Stat(filepath.Join(home, ".bashrc")); err == nil {
			configFile = filepath.Join(home, ".bashrc")
		} else {
			configFile = filepath.Join(home, ".bash_profile")
		}
	} else if strings.Contains(shell, "fish") {
		configFile = filepath.Join(home, ".config", "fish", "config.fish")
		aliasCmd = fmt.Sprintf("\nalias wa '%s exec'\n", binPath)
	} else {
		return fmt.Errorf("unknown shell: %s. Please add alias manually", shell)
	}

	content, err := os.ReadFile(configFile)
	if err == nil {
		if strings.Contains(string(content), fmt.Sprintf("alias wa='%s exec'", binPath)) {
			ui.Info("Alias 'wa' already exists in %s", configFile)
			return nil
		}
	}

	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(aliasCmd); err != nil {
		return err
	}

	ui.Success("Alias added to %s", configFile)
	ui.Info("Usage: wa <tool> (e.g., 'wa curl http://example.com')")

	return nil
}
