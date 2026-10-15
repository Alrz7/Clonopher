package core

import (
	"clonopher/app/classifiers"
	"clonopher/app/providers"
	"clonopher/internal"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"context"
	"encoding/json"
	"time"
)

func (p *Payload) Serve() error {
	for helpers.IsNotOneOf(p.Actions.Current, classifiers.AC_UNDEFINED, classifiers.AC_UNCERTAIN, classifiers.AC_CREATE, classifiers.AC_STAY) {
		req, err := p.Case.Serve()
		if err != nil {
			return err
		}
		if err := p.Classify(req, p); err != nil {
			return err
		}
	}
	return nil
}

func (p *Payload) Classify(req *providers.JevRequest, classifires ...classifiers.Classifier) error {
	for _, cl := range classifires {
		err := cl.Apply(&p.Case, req)
		if err != nil {
			return err
		}
	}

	if p.Options.RunMode == internal.DryRunMode {
		b, err := json.MarshalIndent(req, "", "\t")
		if err != nil {
			loggy.Fatal(err)
		}
		loggy.Info(string(b))
		p.Actions.Set(classifiers.AC_UNDEFINED)
		return nil
	}
	// b, _ := json.MarshalIndent(req, "", "\t")
	// loggy.Info(string(b))

	sctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	req.State = p.GetState()
	resp, err := req.Send(sctx)
	if err != nil {
		return err
	}

	for _, cl := range classifires {
		err := cl.Resolve(&p.Case, resp)
		if err != nil {
			return err
		}
	}

	return nil
}
