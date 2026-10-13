package core

import (
	"clonopher/app/classifiers"
	"clonopher/app/providers"
	"clonopher/internal"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"time"
)

func (p *Payload) Serve() error {
	err := p.Stage.Explore()
	if err != nil {
		return err
	}

	req := p.Case.NewRequest()
	if err := p.Classify(req, p); err != nil {
		return err
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
		return nil
	}

	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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

func (*Payload) Apply(c *classifiers.Case, req *providers.JevRequest) error {
	var q = providers.JevQuestion{
		Type: "choice",
		Instructions: `Choose the best action for the target file using its name, description, and current directory context.
Prefer an existing child when its purpose strongly matches the file. Do not rely on keyword overlap or absolute paths alone.
Choose create_new only when no existing child fits and the file represents a distinct topic worth a new category.
Choose stay_here when the current directory is the appropriate location and no deeper classification is justified.
Choose uncertain when evidence is insufficient or the leading candidates are too similar. Prefer a cautious decision over an unsupported guess.`,
	}

	criteria := make(map[string]string, len(c.Stage.Roots)+3)
	for rootName, root := range c.Stage.Roots {
		infoOptions := helpers.InfoOptions{
			&helpers.InfoOption{Info: `name: %v`, Options: []any{rootName}},
			&helpers.InfoOption{Info: `path: %v`, Options: []any{root.Path}},
			&helpers.InfoOption{Info: `info: %v`, Options: []any{root.Info.Render()}},
			&helpers.InfoOption{Info: `these are the files existing here: %v`, Options: []any{root.Files}},
		}
		info := infoOptions.ConcatIfExists()
		criteria[rootName] = info
	}
	criteria[internal.ClrCreateNewFolder] = `No existing child matches the file's purpose well enough. The file represents a distinct topic that justifies a new category.`
	criteria[internal.ClrStayHere] = `The current directory is the best fit. No existing child is sufficiently relevant, and creating a new category is unnecessary.`
	criteria[internal.ClrUncertain] = `Available evidence is insufficient, or the leading candidates are too similar to distinguish reliably. Do not guess.`

	q.Criteria = criteria
	req.Questions[internal.Action] = q
	return nil
}

func (p *Payload) Resolve(_ *classifiers.Case, resp *providers.JevResponse) error {
	choise := resp.Answers[internal.Action].Choice
	if choise == "" {
		return loggy.Say("empty choise, seponse was not valid")
	}
	Next, ok := p.Stage.Roots[choise]
	switch {
	case ok:
		p.Stage.Next = Next
		Next.Previous = p.Stage
		options := slices.Collect(maps.Keys(p.Stage.Roots))
		slices.Sort(options)
		fmt.Println("options: ", options)
		fmt.Println("moving to --> " + choise)
		p.Stage = Next
		err := p.Serve()
		if err != nil {
			return err
		}
	case choise == internal.ClrCreateNewFolder:
		fmt.Println("seems like we need to create new folder cuh!")
	case choise == internal.ClrStayHere:
		switch p.Options.ActionPolicy {
		case internal.PlcCopy:
			finalPath, err := p.Stage.CopyFileHere(p.Src) // returns final path too
			if err != nil {
				return err
			}
			loggy.Info("Lets Drop it here!", "action", "copy", "final path", finalPath)
		case internal.PlcMove:
			finalPath, err := p.Stage.DropFileHere(p.Src) // returns final path too
			if err != nil {
				return err
			}
			loggy.Info("Lets Drop it here!", "action", "move", "final path", finalPath)
		default:
			finalPath := filepath.Join(p.Stage.Path, filepath.Base(p.Src))
			loggy.Info("Lets Drop it here!", "action", "notification", "final path", finalPath)
		}
	case choise == internal.ClrUncertain:
		fmt.Println("im realy uncertain about this shit fam!")
	}
	return nil
}
