package app

import (
	"clonopher/api"
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"context"
	"flag"
	"path/filepath"
)

func Main() {
	var err error
	config.Default.App.AppConfigDir, err = internal.SetAppDir()
	if err != nil {
		loggy.Fatal(err)
	}
	options := SetConfig()

	err = core.ExistingCls.LoadExistingClassifiers(filepath.Join(options.App.AppConfigDir, internal.ClsDirName))
	if err != nil {
		loggy.Fatal(err)
	}

	switch options.App.RunMode {
	case internal.CliMode, internal.DryRunMode:
		CliRun(options)
	case internal.ApiMode:
		ctx, cancel := context.WithCancel(context.Background())
		apiConf := api.ApiConfig{
			Ctx:     ctx,
			Close:   cancel,
			Options: options,
		}
		apiConf.Run()
	default:
		loggy.Fatalf("run mode [%v] is NOT valid!", options.App.RunMode)
	}
}

func SetConfig() *config.Config {
	// general flags --
	configPath := flag.String("config", "", "config path")
	flag.StringVar(&config.Default.App.RunMode, "mode", internal.CliMode, "run mode options: < cli, api >")
	flag.StringVar(&config.Default.App.ActionPolicy, "policy", internal.PlcNotify, "what should program do after raching the point: options: < copy, move, notify, ask-copy, ask-move >")

	// api flags --
	flag.StringVar(&config.Default.App.Env, "env", "development", "Envirement")
	flag.IntVar(&config.Default.Api.Port, "port", 6520, "api port")

	// cli flags --
	flag.StringVar(&config.Default.Cli.Head, "head", "", "the Head Directory (start point)")
	flag.StringVar(&config.Default.Cli.Src, "src", "", "file path (source)")
	// flag.StringVar(&config.Default.Cli.Info.Description, "info", "", "file Info")

	flag.Parse()

	if *configPath != "" {
		config.Default.Setup(*configPath)
	} else {
		config.Default.Setup(filepath.Join(config.Default.App.AppConfigDir, "clonopher.config.yaml"))
	}
	// ...input validation...

	return &config.Default
}
