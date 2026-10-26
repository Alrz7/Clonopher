package burrow

import (
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/fileflow"
)

// Fs-Helpers
func (s *Stage) DropFileHere(src string) (string, error) {
	dst := filepath.Join(s.Path, filepath.Base(src))
	return fileflow.Move(src, dst)
}

func (s *Stage) CopyFileHere(src string) (string, error) {
	dst := filepath.Join(s.Path, filepath.Base(src))
	return fileflow.Copy(src, dst)
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

// in-memmory helpers

func (s *Stage) setRoots(roots map[string]*Stage) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.setRootsLocked(roots)
}

// *-Locked helpers require the caller to hold s.Mu.

func (s *Stage) setRootsLocked(roots map[string]*Stage) {
	s.Roots = roots
}

func (s *Stage) AddRoot(rootName string, root *Stage) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.addRootLocked(rootName, root)
}

func (s *Stage) addRootLocked(rootName string, root *Stage) {
	// TODO: If the root already exists, merge both StageInfo values
	// through the LLM instead of blindly replacing the existing Stage.
	s.Roots[rootName] = root
}

func (s *Stage) setFiles(files []string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.setFilesLocked(files)
}

func (s *Stage) setFilesLocked(files []string) {
	s.Files = files
}

func (s *Stage) addFile(file string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.addFileLocked(file)
}

func (s *Stage) addFileLocked(file string) {
	s.Files = append(s.Files, file)
}

func (s *Stage) SetInfo(info *StageInfo) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.setInfoLocked(info)
}

func (s *Stage) setInfoLocked(info *StageInfo) {
	s.Info = info
}

type SnpShot struct {
	Path  string              `json:"path"`
	Roots map[string]*SnpShot `json:"substages"`
	Files []string            `json:"existingFiles"`
}

func (s *Stage) SnapShot(depth int) *SnpShot {
	take := func(stage *Stage) *SnpShot {
		var out = SnpShot{
			Path:  stage.Path,
			Roots: map[string]*SnpShot{},
			Files: make([]string, len(stage.Files)),
		}
		copy(out.Files, stage.Files)
		return &out
	}

	s.Mu.RLock()
	res := take(s)
	if depth == 0 {
		s.Mu.RUnlock()
		return res
	}
	var roots = map[string]*Stage{}
	maps.Copy(roots, s.Roots)
	s.Mu.RUnlock()

	for rootName, root := range roots {
		if root == nil {
			continue
		}
		res.Roots[rootName] = root.SnapShot(depth - 1)
	}
	return res
}
