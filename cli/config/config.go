package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/steve/pman/cli/crypto"
)

type Config struct {
	Server       string `json:"server"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	DefaultGroup string `json:"default_group"`
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(home, ".pman")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	
	return configDir, nil
}

func getConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Token != "" {
		decryptedToken, err := crypto.DecryptClientData(config.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt token: %v", err)
		}
		config.Token = decryptedToken
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	configToSave := *c
	if configToSave.Token != "" {
		encryptedToken, err := crypto.EncryptClientData(configToSave.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt token: %v", err)
		}
		configToSave.Token = encryptedToken
	}

	data, err := json.MarshalIndent(configToSave, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func (c *Config) ClearToken() error {
	c.Token = ""
	return c.Save()
}

func GetGroup() string {
	if group := os.Getenv("PMAN_GROUP"); group != "" {
		return group
	}

	config, err := LoadConfig()
	if err != nil {
		return ""
	}

	return config.DefaultGroup
}