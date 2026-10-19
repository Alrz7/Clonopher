package core

import (
	"clonopher/app/burrow"
	"clonopher/app/config"
	"clonopher/app/providers"
	"clonopher/internal"
	"path/filepath"
)

// payload is a high value load that is carried by a worker;
// each one has its own file, bin, desctiption, etc...
type Payload struct {
	Name        string
	Src         string
	Head        string
	Stage       *burrow.Stage
	Info        *ObjectInfo
	JevProvider *providers.Provider
	LlmProvider *providers.Provider
	// Actions     *Actions
	Road    []*burrow.Stage
	Options *Options
}
type Classifier interface {
	Apply(p *Payload, req *providers.JevRequest) error
	Resolve(p *Payload, resp *providers.JevResponse) error
}
type ClsConfig struct {
	Classifiers map[string]*CustomCls `yaml:"classifiers"`
}

type releventCls []struct {
	Cls   *CustomCls
	Power float64
}

type CustomCls struct {
	Name         string                  `yaml:"name"` // its the cls.title thats important, not the key/title inside map
	Type         providers.JevQaTypes    `yaml:"type"`
	Instructions string                  `yaml:"instructions"`
	Brief        string                  `yaml:"brief"`
	Criteria     map[string]CstmCriteria `yaml:"criteria"`
	AddRoots     bool                    `yaml:"add_roots_to_Criteria"`
	// Power        float64                 `yaml:"-"`
}

type CstmCriteria struct {
	// Key         string `yaml:"key"`
	Description string `yaml:"description"`
	Action      string `yaml:"action"`
}

type ObjectInfo struct {
	Description string            `json:"descruption"`
	Options     internal.Envelope `json:"options"`
}
type Options struct {
	RunMode      string
	ActionPolicy string
}

var ExistingCls = ClsConfig{
	Classifiers: map[string]*CustomCls{},
}

func New(c *config.Config) *Payload {
	Payload := Payload{
		Head:        filepath.Base(c.Cli.Head),
		Name:        filepath.Base(c.Cli.Src),
		Src:         c.Cli.Src,
		Info:        &ObjectInfo{},
		Stage:       burrow.New(c.Cli.Head),
		JevProvider: c.GetActiveJevProvider(),
		LlmProvider: c.GetActiveLlmProvider(),
		// Actions: &Actions{
		// 	Current: &ACTUnDefined{
		// 		ShouldReServe: true,
		// 	},
		// 	PerType: map[string]float64{},
		// 	All:     map[*Action]float64{},
		// },
		Options: &Options{
			ActionPolicy: c.App.ActionPolicy,
			RunMode:      c.App.RunMode,
		},
	}
	return &Payload
}

func (p *Payload) NewRequest() *providers.JevRequest {
	state := p.GetState()
	req := providers.JevRequest{
		Provider:  p.JevProvider,
		State:     state,
		Questions: map[string]providers.JevQuestion{},
		Model:     p.JevProvider.ActiveModel,
	}
	return &req
}
