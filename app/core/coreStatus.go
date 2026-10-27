package core

import (
	"clonopher/app/burrow"
	"clonopher/internal/loggy"
	"fmt"
	"sort"
	"sync"
	"time"
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


type Log = loggy.PydLog
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
	Logs       []Log

	//api-spec-paramters
	StageString string   `json:"path"`
	RoadString  []string `json:"road"`
}

func (p *Payload) ShiftStage(stage *burrow.Stage) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.Road = append(pStat.Road, p.Stage)
	pStat.Stage = stage
}

func (p *Payload) ShiftAction(action string) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.Actions = append(pStat.Actions, action)
	pStat.LastAction = action
}

func (p *Payload) SetState(state PayloadState) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.State = state
}

func (p *Payload) AddIteration() {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()

	pStat.Iteration++
}

// -----‌< Log >------

func (p *Payload) ShiftNewError(content string) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()
	pStat.Logs = append(pStat.Logs, Log{Type: loggy.PdLError, Content: content, Timestamp: time.Now()})
}
func (p *Payload) ShiftNewErrorf(format string, params ...string) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()
	pStat.Logs = append(pStat.Logs, Log{Type: loggy.PdLError, Content: fmt.Sprintf(format, params), Timestamp: time.Now()})
}
func (p *Payload) ShiftError(err error) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()
	pStat.Logs = append(pStat.Logs, Log{Type: loggy.PdLError, Content: err.Error(), Timestamp: time.Now()})
}

func (p *Payload) ShiftNewInfo(content string) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()
	pStat.Logs = append(pStat.Logs, Log{Type: loggy.PdLInfo, Content: content, Timestamp: time.Now()})
}

func (p *Payload) ShiftNewInfof(format string, params ...string) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()
	pStat.Logs = append(pStat.Logs, Log{Type: loggy.PdLInfo, Content: fmt.Sprintf(format, params), Timestamp: time.Now()})
}

func (p *Payload) ShiftNewWarn(content string) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()
	pStat.Logs = append(pStat.Logs, Log{Type: loggy.PdLWarn, Content: content, Timestamp: time.Now()})
}
func (p *Payload) ShiftNewWarnf(format string, params ...string) {
	pStat := Stats.Get(p.Id)
	pStat.Mu.Lock()
	defer pStat.Mu.Unlock()
	pStat.Logs = append(pStat.Logs, Log{Type: loggy.PdLWarn, Content: fmt.Sprintf(format, params), Timestamp: time.Now()})
}

type Stdb map[string]*PayloadStatus

var Stats = Stdb{}
var StatsMu = sync.RWMutex{}

func (db *Stdb) AddNew(id, name string, stage *burrow.Stage) {
	status := &PayloadStatus{
		Mu:    &sync.RWMutex{},
		ID:    id,
		Name:  name,
		State: PayloadStateQueued,
		Stage: stage,
	}

	StatsMu.Lock()
	defer StatsMu.Unlock()
	(*db)[id] = status
}

func (db *Stdb) Get(id string) *PayloadStatus {
	StatsMu.RLock()
	defer StatsMu.RUnlock()
	return (*db)[id]
}

// Snapshot returns the registered status pointers in a stable order.
// Callers must lock each status before reading or changing its fields.
func (db *Stdb) Snapshot() []*PayloadStatus {
	StatsMu.RLock()
	statuses := make([]*PayloadStatus, 0, len(*db))
	for _, status := range *db {
		statuses = append(statuses, status)
	}
	StatsMu.RUnlock()

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ID < statuses[j].ID
	})
	return statuses
}
