package core

import (
	"clonopher/app/burrow"
	"clonopher/app/jevkit"
	"clonopher/app/services/composer"
)

// payload is a high value load that is carried by a worker;
// each one has its own file, bin, desctiption, etc...
type Payload struct {
	Name    string
	Head    string
	Src     string
	Info    *composer.FileInfo
	Stage   *burrow.Stage
	Options *Options
}

type Options struct {
	Provider *jevkit.Provider
	Model    jevkit.JevModel
}