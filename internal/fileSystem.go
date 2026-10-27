package internal

import (
	"clonopher/internal/loggy"
	"os"
	"path/filepath"
)

const (
	ConfgDirName = "clonopher"
	ClsDirName   = "classifiers"
	UploadDirName = "ClonopherUploads"
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

func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, loggy.Sayf("invalid path %q: %w", path, err)
	}
	return info.IsDir(), nil
}

func IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, loggy.Sayf("invalid path %q: %w", path, err)
	}
	return info.Mode().IsRegular(), nil
}
