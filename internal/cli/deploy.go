package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/chramiq/workaround/internal/cloudflare"
	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/chramiq/workaround/internal/utils"
)

func Deploy(cfg *config.Config) error {
	ui.Info("Starting deployment check...")

	scriptPath, err := config.GetScriptPath()
	if err != nil {
		return err
	}
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read worker.js: %v", err)
	}

	st, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %v", err)
	}

	totalDeployed := 0

	for _, acc := range cfg.Accounts {
		currentCount := st.CountForAccount(acc.AccountID)
		needed := cfg.Settings.WorkerCount - currentCount

		if needed <= 0 {
			ui.Info("Account %s: OK (%d/%d active)", utils.ShortID(acc.AccountID), currentCount, cfg.Settings.WorkerCount)
			continue
		}

		ui.Info("Account %s: Deploying %d new workers", utils.ShortID(acc.AccountID), needed)

		mgr := cloudflare.NewManager(acc.AccountID, acc.APIToken, nil)

		for i := 0; i < needed; i++ {
			name := utils.GenerateWorkerName(cfg.Settings.WorkerPrefix)
			ui.Step("Deploying %s...", name)

			url, err := mgr.DeployWorker(name, scriptContent)
			if err != nil {
				ui.Error("Deployment failed: %v", err)
				continue
			}

			st.AddWorker(config.WorkerInfo{
				Name:      name,
				URL:       url,
				AccountID: acc.AccountID,
				Created:   time.Now(),
			})
			st.Save()
			totalDeployed++

			time.Sleep(1 * time.Second)
		}
	}

	if totalDeployed > 0 {
		ui.Success("Deployment complete. %d new workers created.", totalDeployed)
	}
	ui.Info("Total pool size: %d workers", len(st.ActiveWorkers))

	return nil
}
