package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
)

func Test(cfg *config.Config) error {
	ui.Info("Starting connectivity test...")

	st, err := config.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %v", err)
	}

	if len(st.ActiveWorkers) == 0 {
		ui.Info("No active workers found. Run 'deploy' first.")
		return nil
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	successCount := 0

	for i, w := range st.ActiveWorkers {
		target := "https://httpbin.org/ip"
		testURL := fmt.Sprintf("%s?url=%s", w.URL, target)

		start := time.Now()
		resp, err := client.Get(testURL)
		latency := time.Since(start)

		if err != nil {
			ui.Error("[%d/%d] %s: FAILED (%v)", i+1, len(st.ActiveWorkers), w.Name, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			ui.Error("[%d/%d] %s: HTTP %d", i+1, len(st.ActiveWorkers), w.Name, resp.StatusCode)
			continue
		}

		var result struct {
			Origin string `json:"origin"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			ui.Error("[%d/%d] %s: Invalid JSON", i+1, len(st.ActiveWorkers), w.Name)
			continue
		}

		ui.Success("[%d/%d] %s: 200 OK (%s) -> IP: %s",
			i+1, len(st.ActiveWorkers), w.Name, latency.Round(time.Millisecond), result.Origin)

		successCount++
	}

	ui.Info("Test complete. %d/%d workers healthy.", successCount, len(st.ActiveWorkers))
	return nil
}
