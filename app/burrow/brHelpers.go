package burrow

import (
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/fileflow"
)

// Fs-Helpers
func (d *Directory) DropFileHere(src string) (string, error) {
	dst := filepath.Join(d.Path, filepath.Base(src))
	return fileflow.Move(src, dst)
}

func (d *Directory) CopyFileHere(src string) (string, error) {
	dst := filepath.Join(d.Path, filepath.Base(src))
	return fileflow.Copy(src, dst)
}
func (d *Directory) SubmitDropFileHere(src string) (string, error) {
	sErr := d.Submit()
	if sErr != nil {
		return "", sErr
	}
	finalPath, dErr := d.DropFileHere(src)
	return finalPath, dErr
}

func (d *Directory) SubmitCopyFileHere(src string) (string, error) {
	sErr := d.Submit()
	if sErr != nil {
		return "", sErr
	}
	finalPath, cErr := d.CopyFileHere(src)
	return finalPath, cErr
}

func (d *Directory) Submit() error {
	return os.MkdirAll(d.Path, 0755)
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

func (d *Directory) setRoots(roots map[string]*Directory) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	d.setRootsLocked(roots)
}

// *-Locked helpers require the caller to hold s.Mu.

func (d *Directory) setRootsLocked(roots map[string]*Directory) {
	d.Children = roots
}

func (d *Directory) AddRoot(rootName string, root *Directory) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	d.addRootLocked(rootName, root)
}

func (d *Directory) addRootLocked(rootName string, root *Directory) {
	// TODO: If the root already exists, merge both DirInfo values
	// through the LLM instead of blindly replacing the existing Directory.
	d.Children[rootName] = root
}

func (d *Directory) setFiles(files []string) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	d.setFilesLocked(files)
}

func (d *Directory) setFilesLocked(files []string) {
	d.Files = files
}

func (d *Directory) addFile(file string) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	d.addFileLocked(file)
}

func (d *Directory) addFileLocked(file string) {
	d.Files = append(d.Files, file)
}

func (d *Directory) SetInfo(info *DirInfo) {
	d.Mu.Lock()
	defer d.Mu.Unlock()

	d.setInfoLocked(info)
}

func (d *Directory) setInfoLocked(info *DirInfo) {
	d.Info = info
}

type SnapShot struct {
	Path     string               `json:"path"`
	Children map[string]*SnapShot `json:"children"`
	Files    []string             `json:"existingFiles"`
}

func (d *Directory) SnapShot(depth int) *SnapShot {
	take := func(dir *Directory) *SnapShot {
		var out = SnapShot{
			Path:     dir.Path,
			Children: map[string]*SnapShot{},
			Files:    make([]string, len(dir.Files)),
		}
		copy(out.Files, dir.Files)
		return &out
	}

	d.Mu.RLock()
	res := take(d)
	if depth == 0 {
		d.Mu.RUnlock()
		return res
	}
	var roots = map[string]*Directory{}
	maps.Copy(roots, d.Children)
	d.Mu.RUnlock()

	for rootName, root := range roots {
		if root == nil {
			continue
		}
		res.Children[rootName] = root.SnapShot(depth - 1)
	}
	return res
}
