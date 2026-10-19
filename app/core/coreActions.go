package core

import (
	"clonopher/app/burrow"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"maps"
	"path/filepath"
	"slices"
)

type Action interface {
	Run(p *Payload) error
	ReServe() bool
	Type() string
}

type Actions struct {
	Current Action
	PerType map[string]float64
	All     map[*Action]float64
	//...
}

func (a *Actions) Set(action Action, power float64) {
	a.PerType[action.Type()] = +power
	a.All[&action] = power
}

func (a *Actions) GetWinningAction() {
	var res struct {
		action *Action
		power  float64
	}
	var c struct {
		tp string
		pw float64
	}
	for tp, pw := range a.PerType {
		if pw > c.pw {
			c.tp = tp
			c.pw = pw
		}
	}
	for action, power := range a.All {
		if (*action).Type() == c.tp {
			if power >= res.power {
				res.action = action
				res.power = power
			}
		}
	}
	if res.action == nil {
		loggy.Warn("no action was dedicated!")
		a.Current = &ACTUncertain{ShouldReServe: false}
	} else {
		a.Current = *res.action
	}
}

// ------------ ActMove ------------
type ActMove struct {
	ShouldReServe bool
	Dst           *burrow.Stage
}

func (a *ActMove) Type() string {
	return internal.ActMove
}

func (a *ActMove) Run(p *Payload) error {
	options := slices.Collect(maps.Keys(p.Stage.Roots))
	slices.Sort(options)
	loggy.Info("options: ", "options", options)
	dstName := filepath.Base(a.Dst.Path)
	loggy.Info("moving to --> " + dstName)
	p.SetNext(a.Dst)
	return nil
}

func (a *ActMove) ReServe() bool {
	return a != nil && a.ShouldReServe
}

// -------------- ACTUncertain -----------
type ACTUnDefined struct {
	ShouldReServe bool
}

func (a *ACTUnDefined) Run(p *Payload) error {
	loggy.Info("action Undefined")
	return nil
}
func (a *ACTUnDefined) ReServe() bool {
	return a != nil && a.ShouldReServe
}

func (a *ACTUnDefined) Type() string {
	return internal.ACTUndefined
}

// -------------- ACTUncertain -----------
type ACTUncertain struct {
	ShouldReServe bool
}

func (a *ACTUncertain) Run(p *Payload) error {
	loggy.Info("im realy uncertain about this!, ngl")
	return nil
}
func (a *ACTUncertain) ReServe() bool {
	return a != nil && a.ShouldReServe
}

func (a *ACTUncertain) Type() string {
	return internal.ACTUncertain
}

// -------------- ACTCreateNewFolder -----------
type ACTCreateNewFolder struct {
	ShouldReServe bool
}

func (a *ACTCreateNewFolder) Type() string {
	return internal.ActCreateNew
}

func (a *ACTCreateNewFolder) Run(p *Payload) error {
	loggy.Info("creating new Folder")
	return p.NewStageWithInfo()
}
func (a *ACTCreateNewFolder) ReServe() bool {
	return a != nil && a.ShouldReServe
}

// ---------------- ACTStayHere --------------
type ACTStayHere struct {
	ShouldReServe bool
}

func (a *ACTStayHere) Type() string {
	return internal.ActStay
}

func (a *ACTStayHere) Run(p *Payload) error {
	switch p.Options.ActionPolicy {
	case internal.PlcCopy:
		finalPath, err := p.Stage.CopyFileHere(p.Src) // returns final path too
		if err != nil {
			return err
		}
		loggy.Info("Lets Drop it here!", "action", "copy", "final path", finalPath)
	case internal.PlcMove:
		finalPath, err := p.Stage.DropFileHere(p.Src) // returns final path too
		if err != nil {
			return err
		}
		loggy.Info("Lets Drop it here!", "action", "move", "final path", finalPath)
	default:
		finalPath := filepath.Join(p.Stage.Path, filepath.Base(p.Src))
		loggy.Info("Lets Drop it here!", "action", "notification", "final path", finalPath)
	}
	return nil
}
func (a *ACTStayHere) ReServe() bool {
	return a != nil && a.ShouldReServe
}
