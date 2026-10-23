package core

import (
	"clonopher/app/burrow"
	"clonopher/app/config"
	"clonopher/app/providers"
	"clonopher/internal"
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
	Name         string                  `yaml:"-"` // its the cls.title thats important, not the key/title inside map
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

func New(c *config.Config, name, head, src, desc, policy, runMode string) (*Payload, error) {
	// need to do the validation here
	stage, ok := burrow.GetFromGlobals(head)
	if !ok {
		var err error
		stage, err = burrow.NewGlobal("", head)
		if err != nil {
			return nil, err
		}
	}
	Payload := Payload{
		Head:        head,
		Name:        name,
		Src:         src,
		Info:        &ObjectInfo{},
		Stage:       stage,
		JevProvider: c.GetActiveJevProvider(),
		LlmProvider: c.GetActiveLlmProvider(),
		Options: &Options{
			ActionPolicy: policy,
			RunMode:      runMode,
		},
	}
	return &Payload, nil
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
