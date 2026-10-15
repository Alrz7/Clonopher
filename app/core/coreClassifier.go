package core

import (
	"clonopher/app/classifiers"
	"clonopher/app/providers"
	"clonopher/internal"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"maps"
	"path/filepath"
	"slices"
)

func (p *Payload) Apply(c *classifiers.Case, req *providers.JevRequest) error {
	var q = providers.JevQuestion{
		Type: "choice",
		Instructions: `Choose the best destination for the target file using its
name, description, and the available directory context.

Follow these priorities:

1. First, determine whether the current directory already fits the file's
   general subject. The current directory is a valid final destination.

2. Prefer an existing child directory when it is a clearly better,
   more specific fit for the file. A perfect match is not required.

3. Choose stay_here when the file belongs to the current directory's
   subject and no existing child is clearly more suitable. Do not create
   a deeper directory merely to classify a file more specifically.

4. Choose create_new when no existing child is a suitable match, but
   the file belongs to a distinct, coherent subcategory of the current
   directory that deserves its own organization.

   Consider the level and purpose of existing sibling directories.
   If the current directory contains separate categories for programming
   languages, frameworks, technologies, or similar topics, a new topic
   at the same level may justify a new sibling directory.

   The file does not need to be unrelated to the current directory.
   A new child directory should be a meaningful specialization of the
   current directory, not an unrelated category.

   Do not require multiple existing files before creating a category.
   However, avoid creating a folder solely to give a single ordinary
   file a more specific location.

5. Choose stay_here when the current directory is a reasonable destination
   and creating another category would add unnecessary complexity.

Evaluate semantic relevance, not just keyword overlap. Use directory
names, descriptions, and existing files as evidence, but do not assume
facts that are not provided. Prefer a sensible existing destination over
unnecessary directory creation.`,
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
	criteria[internal.ClrCreateNewFolder] = `No existing child is a suitable
match, but the file represents a distinct subcategory of the current
directory that deserves its own folder. Consider the existing sibling
categories and their level of abstraction. A new programming language
may justify a sibling folder alongside existing language-specific folders.
The topic must fit within the current directory, not be unrelated to it.
Do not require multiple existing files, but avoid unnecessary folders
for isolated, ordinary files.`

	criteria[internal.ClrStayHere] = `The current directory is an appropriate
destination, and creating a new subcategory would provide little
organizational benefit. Choose this for general-purpose or isolated files
that do not justify a separate category. Do not choose stay_here merely
because the file belongs to the current directory's broad subject.`

	criteria[internal.ClrUncertain] = `Use only when the available evidence
is genuinely insufficient to identify a reasonable destination, or when
multiple candidates are similarly suitable. Lack of a perfect child match
is not sufficient reason to choose uncertain if the current directory fits.`

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
		options := slices.Collect(maps.Keys(p.Stage.Roots))
		slices.Sort(options)
		loggy.Info("options: ", "options", options)
		loggy.Info("moving to --> " + choise)
		p.SetNext(Next)
		p.Actions.Set(classifiers.AC_SERVE)
	case choise == internal.ClrCreateNewFolder:
		loggy.Info("seems like we need to create new folder!")
		err := p.NewStageWithInfo()
		if err != nil {
			return err
		}
		p.Actions.Set(classifiers.AC_CREATE)
	case choise == internal.ClrStayHere:
		p.Actions.Set(classifiers.AC_STAY)
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
		p.Actions.Set(classifiers.AC_UNCERTAIN)
		loggy.Info("im realy uncertain about this shit fam!")
	}
	return nil
}
