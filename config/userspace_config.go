package config

import (
	"os"
	"path/filepath"
)

func NewUserspaceConfig() (*ProjectConfigInterface, error) {
	dirname, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	userspaceConfigPath := filepath.Join(dirname, ".toki.json")

	_, err = os.Stat(userspaceConfigPath)

	if err != nil {
		return nil, nil
	}

	config, err := NewProjectConfig(userspaceConfigPath)

	if err != nil {
		return nil, err
	}

	return &config, err
}
