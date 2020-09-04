package testscommon

import (
	"sync"
)

type busyCPUHandler struct {
	mutOperation   sync.RWMutex
	alreadyStarted bool
	canceled       bool
	numCPUs        int
}

// NewBusyCPUHandler creates a handler that can keep the provided number of CPUs busy
func NewBusyCPUHandler(numLogicalCPUs int) *busyCPUHandler {
	if numLogicalCPUs < 0 {
		return nil
	}

	return &busyCPUHandler{
		alreadyStarted: false,
		canceled:       false,
		numCPUs:        numLogicalCPUs,
	}
}

// Start will start the free running go routines that will eat up CPU time
func (bch *busyCPUHandler) Start() {
	bch.mutOperation.Lock()
	defer bch.mutOperation.Unlock()

	if bch.alreadyStarted {
		return
	}

	bch.alreadyStarted = true

	for i := 0; i < bch.numCPUs; i++ {
		go func() {
			for {
				if bch.shouldCancel() {
					return
				}
			}
		}()
	}
}

func (bch *busyCPUHandler) shouldCancel() bool {
	bch.mutOperation.RLock()
	defer bch.mutOperation.RUnlock()

	return bch.canceled
}

// Stop will stop the previous started go routines
func (bch *busyCPUHandler) Stop() {
	bch.mutOperation.Lock()
	defer bch.mutOperation.Unlock()

	if !bch.alreadyStarted {
		return
	}

	bch.canceled = true
	bch.alreadyStarted = false
}
