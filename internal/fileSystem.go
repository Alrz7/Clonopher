package internal

import (
	"os"
	"path/filepath"
)

const (
	ConfgDirName = "clonopher"
	ClsDirName   = "classifiers"
)

func SetAppDir() (string, error) {
	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(baseConfigDir, ConfgDirName)
	clsDir := filepath.Join(configDir, ClsDirName)

	for _, dir := range []string{configDir, clsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	return configDir, nil
}
