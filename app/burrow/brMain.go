package burrow

import (
	"clonopher/internal"
	"clonopher/internal/helpers"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/fileflow"
	"gopkg.in/yaml.v3"
)

type Stage struct {
	Path  string            `yaml:"path"`
	Roots map[string]*Stage `yaml:"roots"`
	Files []string          `yaml:"existingFiles"`
	Info  *StageInfo        `yaml:"info"`
	Next  *Stage            `yaml:"next"`
}

type StageInfo struct {
	StageID   string    `yaml:"stage_id"`
	Path      string    `yaml:"path"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Version   int       `yaml:"version"`
	Brief     string    `yaml:"brief"`
	// Description string
	// RelatedWords []string
}

func (s *Stage) Explore() error {
	err := filepath.WalkDir(s.Path, func(path string, d fs.DirEntry, err error) error {
		isDir := d.IsDir()
		if isDir {
			if ShouldSkip(d.Name()) {
				return filepath.SkipDir
			}
			if path != s.Path {
				root := Stage{Path: path}
				err = root.ReadHelperFile()
				if err != nil {
					return err
				}
				s.Roots[d.Name()] = &root
				return filepath.SkipDir
			} else {
				err = s.ReadHelperFile()
				if err != nil {
					return err
				}
			}
		} else {
			s.Files = append(s.Files, d.Name())
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
		s.Info = &inf
	}
	return nil
}

func (s *Stage) UpdateHelperFile() error {
	s.Info.UpdatedAt = time.Now()
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

func (s *Stage) WriteNewHelperFile(brief string) error {
	var inf StageInfo
	inf.Path = s.Path
	inf.Brief = brief
	inf.CreatedAt = time.Now()
	inf.UpdatedAt = time.Now()
	sum := sha256.Sum256([]byte(s.Path + time.Now().String()))
	id := hex.EncodeToString(sum[:])
	inf.StageID = id
	s.Info = &inf
	b, err := yaml.Marshal(inf)
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

func (s *StageInfo) Render() string {
	if s == nil {
		return ""
	}
	infoOptions := helpers.InfoOptions{
		&helpers.InfoOption{Info: `brief description: %v`, Options: []any{s.Brief}},
	}
	return infoOptions.ConcatIfExists()
}
