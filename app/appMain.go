package app

import (
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/fileSystem"
	"path/filepath"
)

func Main() error {
	var err error
	config.Default.App.AppConfigDir, err = fileSystem.SetAppDir()
	if err != nil {
		return err
	}
	loadedDefaultConfig, err = config.Default.Setup(filepath.Join(config.Default.App.AppConfigDir, internal.ConfigFileName))
	if err != nil {
		return err
	}
	err = core.ExistingCls.LoadExistingClassifiers(filepath.Join(config.Default.App.AppConfigDir, fileSystem.ClsDirName))
	if err != nil {
		return err
	}

	err = fileSystem.GlobalStagings.LoadAssetConfig()
	if err != nil {
		return err
	}

	setRootFlags()
	setCliFlags()
	setApiFlags()
	rootCmd.AddCommand(apiCmd, cliCmd)

	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
