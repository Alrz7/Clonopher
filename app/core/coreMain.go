package core

import (
	"clonopher/app/providers"
	"clonopher/internal"
	"clonopher/internal/loggy"
	"context"
	"encoding/json"
	"time"
)

func (p *Payload) Serve() (string, error) {
	var releventClassifiers *releventCls
	// iteration := 0

	run := func() (bool, error) {
		// iteration++
		// if iteration >= 10 {
		// 	return false, loggy.Sayf("unusual CLS loop! for %v", p.Name)
		// }
		// loggy.Info(
		// 	"serve iteration started",
		// 	"file", p.Name,
		// 	"iteration", iteration,
		// 	"stage", p.Stage.Path,
		// )

		err := p.Stage.Explore()
		if err != nil {
			return false, err
		}

		err = p.Stage.GuessStageInfo(p.LlmProvider)
		if err != nil {
			return false, err
		}

		req := p.NewRequest()

		if releventClassifiers == nil {
			releventClassifiers, err = p.SelectRelevantClassifiers(&ExistingCls)
			if err != nil {
				return false, err
			}
		}
		actions, err := p.Classify(req, releventClassifiers)
		if err != nil || actions == nil {
			return false, err
		}
		err = actions.Current.Run(p)
		if err != nil {
			return false, err
		}
		// action := actions.Current.Type()
		// stageAfter := p.Stage.Path
		// reserve := actions.Current.ReServe()
		// myRoad := []string{}
		// for _, st := range p.Road {
		// 	myRoad = append(myRoad, st.Path)
		// }

		// loggy.Info(
		// 	"action completed",
		// 	"file", p.Name,
		// 	"iteration", iteration,
		// 	"action", action,
		// 	"reServe", reserve,
		// 	"stage", stageAfter,
		// 	"road", p.Road,
		// )
		return actions.Current.ReServe(), nil
	}

	reserve, err := run()
	if err != nil {
		return "", err
	}
	for reserve {
		// loggy.Info("======== reserve =========")
		// time.Sleep(4 * time.Second)
		reserve, err = run()
		if err != nil {
			return "", err
		}
	}
	return p.Stage.Path, nil
}

func (p *Payload) Classify(req *providers.JevRequest, classifiers *releventCls) (*Actions, error) {
	actions := Actions{
		PerType: make(map[string]float64),
		All:     make(map[*Action]float64),
	}

	for _, cl := range *classifiers {
		err := cl.Cls.Apply(p, req)
		if err != nil {
			return nil, err
		}
	}

	if p.Options.RunMode == internal.DryRunMode {
		b, err := json.MarshalIndent(req, "", "\t")
		if err != nil {
			loggy.Fatal(err)
		}
		loggy.Info(string(b))
		return nil, nil
	}

	// b, _ := json.MarshalIndent(req, "", "\t")
	// loggy.Info(string(b))

	sctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	req.State = p.GetState()
	resp, err := req.Send(sctx)
	if err != nil {
		return nil, err
	}

	for _, cl := range *classifiers {
		err := cl.Cls.Resolve(p, &actions, resp, cl.Power)
		if err != nil {
			return nil, err
		}
	}

	actions.GetWinningAction()
	return &actions, nil
}
