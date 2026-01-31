package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

type Manager struct {
	AccountID  string
	APIToken   string
	HTTPClient *http.Client
}

func NewManager(accountID, token string, client *http.Client) *Manager {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Manager{
		AccountID:  accountID,
		APIToken:   token,
		HTTPClient: client,
	}
}

func (m *Manager) DeployWorker(workerName string, scriptContent []byte) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s", m.AccountID, workerName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 1. Metadata Part (JSON)
	metadataHeader := make(textproto.MIMEHeader)
	metadataHeader.Set("Content-Disposition", `form-data; name="metadata"`)
	metadataHeader.Set("Content-Type", "application/json")
	metaPart, _ := writer.CreatePart(metadataHeader)
	metaPart.Write([]byte(`{"main_module": "worker.js", "body_part": "script"}`))

	// 2. Script Part (JS)
	scriptHeader := make(textproto.MIMEHeader)
	scriptHeader.Set("Content-Disposition", `form-data; name="script"; filename="worker.js"`)
	scriptHeader.Set("Content-Type", "application/javascript")
	scriptPart, _ := writer.CreatePart(scriptHeader)

	scriptPart.Write(scriptContent)

	writer.Close()

	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Set("Authorization", "Bearer "+m.APIToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API Error %d: %s", resp.StatusCode, string(respBody))
	}

	// 3. Enable the Worker on the subdomain
	err = m.enableSubdomain(workerName)
	if err != nil {
		return "", fmt.Errorf("script uploaded but failed to enable subdomain: %v", err)
	}

	// 4. Construct the final URL
	subdomain, err := m.getSubdomain()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.%s.workers.dev", workerName, subdomain), nil
}

func (m *Manager) enableSubdomain(workerName string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s/subdomain", m.AccountID, workerName)

	payload := []byte(`{"enabled": true}`)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+m.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (m *Manager) getSubdomain() (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/subdomain", m.AccountID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+m.APIToken)

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			Subdomain string `json:"subdomain"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Result.Subdomain == "" {
		return "", fmt.Errorf("could not find account subdomain")
	}

	return result.Result.Subdomain, nil
}

func (m *Manager) ListWorkers() ([]string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts", m.AccountID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+m.APIToken)

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var result struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var names []string
	for _, w := range result.Result {
		names = append(names, w.ID)
	}
	return names, nil
}

func (m *Manager) DeleteWorker(workerName string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/workers/scripts/%s", m.AccountID, workerName)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+m.APIToken)

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200 = Deleted, 404 = Already gone
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func (m *Manager) GetDailyUsage() (int, error) {
	url := "https://api.cloudflare.com/client/v4/graphql"

	// Time range: Start of today (UTC) to End of today (UTC)
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-1 * time.Second)

	query := `
		query Viewer($accountTag: string, $start: string, $end: string) {
			viewer {
				accounts(filter: {accountTag: $accountTag}) {
					workersInvocationsAdaptive(limit: 10, filter: {
						datetime_geq: $start,
						datetime_leq: $end
					}) {
						sum {
							requests
						}
					}
				}
			}
		}
	`

	payload := map[string]interface{}{
		"query": query,
		"variables": map[string]string{
			"accountTag": m.AccountID,
			"start":      startOfDay.Format(time.RFC3339),
			"end":        endOfDay.Format(time.RFC3339),
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+m.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Viewer struct {
				Accounts []struct {
					WorkersInvocationsAdaptive []struct {
						Sum struct {
							Requests int `json:"requests"`
						} `json:"sum"`
					} `json:"workersInvocationsAdaptive"`
				} `json:"accounts"`
			} `json:"viewer"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if len(result.Errors) > 0 {
		return 0, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	if len(result.Data.Viewer.Accounts) > 0 && len(result.Data.Viewer.Accounts[0].WorkersInvocationsAdaptive) > 0 {
		return result.Data.Viewer.Accounts[0].WorkersInvocationsAdaptive[0].Sum.Requests, nil
	}

	return 0, nil
}
