package core

import (
	"clonopher/app/jevkit"
	"clonopher/app/services/classifier"
	"clonopher/internal"
	"fmt"
)

func (p *Payload) Serve() error {
	err := p.Stage.Explore()
	if err != nil {
		return err
	}
	
	resp, err := p.Classify()
	if err != nil {
		return err
	}
	p.Act(resp)

	return nil
}

func (p *Payload) Classify() (*jevkit.Response, error) {
	dOpt := classifier.DecideOptions{
		Name:     p.Name,
		Stage:    p.Stage,
		Info:     p.Info,
		Provider: p.Options.Provider,
	}
	req, err := dOpt.NewRequest()
	if err != nil {
		return nil, err
	}
	resp, err := req.SendTo(p.Options.Model)
	if err != nil {
		return nil, err
	}
	fmt.Println(req)
	// var resp *jevkit.Response = &jevkit.Response{}

	return resp, nil
}

func (p *Payload) Act(resp *jevkit.Response) {
	fmt.Println(resp.Answers[internal.Action])
}
