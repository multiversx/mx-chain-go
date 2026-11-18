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
	identifier string
}

// NewGoroutinesManager creates a new GoroutinesManager
func NewGoroutinesManager(
	throttler core.Throttler,
	errorChannel common.BufferedErrChan,
	chanClose chan struct{},
	identifier string,
) (*goroutinesManager, error) {
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
		identifier:   identifier,
	}, nil
}

// ShouldContinueProcessing checks if the processing should be interrupted
func (gm *goroutinesManager) ShouldContinueProcessing() bool {
	select {
	case <-gm.chanClose:
		log.Trace("goroutines manager closed", "identifier", gm.identifier)
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

	log.Trace("starting processing goroutine", "identifier", gm.identifier)
	gm.throttler.StartProcessing()
	return true
}

// EndGoRoutineProcessing ends the processing of a goroutine
func (gm *goroutinesManager) EndGoRoutineProcessing() {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	log.Trace("ending processing goroutine", "identifier", gm.identifier)
	gm.throttler.EndProcessing()
}

// SetNewErrorChannel sets a new error channel
func (gm *goroutinesManager) SetNewErrorChannel(errChan common.BufferedErrChan) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	err := gm.errorChannel.ReadFromChanNonBlocking()
	if err != nil {
		return err
	}

	gm.errorChannel = errChan
	return nil
}

// SetError sets an error in the error channel
func (gm *goroutinesManager) SetError(err error) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	log.Trace("setting error", "identifier", gm.identifier, "error", err)
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
