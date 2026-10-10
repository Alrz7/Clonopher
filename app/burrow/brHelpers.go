package burrow

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/fileflow"
)

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
		Info:  &StageInfo{},
	}
	return &newRoot
}

func (s *Stage) Submit() error {
	return os.MkdirAll(s.Path, 0755)
}

func ShouldSkip(dirname string) bool {
	var ignoredDirNames = []string{
		// Version control metadata
		".git",
		".svn",
		".hg",
		".bzr",

		// Application metadata
		".obsidian",

		// Trash directories
		".trash",
		".trash-1000",
		".trashes",
		"$recycle.bin",
		"system volume information",
	}
	return slices.Contains(ignoredDirNames, dirname)
}
