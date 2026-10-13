package core

import "clonopher/app/classifiers"

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
