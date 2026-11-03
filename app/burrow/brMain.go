package burrow

import (
	"clonopher/internal/fileSystem"
	"clonopher/internal/loggy"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Directory struct {
	Mu       *sync.RWMutex         `yaml:"-"`
	InfMu    *sync.RWMutex         `yaml:"-"`
	Explored atomic.Bool           `yaml:"-"`
	Path     string                `yaml:"path"`
	Children map[string]*Directory `yaml:"children"`
	Files    []string              `yaml:"existingFiles"`
	Info     *DirInfo              `yaml:"info"`
}

type InfoProvider struct {
	Provider string `yaml:"provider" json:"provider"`
	Model    string `yaml:"model" json:"model"`
}

type DirInfo struct {
	DirId     string        `yaml:"dir_id" json:"dirId"`
	Path      string        `yaml:"path" json:"path"`
	CreatedAt time.Time     `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time     `yaml:"updated_at" json:"updated_at"`
	Version   int           `yaml:"version" json:"version"`
	Brief     string        `yaml:"brief" json:"brief"`
	Creator   *InfoProvider `yaml:"creator,omitempty" json:"creator,omitempty"`
	Modifier  *InfoProvider `yaml:"modifier,omitempty" json:"modifier,omitempty"`
}

func New(path string) *Directory {
	new := &Directory{
		Path:     path,
		Children: map[string]*Directory{},
		InfMu:    &sync.RWMutex{},
		Mu:       &sync.RWMutex{},
	}
	new.AddToAll()
	return new
}

// new with options...

var All = make(map[string]*Directory)
var AllMu = sync.RWMutex{}

func (d *Directory) AddToAll() {
	AllMu.Lock()
	defer AllMu.Unlock()
	All[d.Path] = d
}

func GetFromAll(path string) (*Directory, bool) {
	AllMu.RLock()
	defer AllMu.RUnlock()
	res, ok := All[path]
	return res, ok
}

// ----< Globals >----
var Globals = make(map[string]*Directory)
var GlbsMu = sync.RWMutex{}

func (d *Directory) SetAsGlobal(title string) error {
	GlbsMu.Lock()
	defer GlbsMu.Unlock()
	if title == "" {
		var err error
		title, err = filepath.Abs(filepath.Clean(d.Path))
		if err != nil {
			return err
		}
	}
	Globals[title] = d
	return nil
}

func NewGlobal(title, path string) (*Directory, error) {
	ok, err := fileSystem.IsDir(path)
	if ok {
		head := New(path)
		err := head.ExploreGlobal(-1)
		if err != nil {
			return nil, err
		}
		err = head.SetAsGlobal(title)
		if err != nil {
			return nil, err
		}
		return head, nil
	} else {
		if err != nil {
			return nil, err
		} else {
			return nil, loggy.Say("head path is not valid, is it a file?!")
		}
	}
}

func GetFromGlobals(title string) (*Directory, bool) {
	GlbsMu.RLock()
	defer GlbsMu.RUnlock()
	res, ok := Globals[title]
	return res, ok
}
