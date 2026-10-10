package classifier

import (
	"clonopher/app/jevkit"
	"clonopher/internal"
)

type DecideOptions struct {
	StateOptions
}

func (d *DecideOptions) NewRequest() (*jevkit.Request, error) {
	state := d.StateOptions.Render()
	criteria := d.GetCriteria()

	instruction := `Choose the best action for the target file using its name, description, and current directory context.
Prefer an existing child when its purpose strongly matches the file. Do not rely on keyword overlap or absolute paths alone.
Choose create_new only when no existing child fits and the file represents a distinct topic worth a new category.
Choose stay_here when the current directory is the appropriate location and no deeper classification is justified.
Choose uncertain when evidence is insufficient or the leading candidates are too similar. Prefer a cautious decision over an unsupported guess.`

	req := jevkit.Request{
		Provider: d.Provider,
		State:    state,
		Questions: map[string]jevkit.Question{
			internal.Action: {
				Type:         "choice",
				Instructions: instruction,
				Criteria:     criteria,
			},
		},
	}

	return &req, nil
}
