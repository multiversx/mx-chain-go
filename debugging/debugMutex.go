package debugging

import (
	"fmt"
	"runtime"
	"sync"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("debugging")

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
	dm.logOperation("Lock", file, no)

	dm.mutex.Lock()
}

func (dm *DebugMutex) Unlock() {
	_, file, no, _ := runtime.Caller(1)
	dm.logOperation("Unlock", file, no)

	dm.mutex.Unlock()
}

func (dm *DebugMutex) RLock() {
	_, file, no, _ := runtime.Caller(1)
	dm.logOperation("RLock", file, no)

	dm.mutex.RLock()
}

func (dm *DebugMutex) RUnlock() {
	_, file, no, _ := runtime.Caller(1)
	dm.logOperation("RUnlock", file, no)

	dm.mutex.RUnlock()
}

func (dm *DebugMutex) logOperation(operation string, file string, no int) {
	log.Debug(fmt.Sprintf("%s.%s", dm.name, operation), "file", file, "no", no)
}
