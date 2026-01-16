package processMocks

import (
	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
)

// BlocksQueueMock is a mock implementation of the BlocksCache interface
type BlocksQueueMock struct {
	AddOrReplaceCalled           func(pair cache.HeaderBodyPair) error
	RemoveAtNonceAndHigherCalled func(nonce uint64) []uint64
	CleanCalled                  func()
	GetByNonceCalled             func(nonce uint64) (cache.HeaderBodyPair, bool)
}

// GetByNonce -
func (bqm *BlocksQueueMock) GetByNonce(nonce uint64) (cache.HeaderBodyPair, bool) {
	if bqm.GetByNonceCalled != nil {
		return bqm.GetByNonceCalled(nonce)
	}

	return cache.HeaderBodyPair{}, false
}

// GetLastAdded -
func (bqm *BlocksQueueMock) GetLastAdded() (cache.HeaderBodyPair, bool) {
	return cache.HeaderBodyPair{}, false
}

// Remove -
func (bqm *BlocksQueueMock) Remove(_ uint64) {
}

// AddOrReplace -
func (bqm *BlocksQueueMock) AddOrReplace(pair cache.HeaderBodyPair) error {
	if bqm.AddOrReplaceCalled != nil {
		return bqm.AddOrReplaceCalled(pair)
	}
	return nil
}

// RemoveAtNonceAndHigher -
func (bqm *BlocksQueueMock) RemoveAtNonceAndHigher(nonce uint64) []uint64 {
	if bqm.RemoveAtNonceAndHigherCalled != nil {
		return bqm.RemoveAtNonceAndHigherCalled(nonce)
	}
	return nil
}

// Clean -
func (bqm *BlocksQueueMock) Clean() {
	if bqm.CleanCalled != nil {
		bqm.CleanCalled()
	}
}

// IsInterfaceNil -
func (bqm *BlocksQueueMock) IsInterfaceNil() bool {
	return bqm == nil
}
