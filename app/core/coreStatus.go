package core

import (
	"clonopher/app/burrow"
	"sort"
	"sync"
)

type PayloadState string

const (
	PayloadStateQueued               PayloadState = "queued"
	PayloadStateExploring            PayloadState = "exploring"
	PayloadStateClassifying          PayloadState = "classifying"
	PayloadStateGenerating           PayloadState = "generating"
	PayloadStateModifying            PayloadState = "modifying"
	PayloadStateWaiting              PayloadState = "waiting"
	PayloadStateRetrying             PayloadState = "retrying"
	PayloadStateInferring            PayloadState = "inferring"
	PayloadStateSelectingClassifiers PayloadState = "selecting_classifiers"

	PayloadStateUncertain PayloadState = "uncertain"
	PayloadStateCompleted PayloadState = "completed"
	PayloadStateFailed    PayloadState = "failed"
	PayloadStateCancelled PayloadState = "cancelled"
)

type PayloadStatus struct {
	Mu         *sync.RWMutex   `json:"-"`
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Stage      *burrow.Stage   `json:"-"`
	Road       []*burrow.Stage `json:"-"`
	State      PayloadState    `json:"state"`
	Iteration  int             `json:"iteration"`
	LastAction string          `json:"lastAction"`
	Actions    []string        `json:"actions"`
}

func (pd *Payload) ShiftStage(stage *burrow.Stage) {
	pStat := Stats.Get(pd.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.Road = append(pStat.Road, pd.Stage)
	pStat.Stage = stage
}

func (pd *Payload) ShiftAction(action string) {
	pStat := Stats.Get(pd.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.Actions = append(pStat.Actions, action)
	pStat.LastAction = action
}

func (pd *Payload) SetState(state PayloadState) {
	pStat := Stats.Get(pd.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.State = state
}

func (pd *Payload) AddIteration() {
	pStat := Stats.Get(pd.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.Iteration++
}

type stdb map[string]*PayloadStatus

var Stats = stdb{}
var dbMu = sync.RWMutex{}

func (db *stdb) AddNew(id, name string, stage *burrow.Stage) {
	status := &PayloadStatus{
		Mu:    &sync.RWMutex{},
		ID:    id,
		Name:  name,
		State: PayloadStateQueued,
		Stage: stage,
	}

	dbMu.Lock()
	defer dbMu.Unlock()
	(*db)[id] = status
}

func (db *stdb) Get(id string) *PayloadStatus {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return (*db)[id]
}

// Snapshot returns the registered status pointers in a stable order.
// Callers must lock each status before reading or changing its fields.
func (db *stdb) Snapshot() []*PayloadStatus {
	dbMu.RLock()
	statuses := make([]*PayloadStatus, 0, len(*db))
	for _, status := range *db {
		statuses = append(statuses, status)
	}
	dbMu.RUnlock()

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ID < statuses[j].ID
	})
	return statuses
}
