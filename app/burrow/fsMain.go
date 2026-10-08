package burrow

import (
	"clonopher/app/services/composer"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/fileflow"
)

type Stage struct {
	Path  string              `yaml:"path"`
	Roots map[string]*Stage   `yaml:"roots"`
	Files []string            `yaml:"existingFiles"`
	Info  *composer.StageInfo `yaml:"info"`
}

func (s *Stage) Explore() error {
	err := filepath.WalkDir(s.Path, func(path string, d fs.DirEntry, err error) error {
		isDir := d.IsDir()
		if isDir {
			if path != s.Path {
				s.Roots[d.Name()] = &Stage{Path: path}
				return filepath.SkipDir
			}
		} else {
			s.Files = append(s.Files, d.Name())
		}
		return nil
	})
	return err
}

func (s *Stage) DropFileHere(src string) (string, error) {
	return fileflow.Move(src, s.Path)
}

func (s *Stage) CopyFileHere(src string) (string, error) {
	return fileflow.Copy(src, s.Path)
}

func (s *Stage) SubmitDropFileHere(src string) (string, error) {
	sErr := s.Submit()
	if sErr != nil {
		return "", sErr
	}
	finalPath, dErr := s.DropFileHere(src)
	return finalPath, dErr
}

func (s *Stage) SubmitCopyFileHere(src string) (string, error) {
	sErr := s.Submit()
	if sErr != nil {
		return "", sErr
	}
	finalPath, cErr := s.CopyFileHere(src)
	return finalPath, cErr
}

func (s *Stage) AddRoot(join string) *Stage {
	path := filepath.Join(s.Path, join)
	newRoot := Stage{
		Path:  path,
		Roots: make(map[string]*Stage),
		Info:  &composer.StageInfo{},
	}
	return &newRoot
}

func (s *Stage) Submit() error {
	return os.MkdirAll(s.Path, 0755)
}
