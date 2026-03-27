package processMocks

import (
	"sync"

	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
)

// BlocksCacheMock is a mock implementation of the BlocksCache interface
type BlocksCacheMock struct {
	AddOrReplaceCalled           func(pair cache.HeaderBodyPair) error
	RemoveAtNonceAndHigherCalled func(nonce uint64) []uint64
	CleanCalled                  func()
	GetByNonceCalled             func(nonce uint64) (cache.HeaderBodyPair, bool)
	signalOnce                   sync.Once
	signalChan                   chan struct{}
}

// GetByNonce -
func (bqm *BlocksCacheMock) GetByNonce(nonce uint64) (cache.HeaderBodyPair, bool) {
	if bqm.GetByNonceCalled != nil {
		return bqm.GetByNonceCalled(nonce)
	}

	return cache.HeaderBodyPair{}, false
}

// Remove -
func (bqm *BlocksCacheMock) Remove(_ uint64) {
}

// AddOrReplace -
func (bqm *BlocksCacheMock) AddOrReplace(pair cache.HeaderBodyPair) error {
	if bqm.AddOrReplaceCalled != nil {
		return bqm.AddOrReplaceCalled(pair)
	}
	return nil
}

// RemoveAtNonceAndHigher -
func (bqm *BlocksCacheMock) RemoveAtNonceAndHigher(nonce uint64) []uint64 {
	if bqm.RemoveAtNonceAndHigherCalled != nil {
		return bqm.RemoveAtNonceAndHigherCalled(nonce)
	}
	return nil
}

// Clean -
func (bqm *BlocksCacheMock) Clean() {
	if bqm.CleanCalled != nil {
		bqm.CleanCalled()
	}
}

// GetSignalBlockAddedChan -
func (bqm *BlocksCacheMock) GetSignalBlockAddedChan() <-chan struct{} {
	bqm.signalOnce.Do(func() {
		bqm.signalChan = make(chan struct{}, 1)
	})
	return bqm.signalChan
}

// IsInterfaceNil -
func (bqm *BlocksCacheMock) IsInterfaceNil() bool {
	return bqm == nil
}
