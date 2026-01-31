package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type WorkerInfo struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	AccountID string    `json:"account_id"`
	Created   time.Time `json:"created"`
}

type State struct {
	ActiveWorkers []WorkerInfo `json:"active_workers"`
	LastUpdated   time.Time    `json:"last_updated"`
}

func GetStatePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

func LoadState() (*State, error) {
	path, err := GetStatePath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &State{ActiveWorkers: []WorkerInfo{}}, nil
	}

	var s State
	err = json.Unmarshal(file, &s)
	return &s, err
}

func (s *State) Save() error {
	path, err := GetStatePath()
	if err != nil {
		return err
	}
	s.LastUpdated = time.Now()
	data, err := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(path, data, 0600)
}

func (s *State) AddWorker(w WorkerInfo) {
	s.ActiveWorkers = append(s.ActiveWorkers, w)
}

func (s *State) Clear() {
	s.ActiveWorkers = []WorkerInfo{}
}

func (s *State) CountForAccount(accountID string) int {
	count := 0
	for _, w := range s.ActiveWorkers {
		if w.AccountID == accountID {
			count++
		}
	}
	return count
}
