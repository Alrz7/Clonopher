package app

import (
	"clonopher/app/burrow"
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal/loggy"
	"path/filepath"
)

func CliRun(c *config.Config) {
	Payload := core.Payload{
		Head:     filepath.Base(c.Cli.Head),
		Name:     filepath.Base(c.Cli.Src),
		Src:      c.Cli.Src,
		Info:     c.Cli.Info,
		Stage:    burrow.New(c.Cli.Head),
		Provider: c.GetActiveProvider(),
		Options: &core.Options{
			ActionPolicy: c.App.ActionPolicy,
			RunMode:      c.App.RunMode,
		},
	}
	err := Payload.Serve()
	if err != nil {
		loggy.Error(err)
	}
}
