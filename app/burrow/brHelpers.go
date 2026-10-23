package burrow

import (
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