package cli

import (
	"fmt"

	"github.com/chramiq/workaround/internal/cloudflare"
	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/chramiq/workaround/internal/utils"
)

const FreeTierLimit = 100000

func CheckCredits(cfg *config.Config) error {
	ui.Info("Fetching daily usage data from Cloudflare...")
	
	headers := []string{"ACCOUNT", "USED", "LIMIT", "STATUS"}
	var rows [][]string

	for _, acc := range cfg.Accounts {
		mgr := cloudflare.NewManager(acc.AccountID, acc.APIToken, nil)
		usage, err := mgr.GetDailyUsage()
		
		status := "OK"
		limitStr := fmt.Sprintf("%d", FreeTierLimit)
		
		if err != nil {
			status = "ERROR"
			ui.Debug("Failed to get usage for %s: %v", utils.ShortID(acc.AccountID), err)
		} else {
			if usage > int(float64(FreeTierLimit)*0.9) {
				status = "WARNING"
			}
			if usage >= FreeTierLimit {
				status = "EXCEEDED"
			}
		}
		
		usageStr := "-"
		if err == nil {
			usageStr = fmt.Sprintf("%d", usage)
		}

		rows = append(rows, []string{
			utils.ShortID(acc.AccountID),
			usageStr,
			limitStr,
			status,
		})
	}

	ui.PrintTable(headers, rows)
	return nil
}
