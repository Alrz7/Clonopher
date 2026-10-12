package core

import (
	"clonopher/app/jevkit"
	"clonopher/app/services/classifier"
	"clonopher/internal"
	"fmt"
	"maps"
	"slices"
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
	err = p.Act(resp)
	if err != nil {
		return err
	}

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
	// _, err := dOpt.NewRequest()
	if err != nil {
		return nil, err
	}
	// fmt.Println("options: ", slices.Collect(maps.Keys(p.Stage.Roots)))
	// fmt.Println(req)
	resp, err := req.SendTo(p.Options.Model)
	if err != nil {
		return nil, err
	}

	// var resp = &jevkit.Response{}
	return resp, nil
}

func (p *Payload) Act(resp *jevkit.Response) error {
	choise := resp.Answers[internal.Action].Choice
	Next, ok := p.Stage.Roots[choise]
	switch {
	case ok:
		p.Stage.Next = Next
		Next.Previous = p.Stage
		fmt.Println("options: ", slices.Collect(maps.Keys(p.Stage.Roots)))
		fmt.Println("moving to --> " + choise)
		p.Stage = Next
		err := p.Serve()
		if err != nil {
			return err
		}
	case choise == internal.ClrCreateNewFolder:
		fmt.Println("seems like we need to create new folder cuh!")
	case choise == internal.ClrStayHere:
		fmt.Println("lets drop it here")
	case choise == internal.ClrUncertain:
		fmt.Println("im realy uncertain about thisshit fam!")
	}
	return nil
}
