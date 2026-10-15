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
	Actions  *Actions
}

type CsAction int
type Actions struct {
	Current CsAction
	Hitory  []CsAction
}
type Classifier interface {
	Apply(c *Case, req *providers.JevRequest) error
	Resolve(c *Case, resp *providers.JevResponse) error
}

const (
	AC_SERVE = iota
	AC_UNDEFINED
	AC_UNCERTAIN
	AC_CREATE
	AC_STAY
)

func (a *Actions) Set(new CsAction) {
	a.Hitory = append(a.Hitory, a.Current)
	a.Current = new
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

func (c *Case) Serve() (*providers.JevRequest, error) {
	err := c.Stage.Explore()
	if err != nil {
		return nil, err
	}

	req := c.NewRequest()
	return req, err
}

func (c *Case) ServeCopy() (*providers.JevRequest, error) {
	cp := *c
	return cp.Serve()
}

func (c *Case) Rebase(dst *Case) {
	*dst = *c
	//...
}
