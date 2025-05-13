package debugging

import (
	"runtime"
	"sync"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("debugMutex")

type DebugMutex struct {
	name  string
	mutex sync.RWMutex
}

func NewDebugMutex(name string) *DebugMutex {
	return &DebugMutex{
		name: name,
	}
}

func (dm *DebugMutex) Lock() {
	_, file, no, _ := runtime.Caller(1)
	log.Debug("Lock", "mutex", dm.name, "file", file, "no", no)

	dm.mutex.Lock()
}

func (dm *DebugMutex) Unlock() {
	_, file, no, _ := runtime.Caller(1)
	log.Debug("Lock", "mutex", dm.name, "file", file, "no", no)

	dm.mutex.Unlock()
}

func (dm *DebugMutex) RLock() {
	_, file, no, _ := runtime.Caller(1)
	log.Debug("Lock", "mutex", dm.name, "file", file, "no", no)

	dm.mutex.RLock()
}

func (dm *DebugMutex) RUnlock() {
	_, file, no, _ := runtime.Caller(1)
	log.Debug("Lock", "mutex", dm.name, "file", file, "no", no)

	dm.mutex.RUnlock()
}
