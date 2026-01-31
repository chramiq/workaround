package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Account struct {
	AccountID string `json:"account_id"`
	APIToken  string `json:"api_token"`
}

type Settings struct {
	WorkerCount   int    `json:"worker_count"`
	WorkerPrefix  string `json:"worker_prefix"`
	UpstreamProxy string `json:"upstream_proxy"`
}

type Config struct {
	Accounts []Account `json:"accounts"`
	Settings Settings  `json:"settings"`
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "workaround")
	return configDir, nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func GetScriptPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "worker.js"), nil
}

func GetUserAgentsPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "useragents.txt"), nil
}

func Initialize() (bool, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return false, err
		}
	}

	created := false
	configPath := filepath.Join(dir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := Config{
			Accounts: []Account{
				{AccountID: "YOUR_ACCOUNT_ID_HERE", APIToken: "YOUR_API_TOKEN_HERE"},
			},
			Settings: Settings{
				WorkerCount:   2,
				WorkerPrefix:  "wa-",
				UpstreamProxy: "",
			},
		}
		data, _ := json.MarshalIndent(defaultConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return false, fmt.Errorf("failed to create config: %w", err)
		}
		created = true
	}

	scriptPath := filepath.Join(dir, "worker.js")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		if err := os.WriteFile(scriptPath, DefaultWorkerScript, 0644); err != nil {
			return false, fmt.Errorf("failed to create worker script: %w", err)
		}
	}

	uaPath := filepath.Join(dir, "useragents.txt")
	if _, err := os.Stat(uaPath); os.IsNotExist(err) {
		if err := os.WriteFile(uaPath, DefaultUserAgents, 0644); err != nil {
			return false, fmt.Errorf("failed to create UA list: %w", err)
		}
	}

	return created, nil
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
