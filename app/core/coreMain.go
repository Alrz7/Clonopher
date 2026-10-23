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
	// loggy.Info("-start-")
	var releventClassifiers *releventCls

	done := make(chan struct{})
	loggy.Info(p.Options.RunMode)
	tui, err := StartPayloadTUI(p.Options.RunMode == internal.CliMode, done)
	if err != nil {
		p.SetState(PayloadStateFailed)
		return "", err
	}

	run := func() (bool, error) {
		p.AddIteration()

		if !p.Stage.Explored.Load() {
			p.SetState(PayloadStateExploring)
			err := p.Stage.Explore()
			if err != nil {
				return false, err
			}
		}

		p.Stage.Mu.RLock()
		if p.Stage.Info == nil {
			p.Stage.Mu.RUnlock()
			p.SetState(PayloadStateInferring)
			err := p.Stage.InferStageInfo(p.LlmProvider)
			if err != nil {
				return false, err
			}
		} else {
			p.Stage.Mu.RUnlock()
		}

		req := p.NewRequest()

		var err error
		if releventClassifiers == nil {
			p.SetState(PayloadStateSelectingClassifiers)
			releventClassifiers, err = p.SelectRelevantClassifiers(&ExistingCls)
			if err != nil {
				return false, err
			}
		}

		p.SetState(PayloadStateClassifying)
		actions, err := p.Classify(req, releventClassifiers)
		if err != nil || actions == nil {
			return false, err
		}

		p.SetState(PayloadStateModifying)
		err = actions.Current.Run(p)
		if err != nil {
			return false, err
		}
		p.ShiftAction(actions.Current.Type())
		return actions.Current.ReServe(), nil
	}

	reserve, err := run()
	if err != nil {
		p.SetState(PayloadStateFailed)
		return "", err
	}
	for reserve {
		reserve, err = run()
		if err != nil {
			p.SetState(PayloadStateFailed)
			return "", err
		}
	}

	p.SetState(PayloadStateCompleted)

	close(done)
	if err := tui.Wait(); err != nil {
		p.SetState(PayloadStateFailed)
		return "", err
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
