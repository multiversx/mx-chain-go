package processMocks

import "github.com/multiversx/mx-chain-go/process/asyncExecution/queue"

// BlocksQueueMock is a mock implementation of the BlocksQueue interface
type BlocksQueueMock struct {
	AddOrReplaceCalled           func(pair queue.HeaderBodyPair) error
	PopCalled                    func() (queue.HeaderBodyPair, bool)
	PeekCalled                   func() (queue.HeaderBodyPair, bool)
	ValidateQueueIntegrityCalled func() error
	RemoveAtNonceAndHigherCalled func(nonce uint64) []uint64
	CleanCalled                  func(lastAddedNonce uint64)
	CloseCalled                  func()
}

// AddOrReplace -
func (bqm *BlocksQueueMock) AddOrReplace(pair queue.HeaderBodyPair) error {
	if bqm.AddOrReplaceCalled != nil {
		return bqm.AddOrReplaceCalled(pair)
	}
	return nil
}

// Pop -
func (bqm *BlocksQueueMock) Pop() (queue.HeaderBodyPair, bool) {
	if bqm.PopCalled != nil {
		return bqm.PopCalled()
	}
	return queue.HeaderBodyPair{}, false
}

// Peek -
func (bqm *BlocksQueueMock) Peek() (queue.HeaderBodyPair, bool) {
	if bqm.PeekCalled != nil {
		return bqm.PeekCalled()
	}
	return queue.HeaderBodyPair{}, false
}

// ValidateQueueIntegrity -
func (bqm *BlocksQueueMock) ValidateQueueIntegrity() error {
	if bqm.ValidateQueueIntegrityCalled != nil {
		return bqm.ValidateQueueIntegrityCalled()
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
func (bqm *BlocksQueueMock) Clean(lastAddedNonce uint64) {
	if bqm.CleanCalled != nil {
		bqm.CleanCalled(lastAddedNonce)
	}
}

// Close -
func (bqm *BlocksQueueMock) Close() {
	if bqm.CloseCalled != nil {
		bqm.CloseCalled()
	}
}

// IsInterfaceNil -
func (bqm *BlocksQueueMock) IsInterfaceNil() bool {
	return bqm == nil
}
