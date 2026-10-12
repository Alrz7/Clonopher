package app

import (
	"clonopher/api"
	"clonopher/app/config"
	"clonopher/app/services/composer"
	"context"
	"flag"
)

const (
	cliMode string = "cli"
	apiMode string = "api"
)

type CliRunOptions struct {
	Head string
	Src  string
}

func Main() {
	options := SetConfig()
	if options.App.RunMode == "cli" {
		CliRun(options)
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		apiConf := api.ApiConfig{
			Ctx:     ctx,
			Close:   cancel,
			Options: options,
		}
		apiConf.Run()
	}

}

func SetConfig() *config.Config {
	var options *config.Config = &config.Config{
		App: &config.App{},
		Api: &config.ApiOptions{},
		Cli: &config.CliOptions{
			Info: &composer.FileInfo{},
		},
	}
	// general flags --
	flag.StringVar(&options.App.RunMode, "mode", cliMode, "run mode options: < cli, api >")
	configPath := flag.String("config", "", "config path")

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
