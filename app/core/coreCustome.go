package core

import (
	"clonopher/app/providers"
	"clonopher/internal"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
)

func (cls *CustomCls) Apply(p *Payload, req *providers.JevRequest) error {
	var q = providers.JevQuestion{
		Type:         cls.Type,
		Instructions: cls.Instructions,
	}
	criteria := make(map[string]string)

	for key, crt := range cls.Criteria {
		criteria[key] = crt.Description
	}
	if cls.AddRoots {
		for rootName, root := range p.Dir.Children {
			infoOptions := helpers.InfoOptions{
				&helpers.InfoOption{Info: `name: %v`, Options: []any{rootName}},
				&helpers.InfoOption{Info: `path: %v`, Options: []any{root.Path}},
				&helpers.InfoOption{Info: `info: %v`, Options: []any{root.Info.Render()}},
				&helpers.InfoOption{Info: `files existing here: %v`, Options: []any{root.Files}},
			}
			info := infoOptions.ConcatIfExists()
			criteria[rootName] = info
		}
	}

	q.Criteria = criteria
	req.Questions[cls.Name] = q
	return nil
}

func (cls *CustomCls) Resolve(p *Payload, ac *Actions, resp *providers.JevResponse, power float64) error {
	answer, ok := resp.Answers[cls.Name]
	if !ok {
		return loggy.Sayf("answer not found or the given Key was wrong")
	}

	Next, ok := p.Dir.Children[answer.Choice]
	if ok {
		action := ActMove{
			ShouldReServe: true,
			Dst:           Next,
		}
		ac.Set(&action, power)
	}
	act, ok := cls.Criteria[answer.Choice]
	if ok {
		switch act.Action {
		case internal.ACTUncertain:
			action := ACTUncertain{
				ShouldReServe: false,
			}
			ac.Set(&action, power)
		case internal.ActCreateNew:
			action := ACTCreateNewFolder{
				ShouldReServe: true,
			}
			ac.Set(&action, power)
		case internal.ActStay:
			action := ACTStayHere{
				ShouldReServe: false,
			}
			ac.Set(&action, power)
		}
	}
	return nil
}

//----- Default -----

var ClsDefault = CustomCls{
	Name: "Default",
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
	Brief:    "this is a Default Classifier for General Payloads containing any topic!",
	AddRoots: true,
	Criteria: map[string]CstmCriteria{
		"CreateNewFolder": {
			Description: `No existing child is a suitable
match, but the file represents a distinct subcategory of the current
directory that deserves its own folder. Consider the existing sibling
categories and their level of abstraction. A new programming language
may justify a sibling folder alongside existing language-specific folders.
The topic must fit within the current directory, not be unrelated to it.
Do not require multiple existing files, but avoid unnecessary folders
for isolated, ordinary files.`,
			Action: internal.ActCreateNew,
		},
		"stayHere": {
			Description: `The current directory is an appropriate
destination, and creating a new subcategory would provide little
organizational benefit. Choose this for general-purpose or isolated files
that do not justify a separate category. Do not choose stay_here merely
because the file belongs to the current directory's broad subject.`,
			Action: internal.ActStay,
		},
		"uncertain": {
			Description: `Use only when the available evidence
is genuinely insufficient to identify a reasonable destination, or when
multiple candidates are similarly suitable. Lack of a perfect child match
is not sufficient reason to choose uncertain if the current directory fits.`,
			Action: internal.ACTUncertain,
		},
	},
}
