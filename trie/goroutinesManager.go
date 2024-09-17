package trie

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
)

type goroutinesManager struct {
	throttler    core.Throttler
	errorChannel common.BufferedErrChan
	chanClose    chan struct{}

	mutex      sync.RWMutex
	canProcess bool
}

func NewGoroutinesManager(
	throttler core.Throttler,
	errorChannel common.BufferedErrChan,
	chanClose chan struct{},
) (common.TrieGoroutinesManager, error) {
	// TODO: check arguments and add logging in this file

	return &goroutinesManager{
		throttler:    throttler,
		errorChannel: errorChannel,
		chanClose:    chanClose,
		canProcess:   true,
	}, nil
}

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

func (gm *goroutinesManager) CanStartGoRoutine() bool {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if !gm.throttler.CanProcess() {
		return false
	}

	gm.throttler.StartProcessing()
	return true
}

func (gm *goroutinesManager) EndGoRoutineProcessing() {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	gm.throttler.EndProcessing()
}

func (gm *goroutinesManager) SetError(err error) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	gm.errorChannel.WriteInChanNonBlocking(err)
	gm.canProcess = false
}

func (gm *goroutinesManager) GetError() error {
	return gm.errorChannel.ReadFromChanNonBlocking()
}

func (gm *goroutinesManager) IsInterfaceNil() bool {
	return gm == nil
}
