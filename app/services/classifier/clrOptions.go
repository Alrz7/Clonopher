package classifier

import (
	"clonopher/app/burrow"
	"clonopher/app/jevkit"
	"clonopher/app/services/composer"
	"clonopher/internal/helpers"
	"maps"
	"slices"
)

type StateOptions struct {
	Name     string
	Stage    *burrow.Stage
	Info     *composer.FileInfo
	Provider *jevkit.Provider
}

func (s *StateOptions) Render() string {
	infoOptions := helpers.InfoOptions{
		&helpers.InfoOption{Info: `Current directory %v Target file: %v`, Options: []any{s.Stage.Path, s.Name}},

		&helpers.InfoOption{Info: `Target file description: %v`, Options: []any{s.Info.Render()}},

		&helpers.InfoOption{Info: `Current directory info: %v`, Options: []any{s.Stage.Info.Render()}},

		&helpers.InfoOption{Info: `Representative files in current directory: %v`, Options: []any{s.Stage.Files}},

		&helpers.InfoOption{Info: `Available child directories: %v`, Options: []any{slices.Sorted(maps.Keys(s.Stage.Roots))}},
	}
	return infoOptions.ConcatIfExists()
}

func (d *DecideOptions) GetCriteria() map[string]string {
	res := make(map[string]string, len(d.Stage.Roots)+3)

	for rootName, root := range d.Stage.Roots {
		infoOptions := helpers.InfoOptions{
			&helpers.InfoOption{Info: `name: %v`, Options: []any{rootName}},
			&helpers.InfoOption{Info: `path: %v`, Options: []any{root.Path}},
			&helpers.InfoOption{Info: `info: %v`, Options: []any{root.Info.Render()}},
			&helpers.InfoOption{Info: `these are the files existing here: %v`, Options: []any{root.Files}},
		}
		info := infoOptions.ConcatIfExists()
		res[rootName] = info
	}

	res["create_new"] = `No existing child matches the file's purpose well enough. The file represents a distinct topic that justifies a new category.`

	res["stay_here"] = `The current directory is the best fit. No existing child is sufficiently relevant, and creating a new category is unnecessary.`

	res["uncertain"] = `Available evidence is insufficient, or the leading candidates are too similar to distinguish reliably. Do not guess.`

	return res
}

// type InstructionOptions struct {

// }

// func (i *InstructionOptions) Render() string{
// 	infoOptions := internal.InfoOptions{
// 	}
// 	infoOptions.ConcatIfExists()
// }
