package burrow

import (
	"clonopher/app/providers"
	"clonopher/internal"
	"clonopher/internal/helpers"
	"clonopher/internal/loggy"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/spf13/fileflow"
	"gopkg.in/yaml.v3"
)

type Stage struct {
	Mu    *sync.Mutex
	Path  string            `yaml:"path"`
	Roots map[string]*Stage `yaml:"roots"`
	Files []string          `yaml:"existingFiles"`
	Info  *StageInfo        `yaml:"info"`
}

type InfoProvider struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

type StageInfo struct {
	StageID   string        `yaml:"stage_id"`
	Path      string        `yaml:"path"`
	CreatedAt time.Time     `yaml:"created_at"`
	UpdatedAt time.Time     `yaml:"updated_at"`
	Version   int           `yaml:"version"`
	Brief     string        `yaml:"brief"`
	Creator   *InfoProvider `yaml:"creator,omitempty"`
	Modifier  *InfoProvider `yaml:"modifier,omitempty"`
	// Description string
	// RelatedWords []string
}

func New(path string) *Stage {
	return &Stage{
		Path:  path,
		Roots: map[string]*Stage{},
		Mu:    &sync.Mutex{},
	}
}

// new with options...

func (s *Stage) Explore() error {
	err := filepath.WalkDir(s.Path, func(path string, d fs.DirEntry, err error) error {
		isDir := d.IsDir()
		if isDir {
			if ShouldSkip(d.Name()) {
				return filepath.SkipDir
			}
			if path != s.Path {
				root := New(path)
				err = root.ReadHelperFile()
				if err != nil {
					return err
				}
				s.AddRoot(d.Name(), root)
				return filepath.SkipDir
			} else {
				err = s.ReadHelperFile()
				if err != nil {
					return err
				}
			}
		} else {
			s.AddFile(d.Name())
		}
		return nil
	})
	return err
}

func (s *Stage) ReadHelperFile() error {
	if fileflow.Exists(filepath.Join(s.Path, internal.HelperFileName)) {
		b, err := os.ReadFile(filepath.Join(s.Path, internal.HelperFileName))
		if err != nil {
			return err
		}
		var inf StageInfo
		err = yaml.Unmarshal(b, &inf)
		if err != nil {
			return err
		}

		if filepath.Clean(inf.Path) != filepath.Clean(s.Path) {
			loggy.Warnf("helper file's path <%v> \n does NOT match the stage its placed in <%v> \n", inf.Path, s.Path)
		}
		s.SetInfo(&inf)
	}
	return nil
}

func (s *Stage) GuessStageInfo(provider *providers.Provider) error {
	loggy.Info("generating Info...")
	feed := helpers.InfoOptions{
		&helpers.InfoOption{Info: `dir-path: %v`, Options: []any{s.Path}},
		&helpers.InfoOption{Info: `sub-dirs: %v`, Options: []any{slices.Collect(maps.Keys(s.Roots))}},
		&helpers.InfoOption{Info: `dir-files: %v`, Options: []any{s.Files}},
	}.ConcatIfExists()

	query := `You are Clonopher's directory metadata generator.

Your task is to produce a concise, factual brief describing the
current directory for use by a separate file-classification model.

Rules:
- Output only the final directory brief.
- Never output your analysis, reasoning, plans, or conversational text.
- Write 35–65 words in plain text.
- Explain the directory's purpose and organization using the supplied context.
- Treat every subdirectory name as a separate directory.
- Do not invent files, folders, file contents, or filesystem changes.
- Do not recommend actions or discuss where a file should be moved.
- If evidence is limited, describe only what the evidence supports.
- Do not include a heading, bullet points, or Markdown.
`

	req := providers.LlmRequest{
		Provider: provider,
		Model:    provider.ActiveModel,
		Stream:   false,
		Messages: []providers.LlmMessage{
			{Role: "system", Content: query},
			{Role: "user", Content: feed},
		},
	}
	sctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	resp, err := req.SendTo(sctx, req.Model)
	if err != nil {
		return err
	}
	respContent, err := helpers.ExtractContent(*resp)
	if err != nil {
		return err
	}
	s.Info = &StageInfo{Path: s.Path, CreatedAt: time.Now(), UpdatedAt: time.Now(), Brief: respContent, Modifier: &InfoProvider{Provider: provider.Name, Model: req.Model}}
	loggy.Infof("Info generated! <%v...>", respContent)
	return s.WriteNewHelperFile(s.Info)
}

func (s *Stage) UpdateHelperFile() error {
	b, err := yaml.Marshal(s.Info)
	if err != nil {
		return err
	}
	header := []byte("# This file is Managed by Clonopher. Manual edits may be overwritten.\n\n")
	content := append(header, b...)
	err = os.WriteFile(filepath.Join(s.Path, internal.HelperFileName), content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *Stage) WriteNewHelperFile(inf *StageInfo) error {
	sum := sha256.Sum256([]byte(s.Path + time.Now().String()))
	id := hex.EncodeToString(sum[:])
	inf.StageID = id
	s.SetInfo(inf)
	b, err := yaml.Marshal(inf)
	if err != nil {
		return err
	}
	header := []byte("# This file is Managed by Clonopher. Manual edits may be overwritten.\n\n")
	content := append(header, b...)
	err = s.Submit()
	if err != nil {
		return err
	}
	path := filepath.Join(s.Path, internal.HelperFileName)
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *StageInfo) Render() string {
	if s == nil {
		return ""
	}
	infoOptions := helpers.InfoOptions{
		&helpers.InfoOption{Info: `brief description: %v`, Options: []any{s.Brief}},
	}
	return infoOptions.ConcatIfExists()
}
