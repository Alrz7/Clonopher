package burrow

import (
	"io/fs"
	"maps"
	"path/filepath"
)

func (s *Directory) Explore() error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if s.Explored.Load() {
		return nil
	}

	err := filepath.WalkDir(s.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
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
				s.addRootLocked(d.Name(), root)
				return filepath.SkipDir
			} else {
				err = s.readHelperFileLocked()
				if err != nil {
					return err
				}
			}
		} else {
			s.addFileLocked(d.Name())
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.Explored.Store(true)
	return nil
}

/*
Pre-Explores an Directory, its mainly used for Globals(Heads)
depth == 0 --> only explores Head itself
depth < 0  --> explores till end (every sub-Directories)
*/
func (d *Directory) ExploreGlobal(depth int) error {
	exploreIfNot := func() error {
		if !d.Explored.Load() {
			err := d.Explore()
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := exploreIfNot()
	if err != nil {
		return err
	}

	d.Mu.RLock()
	if depth == 0 {
		d.Mu.RUnlock()
		return nil
	}
	var roots = map[string]*Directory{}
	maps.Copy(roots, d.Children)
	d.Mu.RUnlock()

	for _, root := range roots {
		if root == nil {
			continue
		}
		err := root.ExploreGlobal(depth - 1)
		if err != nil {
			return err
		}
	}
	return nil
}
