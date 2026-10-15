package core

import (
	"clonopher/app/burrow"
	"clonopher/app/classifiers"
	"clonopher/app/config"
	"path/filepath"
)

// payload is a high value load that is carried by a worker;
// each one has its own file, bin, desctiption, etc...
type Payload struct {
	Src     string
	Head    string
	Options *Options
	classifiers.Case
}

type Options struct {
	RunMode      string
	ActionPolicy string
}

func New(c *config.Config) *Payload {
	Payload := Payload{
		Head:     filepath.Base(c.Cli.Head),
		Name:     filepath.Base(c.Cli.Src),
		Src:      c.Cli.Src,
		Info:     c.Cli.Info,
		Stage:    burrow.New(c.Cli.Head),
		Provider: c.GetActiveProvider(),
		Actions:  &classifiers.Actions{},
		Options: &Options{
			ActionPolicy: c.App.ActionPolicy,
			RunMode:      c.App.RunMode,
		},
	}
	return &Payload
}
