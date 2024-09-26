package trie

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type goroutinesManager struct {
	throttler    core.Throttler
	errorChannel common.BufferedErrChan
	chanClose    chan struct{}

	mutex      sync.RWMutex
	canProcess bool
}

// NewGoroutinesManager creates a new GoroutinesManager
func NewGoroutinesManager(
	throttler core.Throttler,
	errorChannel common.BufferedErrChan,
	chanClose chan struct{},
) (common.TrieGoroutinesManager, error) {
	if check.IfNil(throttler) {
		return nil, ErrNilThrottler
	}
	if check.IfNil(errorChannel) {
		return nil, ErrNilBufferedErrChan
	}
	if chanClose == nil {
		return nil, ErrNilChanClose
	}

	return &goroutinesManager{
		throttler:    throttler,
		errorChannel: errorChannel,
		chanClose:    chanClose,
		canProcess:   true,
	}, nil
}

// ShouldContinueProcessing checks if the processing should be interrupted
func (gm *goroutinesManager) ShouldContinueProcessing() bool {
	select {
	case <-gm.chanClose:
		return false
	default:
		gm.mutex.RLock()
		defer gm.mutex.RUnlock()

		return gm.canProcess
	}
}

// CanStartGoRoutine checks if a new goroutine can be started. If yes, it adds a new goroutine to the throttler
func (gm *goroutinesManager) CanStartGoRoutine() bool {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if !gm.throttler.CanProcess() {
		return false
	}

	gm.throttler.StartProcessing()
	return true
}

// EndGoRoutineProcessing ends the processing of a goroutine
func (gm *goroutinesManager) EndGoRoutineProcessing() {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	gm.throttler.EndProcessing()
}

// SetError sets an error in the error channel
func (gm *goroutinesManager) SetError(err error) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	gm.errorChannel.WriteInChanNonBlocking(err)
	gm.canProcess = false
}

// GetError gets an error from the error channel
func (gm *goroutinesManager) GetError() error {
	return gm.errorChannel.ReadFromChanNonBlocking()
}

// IsInterfaceNil returns true if there is no value under the interface
func (gm *goroutinesManager) IsInterfaceNil() bool {
	return gm == nil
}
