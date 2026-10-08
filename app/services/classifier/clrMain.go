package classifier

import (
	"clonopher/app/burrow"
	"clonopher/app/jevkit"
	"clonopher/app/services/composer"
	"fmt"

	"go.yaml.in/yaml/v3"
)

type DecideOptions struct {
	Name     string
	Stage    *burrow.Stage
	Info     *composer.FileInfo
	Provider *jevkit.Provider
}

// AI-facing representation of the current decision context.
type stageSnapshot struct {
	Path          string   `yaml:"currentPath"`
	Description   string   `yaml:"description,omitempty"`
	ExistingFiles []string `yaml:"existingFiles,omitempty"`
}

type candidateSnapshot struct {
	Path          string   `yaml:"path"`
	Description   string   `yaml:"description,omitempty"`
	ExistingFiles []string `yaml:"existingFiles,omitempty"`
}

type decisionSnapshot struct {
	Target     stageSnapshot                `yaml:"target"`
	Candidates map[string]candidateSnapshot `yaml:"candidates"`
}

func (d *DecideOptions) NewRequest() (*jevkit.Request, error) {
	state, err := d.buildState()
	if err != nil {
		return nil, err
	}

	criteria, err := d.getCriteria()
	if err != nil {
		return nil, err
	}

	instruction := fmt.Sprintf(`
Decide what should happen to the file (%v) in the current filesystem stage.

Evaluate the file against the available child stages using semantic meaning,
context, descriptions, and existing contents.

Prefer the most specific existing child when there is a strong semantic match.
Do not choose a child merely because its name contains a matching keyword.

Choose "create_new" when none of the existing children represents the file
well enough and a new semantic branch is justified.

Choose "stay_here" when the current stage itself is the appropriate location
and moving deeper would be unnecessary.

Choose "uncertain" when the available information is insufficient or the
classification is genuinely ambiguous.

Return exactly one option.
`, d.Name)

	req := jevkit.Request{
		Provider: d.Provider,
		State:    state,
		Questions: map[string]jevkit.Question{
			"action": {
				Type: "choice",

				Instructions: instruction,

				Criteria: criteria,
			},
		},
	}

	return &req, nil
}

func (d *DecideOptions) buildState() (string, error) {
	current := stageSnapshot{
		Path: d.Stage.Path,
	}

	if d.Stage.Info != nil {
		current.Description = d.Stage.Info.Join()
	}

	current.ExistingFiles = d.Stage.Files

	snapshot := decisionSnapshot{
		Target:     current,
		Candidates: make(map[string]candidateSnapshot),
	}

	// fileInfo := ""
	// if d.Info != nil {
	// 	fileInfo = d.Info.Join()
	// }

	// The file being classified is part of the state.
	// We temporarily keep it under target metadata until FileInfo grows further.
	// snapshot.Target.Description = fileInfo

	for name, root := range d.Stage.Roots {
		candidate := candidateSnapshot{
			Path:          root.Path,
			ExistingFiles: root.Files,
		}

		if root.Info != nil {
			candidate.Description = root.Info.Join()
		}

		snapshot.Candidates[name] = candidate
	}

	b, err := yaml.Marshal(snapshot)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (d *DecideOptions) getCriteria() (map[string]string, error) {
	crt := make(map[string]string, len(d.Stage.Roots)+3)

	for rootName, root := range d.Stage.Roots {
		info := ""
		if root.Info != nil {
			info = root.Info.Join()
		}

		b, err := yaml.Marshal(candidateSnapshot{
			Path:          root.Path,
			Description:   info,
			ExistingFiles: root.Files,
		})
		if err != nil {
			return nil, err
		}

		crt[rootName] = string(b)
	}

	crt["create_new"] = `
No existing child stage is a sufficiently strong semantic match.
A new child concept should be created if the supplied file represents a
distinct and meaningful topic that does not fit the existing taxonomy.
`

	crt["stay_here"] = `
The current stage itself is the appropriate semantic location.
Do not create or enter a more specific child stage unless there is a strong
reason to do so.
`

	crt["uncertain"] = `
The available information is insufficient for a reliable decision, or two
or more candidates are too semantically close to distinguish safely.
Prefer uncertainty over a weak or forced classification.
`

	return crt, nil
}
