package fileSystem

import (
	"clonopher/app/config"
	"clonopher/internal/loggy"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/Bios-Marcel/wastebasket/v2"
	"github.com/spf13/fileflow"
	"go.yaml.in/yaml/v3"
)

const (
	StgIdle   AssetStat = "idle"
	StgActive AssetStat = "active"
	StgDone   AssetStat = "Done"
	StgClean  AssetStat = "Clean"
)
const StagingFileName = "StagingProfiles.yaml"

type Asset struct {
	Mu       *sync.RWMutex `yaml:"-" json:"-"`
	ID       string        `yaml:"id" json:"id"`
	FileName string        `yaml:"file_name" json:"fileName"`
	Src      string        `yaml:"src" json:"src"`
	Status   AssetStat     `yaml:"status" json:"status"`
}

type Stdb map[string]*Asset

var GlobalStagings = Stdb{}
var GlMu = sync.RWMutex{}

type AssetStat string

type AssetStats []AssetStat

func (stats AssetStats) Contains(target AssetStat) bool {
	for _, stat := range stats {
		if stat == target {
			return true
		}
	}
	return false
}

func (db *Stdb) LoadAssetConfig() error {
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

	c := false
	for _, asset := range *db {
		asset.Mu = &sync.RWMutex{}
		if asset.Status == StgActive {
			asset.Status = StgIdle
			if c == false {
				c = true
			}
		}
	}

	if c == true {
		err = db.UpdateConfig()
		if err != nil {
			return err
		}
	}

	for id, asset := range *db {
		err = db.LockedCleanByCon(id, asset, []AssetStat{StgClean})
		if err != nil {
			return err
		}
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
	return db.UpdateConfig()
}

func (db *Stdb) GetById(id string) (*Asset, bool) {
	GlMu.RLock()
	defer GlMu.RUnlock()
	res, ok := (*db)[id]
	return res, ok
}

func (db *Stdb) LockedCleanByCon(id string, asset *Asset, cons AssetStats) error {
	asset.Mu.Lock()
	status := asset.Status
	src := asset.Src

	con := cons.Contains(status)

	if con {
		asset.Status = StgClean
		asset.Mu.Unlock()

		err := db.UpdateConfig()
		if err != nil {
			return err
		}

		_, err = os.Stat(src)

		switch {
		case err == nil:
			if err := wastebasket.Trash(src); err != nil {
				return loggy.Sayr("failed to move asset to trash: %w", err)
			}
		case errors.Is(err, fs.ErrNotExist):
			break
		default:
			return err
		}

		delete(*db, id)

		err = db.UpdateConfig()
		if err != nil {
			return err
		}

		return nil

	} else {
		asset.Mu.Unlock()
		// return loggy.Sayf("Asset's Status does not pass the Condition!, status: %v", status)
		return nil
	}
}

func (db *Stdb) DeleteById(id string) error {
	asset, ok := db.GetById(id)
	if !ok {
		return nil
	}

	GlMu.Lock()
	defer GlMu.Unlock()

	err := db.LockedCleanByCon(id, asset, []AssetStat{StgDone, StgClean, StgIdle})
	if err != nil {
		return err
	}

	for asId, as := range *db {
		err = db.LockedCleanByCon(asId, as, []AssetStat{StgClean})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Asset) SetAsDone() {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.Status = StgDone
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

func (db *Stdb) UpdateConfig() error {
	snapShot := Stdb{}

	for id, asset := range *db {
		asset.Mu.RLock()
		snp := *asset
		snp.Mu = nil
		snapShot[id] = &snp
		asset.Mu.RUnlock()
	}

	b, err := yaml.Marshal(snapShot)
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
