package classifiers

import (
	"clonopher/internal"
	"clonopher/internal/helpers"
	"maps"
	"slices"
)

func (c *Case) GetState() string {
	infoOptions := helpers.InfoOptions{
		&helpers.InfoOption{Info: `Current directory %v Target file: %v`, Options: []any{c.Stage.Path, c.Name}},

		&helpers.InfoOption{Info: `File Info: %v`, Options: []any{c.Info.Render()}},

		&helpers.InfoOption{Info: `Current directory info: %v`, Options: []any{c.Stage.Info.Render()}},

		&helpers.InfoOption{Info: `Representative files in current directory: %v`, Options: []any{c.Stage.Files}},

		&helpers.InfoOption{Info: `Available child directories: %v`, Options: []any{slices.Sorted(maps.Keys(c.Stage.Roots))}},
	}
	return infoOptions.ConcatIfExists()
}

func (c *Case) GetCriteria() map[string]string {
	res := make(map[string]string, len(c.Stage.Roots)+3)

	for rootName, root := range c.Stage.Roots {
		infoOptions := helpers.InfoOptions{
			&helpers.InfoOption{Info: `name: %v`, Options: []any{rootName}},
			&helpers.InfoOption{Info: `path: %v`, Options: []any{root.Path}},
			&helpers.InfoOption{Info: `info: %v`, Options: []any{root.Info.Render()}},
			&helpers.InfoOption{Info: `these are the files existing here: %v`, Options: []any{root.Files}},
		}
		info := infoOptions.ConcatIfExists()
		res[rootName] = info
	}

	res[internal.ClrCreateNewFolder] = `No existing child matches the file's purpose well enough. The file represents a distinct topic that justifies a new category.`

	res[internal.ClrStayHere] = `The current directory is the best fit. No existing child is sufficiently relevant, and creating a new category is unnecessary.`

	res[internal.ClrUncertain] = `Available evidence is insufficient, or the leading candidates are too similar to distinguish reliably. Do not guess.`

	return res
}

// type InstructionOptions struct {

// }

// func (i *InstructionOptions) Render() string{
// 	infoOptions := internal.InfoOptions{
// 	}
// 	infoOptions.ConcatIfExists()
// }
