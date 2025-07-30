package config

import (
	"errors"
	"os"
	"strconv"
)

type EnvConfig struct {
	EncryptionKey      string
	DomainName         string
	DefaultExpireDays  int
}

func ValidateEnvVars() error {
	if os.Getenv("PMAN_ENCRYPTION_KEY") == "" {
		return errors.New("PMAN_ENCRYPTION_KEY environment variable is required")
	}
	if os.Getenv("PMAN_DOMAIN_NAME") == "" {
		return errors.New("PMAN_DOMAIN_NAME environment variable is required")
	}
	return nil
}

func GetEnvConfig() *EnvConfig {
	expireDays := 24
	if expireStr := os.Getenv("PMAN_DEFAULT_EXPIRE_DAYS"); expireStr != "" {
		if days, err := strconv.Atoi(expireStr); err == nil {
			expireDays = days
		}
	}

	return &EnvConfig{
		EncryptionKey:     os.Getenv("PMAN_ENCRYPTION_KEY"),
		DomainName:        os.Getenv("PMAN_DOMAIN_NAME"),
		DefaultExpireDays: expireDays,
	}
}