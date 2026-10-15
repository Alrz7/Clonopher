package app

import (
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal/loggy"
)

func CliRun(c *config.Config) {
	Payload := core.New(c)
	err := Payload.Serve()
	if err != nil {
		loggy.Error(err)
	}
}
