package fileSystem

import (
	"clonopher/app/config"
	"clonopher/internal/loggy"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/spf13/fileflow"
	"go.yaml.in/yaml/v3"
)

type StagingStat string

const (
	StgIdle   StagingStat = "idle"
	StgActive StagingStat = "active"
	StgDone   StagingStat = "Done"
)
const StagingFileName = "StagingProfiles.yaml"

type Asset struct {
	Mu       *sync.RWMutex `yaml:"-" json:"-"`
	ID       string        `yaml:"id" json:"id"`
	FileName string        `yaml:"file_name" json:"fileName"`
	Src      string        `yaml:"src" json:"src"`
	Status   StagingStat   `yaml:"status" json:"status"`
}

type Stdb map[string]*Asset

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

	status := &Asset{
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

func (db *Stdb) GetById(id string) (*Asset, bool) {
	GlMu.Lock()
	defer GlMu.Unlock()
	res, ok := (*db)[id]
	return res, ok
}

func (db *Stdb) DeleteById(id string) error {
	staging, ok := db.GetById(id)
	if !ok {
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

func (db *Stdb) GetIdleStagings() map[string]Asset {
	res := map[string]Asset{}
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

type Assets []*Asset

func (assets Assets) RollbackStats() {
	for _, file := range assets {
		file.Mu.Lock()
		if file.Status == StgActive {
			file.Status = StgIdle
		}
		file.Mu.Unlock()
	}
}

func ParseAssets(stagings []string) (Assets, error) {
	assets := Assets{}

	for _, st := range stagings {
		file, ok := GlobalStagings.GetById(st)
		if !ok {
			return assets, loggy.Sayf("id %v vas invalid, no Staging file with this id was found!", st)

		}

		if slices.Contains(assets, file) {
			continue
		}

		file.Mu.Lock()
		if file.Status == StgIdle {
			ok, err := IsFile(file.Src)
			if !ok {
				appErr := loggy.Sayf("something went wrong with <%v>", file.Src)
				if err != nil {
					appErr = loggy.Sayf("something went wrong with <%v>, err: %v", file.Src, err.Error())
				}
				loggy.AddNew(appErr.Error(), loggy.PdLError)
				file.Mu.Unlock()
				return assets, appErr
			}
			assets = append(assets, file)
			file.Status = StgActive
		} else {
			file.Mu.Unlock()
			return assets, loggy.Sayf("there was a Conflict while selecting Asset <%v>", file.Src)
		}
		file.Mu.Unlock()
	}
	return assets, nil
}
