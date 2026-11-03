package fileSystem

import (
	"clonopher/app/config"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/fileflow"
	"go.yaml.in/yaml/v3"
)

type StagingStat string

const (
	StgIdle   StagingStat = "idle"
	StgAction StagingStat = "action"
	StgDone   StagingStat = "Done"
)
const StagingFileName = "StagingProfiles.yaml"

type Staging struct {
	Mu       *sync.RWMutex `yaml:"-" json:"-"`
	ID       string        `yaml:"id" json:"id"`
	FileName string        `yaml:"file_name" json:"fileName"`
	Src      string        `yaml:"src" json:"src"`
	Status   StagingStat   `yaml:"status" json:"status"`
}

type Stdb map[string]*Staging

var GlobalStagings = Stdb{}
var GlMu = sync.RWMutex{}

func (db *Stdb) LoadExistingStagings() error {
	path := filepath.Join(config.Default.App.AppConfigDir, StagingFileName)
	ext := fileflow.Exists(path)
	if !ext {
		return nil
	}
	ok, err := IsFile(path)
	if !ok {
		if err != nil {
			return err
		}
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, db)
	if err != nil {
		return err
	}
	for _, st := range *db {
		st.Mu = &sync.RWMutex{}
	}
	return nil
}

func (db *Stdb) AddNewStaging(name, src string) error {
	sum := sha256.Sum256([]byte(name + src + time.Now().String()))
	id := hex.EncodeToString(sum[:])

	status := &Staging{
		Mu:       &sync.RWMutex{},
		ID:       id,
		FileName: name,
		Src:      src,
		Status:   StgIdle,
	}

	GlMu.Lock()
	defer GlMu.Unlock()
	(*db)[id] = status
	return db.WriteStgs()
}

func (db *Stdb) GetById(id string) *Staging {
	GlMu.Lock()
	defer GlMu.Unlock()
	return (*db)[id]
}

func (db *Stdb) DeleteById(id string) error {
	staging := db.GetById(id)
	if staging == nil {
		return nil
	}

	GlMu.Lock()
	defer GlMu.Unlock()
	delete(*db, id)
	err := db.WriteStgs()
	if err != nil {
		return err
	}

	staging.Mu.Lock()
	staging.Status = StgDone
	staging.Mu.Unlock()

	return nil
}

func (db *Stdb) GetIdleStagings() map[string]Staging {
	res := map[string]Staging{}
	GlMu.RLock()
	for id, st := range *db {
		st.Mu.RLock()
		if st.Status == StgIdle {
			res[id] = *st
		}
		st.Mu.RUnlock()
	}
	defer GlMu.RUnlock()
	return res
}

func (db *Stdb) WriteStgs() error {
	b, err := yaml.Marshal(db)
	if err != nil {
		return err
	}
	header := []byte("# This file is Managed by Clonopher. please do NOT modify!.\n\n")
	content := append(header, b...)

	path := filepath.Join(config.Default.App.AppConfigDir, StagingFileName)
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
