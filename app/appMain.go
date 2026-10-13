package app

import (
	"clonopher/api"
	"clonopher/app/classifiers"
	"clonopher/app/config"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"context"
	"flag"
)

func Main() {
	options := SetConfig()
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
	var options *config.Config = &config.Config{
		App: &config.App{},
		Api: &config.ApiOptions{},
		Cli: &config.CliOptions{
			Info: &classifiers.ObjectInfo{},
		},
	}
	// general flags --
	configPath := flag.String("config", "", "config path")
	flag.StringVar(&options.App.RunMode, "mode", internal.CliMode, "run mode options: < cli, api >")
	flag.StringVar(&options.App.ActionPolicy, "policy", internal.PlcNotify, "what should program do after raching the point: options: < copy, move, notify, ask-copy, ask-move >")

	// api flags --
	flag.StringVar(&options.App.Env, "env", "development", "Envirement")
	flag.IntVar(&options.Api.Port, "port", 6520, "api port")

	// cli flags --
	flag.StringVar(&options.Cli.Head, "head", "", "the Head Directory (start point)")
	flag.StringVar(&options.Cli.Src, "src", "", "file path (source)")
	flag.StringVar(&options.Cli.Info.Description, "info", "", "file Info")

	flag.Parse()

	if *configPath != "" {
		options.Setup(*configPath)
	}

	// ...input validation...

	return options
}
