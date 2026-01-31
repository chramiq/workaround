package cli

import (
	"fmt"
	"time"

	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
	"github.com/chramiq/workaround/internal/utils"
)

func List(cfg *config.Config) error {
	st, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %v", err)
	}

	if len(st.ActiveWorkers) == 0 {
		ui.Info("No active workers found.")
		ui.Info("Run 'workaround deploy' to create some.")
		return nil
	}

	ui.Info("Current Worker Pool (%d total)", len(st.ActiveWorkers))
	
	headers := []string{"#", "NAME", "ACCOUNT", "CREATED"}
	var rows [][]string

	for i, w := range st.ActiveWorkers {
		age := time.Since(w.Created).Round(time.Minute)
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			w.Name,
			utils.ShortID(w.AccountID),
			fmt.Sprintf("%s ago", age),
		})
	}

	ui.PrintTable(headers, rows)
	return nil
}
