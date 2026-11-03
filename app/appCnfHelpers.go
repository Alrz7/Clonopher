package app

import (
	"clonopher/app/burrow"
	"clonopher/app/config"
	"clonopher/app/core"
	"clonopher/internal"
	"clonopher/internal/fileSystem"
	"clonopher/internal/loggy"
	"crypto/rand"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

func ParseRoots(roots string) error {
	if roots != "" {
		res := map[string]string{}

		for pair := range strings.SplitSeq(roots, ",") {
			pair = strings.TrimSpace(pair)
			root := strings.SplitN(pair, ":", 2)
			if len(root) != 2 {
				return loggy.Say("uncorrect format for roots!")
			}
			key := strings.TrimSpace(root[0])
			val := strings.TrimSpace(root[1])
			if key == "" || val == "" {
				continue
			}
			res[key] = val
		}

		for key, val := range res {
			ok, err := fileSystem.IsDir(val)
			if ok {
				config.Default.App.Roots[key] = val
			} else {
				if err != nil {
					return err
				}
				return loggy.Sayf("path <%v> is not defined or accessible!", val)
			}
		}
	}

	for root, path := range config.Default.App.Roots {
		_, err := burrow.NewGlobal(root, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseSources(sources, infos, root, actionPolicy string) ([]*core.Payload, error) {
	if strings.TrimSpace(sources) == "" {
		return nil, nil
	}

	srcGroups, err := parseSourceGroups(sources)
	if err != nil {
		return nil, err
	}

	var infList []string

	if strings.TrimSpace(infos) != "" {
		infList = strings.Split(infos, ",")

		for i := range infList {
			infList[i] = strings.TrimSpace(infList[i])

			if infList[i] == "-" {
				infList[i] = ""
			}
		}
	}

	if len(infList) > 1 && len(infList) != len(srcGroups) {
		return nil, loggy.Sayf(
			"sources & infos do not pair on count: %d sources, %d infos",
			len(srcGroups), len(infList),
		)
	}

	var res []*core.Payload

	for i, srcs := range srcGroups {
		desc := ""

		if len(infList) == 1 {
			desc = infList[0]
		} else if len(infList) > 1 {
			desc = infList[i]
		}

		name := filepath.Base(srcs[0])
		nameRunes := []rune(name)

		if len(nameRunes) > 35 {
			name = string(nameRunes[:35]) + "..."
		}

		var assets []*fileSystem.Asset
		for _, src := range srcs {
			asset := fileSystem.Asset{
				Mu:       &sync.RWMutex{},
				ID:       rand.Text(),
				FileName: name,
				Src:      src,
				Status:   fileSystem.StgActive,
			}
			assets = append(assets, &asset)
		}

		p, err := core.New(
			&config.Default,
			name,
			root,
			assets,
			desc,
			actionPolicy,
			internal.CliMode,
		)
		if err != nil {
			return nil, err
		}

		if p != nil {
			res = append(res, p)
		}
	}

	return res, nil
}

func parseSourceGroups(sources string) ([][]string, error) {
	var (
		groups     [][]string
		group      []string
		current    strings.Builder
		inGroup    bool
		afterGroup bool
	)

	for _, char := range sources {
		switch char {
		case '(':
			if inGroup || afterGroup || strings.TrimSpace(current.String()) != "" {
				return nil, loggy.Say("invalid source group")
			}

			current.Reset()
			group = make([]string, 0, 2)
			inGroup = true

		case ')':
			if !inGroup {
				return nil, loggy.Say("unexpected closing parenthesis")
			}

			path := strings.TrimSpace(current.String())
			if path == "" {
				return nil, loggy.Say("empty source path in group")
			}

			group = append(group, path)
			groups = append(groups, group)

			group = nil
			current.Reset()
			inGroup = false
			afterGroup = true

		case ',':
			if inGroup {
				path := strings.TrimSpace(current.String())
				if path == "" {
					return nil, loggy.Say("empty source path in group")
				}

				group = append(group, path)
				current.Reset()
				continue
			}

			if afterGroup {
				afterGroup = false
				current.Reset()
				continue
			}

			path := strings.TrimSpace(current.String())
			if path == "" {
				return nil, loggy.Say("empty source path")
			}

			groups = append(groups, []string{path})
			current.Reset()

		default:
			if afterGroup {
				if unicode.IsSpace(char) {
					continue
				}

				return nil, loggy.Say("missing comma after source group")
			}

			current.WriteRune(char)
		}
	}

	if inGroup {
		return nil, loggy.Say("unclosed source group")
	}

	if afterGroup {
		return groups, nil
	}

	path := strings.TrimSpace(current.String())
	if path == "" {
		return nil, loggy.Say("empty source path")
	}

	groups = append(groups, []string{path})

	return groups, nil
}
