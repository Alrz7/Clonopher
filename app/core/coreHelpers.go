package core

import (
	"clonopher/app/burrow"
	"clonopher/app/providers"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"context"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func (c *ClsConfig) LoadExistingClassifiers(dir string) error {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			name := d.Name()
			if strings.Contains(name, ".yaml") {
				b, err := os.ReadFile(path)
				if err != nil {
					loggy.Fatal(err)
				}
				var newLoad ClsConfig
				err = yaml.Unmarshal(b, &newLoad)
				if err != nil {
					return err
				}
				for clsName, cls := range newLoad.Classifiers {
					if _, ok := c.Classifiers[name]; ok {
						loggy.Warnf("a classifier called '%v' already exists!, rename the second one here in [%v] to prevent owerWrite!", name, path)
					}
					c.Classifiers[clsName] = cls
				}
			}
		}
		return nil
	})
	return err
}

func (p *Payload) SelectRelevantClassifiers(c *ClsConfig) (*releventCls, error) {
	state := p.GetState()
	req := providers.JevRequest{
		Provider:  p.JevProvider,
		Model:     p.JevProvider.ActiveModel,
		State:     state,
		Questions: map[string]providers.JevQuestion{},
	}

	for clsTitle, cls := range c.Classifiers {
		if cls.Name == ClsMain.Name {
			continue
		}
		req.Questions[clsTitle] = providers.JevQuestion{
			Type: providers.JevQANoul,
			Instructions: fmt.Sprintf(
				"Is classifier %q relevant to classifying the current payload?\nBrief: %s",
				clsTitle,
				cls.Brief,
			), Criteria: map[string]string{
				"true":  "yes this classifier is relevent and can be used for this payload",
				"false": "No this classifier is NOT relevent and can NOT be used for this payload",
			},
		}
	}
	sctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	resp, err := req.Send(sctx)
	if err != nil {
		return nil, err
	}

	var relvs releventCls
	for clsTitle, res := range resp.Answers {
		cls, ok := c.Classifiers[clsTitle]
		if !ok || cls == nil {
			return nil, loggy.Sayf("classifier '%v' was not defined", clsTitle)
		}
		if cls.Name == ClsMain.Name {
			continue
		}
		if res.Noul == nil {
			return nil, loggy.Sayf("nil answer for classifier: '%v'", clsTitle)
		}
		if *res.Noul < 0.65 {
			continue
		}
		relvs = append(relvs, struct {
			Cls   *CustomCls
			Power float64
		}{Cls: cls, Power: *res.Noul})
	}

	if len(relvs) == 0 {
		var fallback = struct {
			Cls   *CustomCls
			Power float64
		}{Power: 1.0}

		customDefault, ok := c.Classifiers[ClsMain.Name]
		if ok {
			fallback.Cls = customDefault
		} else {
			fallback.Cls = &ClsMain
		}
		relvs = append(relvs, fallback)
	}

	return &relvs, nil
}

func (p *Payload) GetState() string {
	infoOptions := helpers.InfoOptions{
		&helpers.InfoOption{Info: `Current directory %v Target file: %v`, Options: []any{p.Stage.Path, p.Name}},

		&helpers.InfoOption{Info: `File Info: %v`, Options: []any{p.Info.Render()}},

		&helpers.InfoOption{Info: `Current directory info: %v`, Options: []any{p.Stage.Info.Render()}},

		&helpers.InfoOption{Info: `Representative files in current directory: %v`, Options: []any{p.Stage.Files}},

		&helpers.InfoOption{Info: `Available child directories: %v`, Options: []any{slices.Sorted(maps.Keys(p.Stage.Roots))}},
	}
	return infoOptions.ConcatIfExists()
}

func (p *Payload) SetNext(next *burrow.Stage) {
	p.Road = append(p.Road, p.Stage)
	p.Stage = next
}
