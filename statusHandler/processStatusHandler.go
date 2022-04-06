package statusHandler

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("statusHandler")

type processStatusHandler struct {
	mutStatus sync.RWMutex
	isIdle    bool
}

// NewProcessStatusHandler creates a new instance of type processStatusHandler
func NewProcessStatusHandler() *processStatusHandler {
	return &processStatusHandler{
		isIdle: true, // always start as idle so the initial snapshots (if required) will work
	}
}

// SetToBusy will set the internal state to "busy"
func (psh *processStatusHandler) SetToBusy(reason string) {
	log.Debug("processStatusHandler.SetToBusy", "reason", reason)

	psh.mutStatus.Lock()
	psh.isIdle = false
	psh.mutStatus.Unlock()
}

// SetToIdle will set the internal state to "idle"
func (psh *processStatusHandler) SetToIdle() {
	log.Debug("processStatusHandler.SetToIdle")

	psh.mutStatus.Lock()
	psh.isIdle = true
	psh.mutStatus.Unlock()
}

// IsIdle returns true if the node is idle
func (psh *processStatusHandler) IsIdle() bool {
	psh.mutStatus.RLock()
	defer psh.mutStatus.RUnlock()

	return psh.isIdle
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *processStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}
