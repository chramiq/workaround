package cli

import (
	"strings"

	"github.com/chramiq/workaround/internal/cloudflare"
	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/chramiq/workaround/internal/utils"
)

func Cleanup(cfg *config.Config) error {
	totalDeleted := 0

	for _, acc := range cfg.Accounts {
		mgr := cloudflare.NewManager(acc.AccountID, acc.APIToken, nil)

		workers, err := mgr.ListWorkers()
		if err != nil {
			ui.Error("Failed to list workers for account %s: %v", utils.ShortID(acc.AccountID), err)
			continue
		}

		for _, name := range workers {
			if strings.HasPrefix(name, cfg.Settings.WorkerPrefix) {
				ui.Step("Deleting %s...", name)

				if err := mgr.DeleteWorker(name); err != nil {
					ui.Error("Failed to delete %s: %v", name, err)
				} else {
					totalDeleted++
				}
			} else {
				ui.Debug("Skipping unrelated worker: %s", name)
			}
		}
	}

	st, err := config.LoadState()
	if err == nil {
		st.Clear()
		st.Save()
	}

	if totalDeleted > 0 {
		ui.Success("Cleanup complete. Deleted %d workers.", totalDeleted)
	}

	return nil
}
