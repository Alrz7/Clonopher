package classifiers

import (
	"clonopher/app/burrow"
	"clonopher/app/providers"
)

type Case struct {
	Name     string
	Stage    *burrow.Stage
	Info     *ObjectInfo
	Provider *providers.Provider
}

type Classifier interface {
	Apply(c *Case, req *providers.JevRequest) error
	Resolve(c *Case, resp *providers.JevResponse) error
}

func (c *Case) NewRequest() *providers.JevRequest {
	state := c.GetState()
	req := providers.JevRequest{
		Provider:  c.Provider,
		State:     state,
		Questions: map[string]providers.JevQuestion{},
		Model:     c.Provider.ActiveModel,
	}
	return &req
}
