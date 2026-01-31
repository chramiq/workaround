package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/proxy"
	"github.com/chramiq/workaround/internal/ui"
)

func RunWrapper(cfg *config.Config, args []string, useHTTP, unsafeHTTP, randomUA, forceNew bool) error {
	// 1. Load Workers

	st, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %v", err)
	}
	if len(st.ActiveWorkers) == 0 {
		return fmt.Errorf("no workers found. run 'workaround deploy' first")
	}

	// 2. Load User Agents
	var userAgents []string
	uaPath, _ := config.GetUserAgentsPath()
	if content, err := os.ReadFile(uaPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				userAgents = append(userAgents, strings.TrimSpace(line))
			}
		}
	}

	// 3. Determine Target Scheme
	targetScheme := "https"
	if useHTTP || unsafeHTTP {
		targetScheme = "http"
	}

	// 4. Start Proxy Server
	// We pass 0 as port (implicit in Start logic usually, or we bind ":0")
	// The current Server implementation takes ":0" in Start(addr).
	
	proxyServer := proxy.NewServer(
		st.ActiveWorkers,
		targetScheme,
		userAgents,
		randomUA,
		cfg.Settings.UpstreamProxy,
		forceNew,
	)
	
	// Start on random port
	addr, err := proxyServer.Start("127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start proxy: %v", err)
	}
	
	ui.Info("Proxy started on %s", addr)

	// 5. Construct Command
	if len(args) == 0 {
		return fmt.Errorf("no command provided")
	}
	
	cmdName := args[0]
	cmdArgs := args[1:]
	
	// Set Environment
	proxyURL := fmt.Sprintf("http://%s", addr)
	os.Setenv("HTTP_PROXY", proxyURL)
	os.Setenv("HTTPS_PROXY", proxyURL)
	os.Setenv("http_proxy", proxyURL)
	os.Setenv("https_proxy", proxyURL)
	
	// Execute
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// Don't error out, just return the exit code if possible
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		return err
	}
	
	return nil
}
