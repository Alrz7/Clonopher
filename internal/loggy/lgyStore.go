package loggy

import (
	"fmt"
	"sync"
	"time"
)

type PydLog struct {
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	PdLError = "error"
	PdLWarn  = "warn"
	PdLInfo  = "info"
)

var GLog = []PydLog{}
var Glmu = sync.RWMutex{}

func AddNew(content, Type string) {
	Glmu.Lock()
	defer Glmu.Unlock()
	GLog = append(GLog, PydLog{Content: content, Type: Type, Timestamp: time.Now()})
}

func AddNewf(format, Type string, params ...string) {
	AddNew(fmt.Sprintf(format, params), Type)
}
func AddNewr(content, Type string, err error) {
	AddNew(fmt.Sprintf(content+": %v", err.Error()), Type)
}

func GetAll() []PydLog {
	Glmu.RLock()
	defer Glmu.RUnlock()

	snapshot := make([]PydLog, len(GLog))
	copy(snapshot, GLog)
	return snapshot
}
