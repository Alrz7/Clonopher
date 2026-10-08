package core

import (
	"clonopher/app/services/classifier"
	"fmt"
)

func (p *Payload) Serve() error {
	err := p.Stage.Explore()
	if err != nil {
		return err
	}
	dOpt := classifier.DecideOptions{
		Name:     p.Name,
		Stage:    p.Stage,
		Info:     p.Info,
		Provider: p.Options.Provider,
	}
	req, err := dOpt.NewRequest()
	if err != nil {
		return err
	}
	resp, err := req.SendTo(p.Options.Model)
	if err != nil {
		return err
	}
	fmt.Println(req)
	fmt.Println("----------result-----------")
	fmt.Println(resp)
	return nil
}
