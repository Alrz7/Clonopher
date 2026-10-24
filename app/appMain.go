package app

import (
	"clonopher/api"
	"clonopher/app/burrow"
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"context"
	"path/filepath"
	"sync"
)

func Main() {
	var err error
	config.Default.App.AppConfigDir, err = internal.SetAppDir()
	if err != nil {
		loggy.Fatal(err)
	}
	config.Default.Setup(filepath.Join(config.Default.App.AppConfigDir, "clonopher.config.yaml"))

	payloads, ok, err := SetConfig()
	if err != nil {
		loggy.Fatal(err)
	}

	for head, path := range config.Default.App.Heads {
		_, err := burrow.NewGlobal(head, path)
		if err != nil {
			loggy.Fatal(err)
		}
	}

	err = core.ExistingCls.LoadExistingClassifiers(filepath.Join(config.Default.App.AppConfigDir, internal.ClsDirName))
	if err != nil {
		loggy.Fatal(err)
	}

	switch config.Default.App.RunMode {
	case internal.CliMode, internal.DryRunMode:
		if payloads == nil || !ok {
			return
		}

		done := make(chan struct{})
		tui, err := core.StartPayloadTUI(done, config.Default.Cli.LogLevel)
		if err != nil {
			loggy.Fatal(err)
		}

		wg := sync.WaitGroup{}
		for _, p := range payloads {
			run := func() {
				_, err := p.Serve()
				if err != nil {
					loggy.Error(err)
				}
			}
			wg.Go(run)
		}

		wg.Wait()

		close(done)
		if err := tui.Wait(); err != nil {
			loggy.Fatal(err)
		}

	case internal.ApiMode:
		if !ok {
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		apiConf := api.ApiConfig{
			Ctx:     ctx,
			Close:   cancel,
			Options: &config.Default,
		}
		apiConf.Run()
	default:
		loggy.Fatalf("run mode [%v] is NOT valid! use -h or --help to see the options.", config.Default.App.RunMode)
	}
}
