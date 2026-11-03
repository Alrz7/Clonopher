package core

import (
	"clonopher/app/burrow"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"maps"
	"path/filepath"
	"slices"
	"strings"
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
	a.PerType[action.Type()] += power
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
	Dst           *burrow.Directory
}

func (a *ActMove) Type() string {
	return internal.ActMove
}

func (a *ActMove) Run(p *Payload) error {
	p.Dir.Mu.RLock()
	defer p.Dir.Mu.RUnlock()
	options := slices.Collect(maps.Keys(p.Dir.Children))
	slices.Sort(options)
	loggy.AddNewf("options: %v", loggy.PdLInfo, strings.Join(options, ","))
	// filepath.Base(a.Dst.Path)
	dstName := filepath.Base(a.Dst.Path)
	loggy.AddNewf("moving to --> %v", loggy.PdLInfo, dstName)
	p.SetNext(a.Dst)
	return nil
}

func (a *ActMove) ReServe() bool {
	return a != nil && a.ShouldReServe
}

// -------------- ACTUndefined -----------
type ACTUnDefined struct {
	ShouldReServe bool
}

func (a *ACTUnDefined) Run(p *Payload) error {
	loggy.AddNew("action Undefined", loggy.PdLInfo)
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
	loggy.AddNew("im realy uncertain about this!, ngl", loggy.PdLInfo)
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
	loggy.AddNewf("creating new Folder", loggy.PdLInfo)
	return p.NewDirectoryWithInfo()
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
	p.ShiftNewInfof("Session Complete! Final destination: <%v>", p.Dir.Path)
	return nil
}

func (a *ACTStayHere) ReServe() bool {
	return a != nil && a.ShouldReServe
}

//--------------

func (p *Payload) Drop() error {
	for _, src := range p.Srcs {
		switch p.Options.ActionPolicy {
		case internal.PlcCopy:
			finalPath, err := p.Dir.CopyFileHere(src)
			if err != nil {
				return err
			}
			loggy.AddNewf("Lets Drop it here: %v", loggy.PdLInfo, finalPath)
		case internal.PlcMove:
			finalPath, err := p.Dir.DropFileHere(src)
			if err != nil {
				return err
			}
			loggy.AddNewf("Lets Drop it here: %v", loggy.PdLInfo, finalPath)
		case internal.PlcNone:
			return nil
		default:
			finalPath := filepath.Join(p.Dir.Path, filepath.Base(src))
			loggy.AddNewf("Lets Drop it here; ", loggy.PdLInfo, finalPath)
		}
	}
	return nil
}
