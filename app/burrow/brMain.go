package burrow

import (
	"clonopher/internal"
	"clonopher/internal/loggy"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Stage struct {
	Mu       *sync.RWMutex     `yaml:"-"`
	InfMu    *sync.RWMutex     `yaml:"-"`
	Explored atomic.Bool       `yaml:"-"`
	Path     string            `yaml:"path"`
	Roots    map[string]*Stage `yaml:"substages"`
	Files    []string          `yaml:"existingFiles"`
	Info     *StageInfo        `yaml:"info"`
}

type InfoProvider struct {
	Provider string `yaml:"provider" json:"provider"`
	Model    string `yaml:"model" json:"model"`
}

type StageInfo struct {
	StageID   string        `yaml:"stage_id" json:"stage_id"`
	Path      string        `yaml:"path" json:"path"`
	CreatedAt time.Time     `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time     `yaml:"updated_at" json:"updated_at"`
	Version   int           `yaml:"version" json:"version"`
	Brief     string        `yaml:"brief" json:"brief"`
	Creator   *InfoProvider `yaml:"creator,omitempty" json:"creator,omitempty"`
	Modifier  *InfoProvider `yaml:"modifier,omitempty" json:"modifier,omitempty"`
}

func New(path string) *Stage {
	new := &Stage{
		Path:  path,
		Roots: map[string]*Stage{},
		InfMu: &sync.RWMutex{},
		Mu:    &sync.RWMutex{},
	}
	new.AddToAll()
	return new
}

// new with options...

var All = make(map[string]*Stage)
var AllMu = sync.RWMutex{}

func (s *Stage) AddToAll() {
	AllMu.Lock()
	defer AllMu.Unlock()
	All[s.Path] = s
}

func GetFromAll(path string) (*Stage, bool) {
	AllMu.RLock()
	defer AllMu.RUnlock()
	res, ok := All[path]
	return res, ok
}

// ----< Globals >----
var Globals = make(map[string]*Stage)
var GlbsMu = sync.RWMutex{}

func (s *Stage) SetAsGlobal(title string) error {
	GlbsMu.Lock()
	defer GlbsMu.Unlock()
	if title == "" {
		var err error
		title, err = filepath.Abs(filepath.Clean(s.Path))
		if err != nil {
			return err
		}
	}
	Globals[title] = s
	return nil
}

func NewGlobal(title, path string) (*Stage, error) {
	ok, err := internal.IsDir(path)
	if ok {
		head := New(path)
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

func GetFromGlobals(title string) (*Stage, bool) {
	GlbsMu.RLock()
	defer GlbsMu.RUnlock()
	res, ok := Globals[title]
	return res, ok
}
