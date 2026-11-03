package core

import (
	"clonopher/app/burrow"
	"clonopher/app/config"
	"clonopher/app/providers"
	"clonopher/internal"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// payload is a high value load that is carried by a worker;
// each one has its own file, bin, desctiption, etc...
type Payload struct {
	Id          string
	Name        string
	Srcs        []string
	Head        string
	Dir         *burrow.Directory
	Info        *ObjectInfo
	JevProvider *providers.Provider
	LlmProvider *providers.Provider
	Options     *Options
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

func New(c *config.Config, name, head string, src []string, desc, policy, runMode string) (*Payload, error) {
	// need to do the validation here
	dir, ok := burrow.GetFromGlobals(head)
	if !ok {
		var err error
		dir, err = burrow.NewGlobal("", head)
		if err != nil {
			return nil, err
		}
	}

	sum := sha256.Sum256([]byte(name + head + strings.Join(src, "-") + policy + runMode + time.Now().String()))
	id := hex.EncodeToString(sum[:])

	Payload := Payload{
		Id:          id,
		Head:        head,
		Name:        name,
		Srcs:        src,
		Info:        &ObjectInfo{},
		Dir:         dir,
		JevProvider: c.GetActiveJevProvider(),
		LlmProvider: c.GetActiveLlmProvider(),
		Options: &Options{
			ActionPolicy: policy,
			RunMode:      runMode,
		},
	}
	Stats.AddNew(id, name, dir)

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
