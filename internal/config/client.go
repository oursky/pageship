package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func ClientConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "pageship")
	if err := os.MkdirAll(dir, 0777); err != nil {
		return "", err
	}
	return filepath.Join(dir, "client.json"), nil
}

type ClientConfig struct {
	APIServer      string `json:"apiServer,omitempty"`
	GitHubUsername string `json:"githubUsername,omitempty"`
	SSHKeyFile     string `json:"sshKeyFile,omitempty"`
	AuthToken      string `json:"authToken,omitempty"`
}

func LoadClientConfig() (*ClientConfig, error) {
	path, err := ClientConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new if not exist
			return &ClientConfig{}, nil
		}
		return nil, err
	}

	var conf ClientConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *ClientConfig) Save() error {
	path, err := ClientConfigPath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0700)
	if err != nil {
		return err
	}

	return err
}
