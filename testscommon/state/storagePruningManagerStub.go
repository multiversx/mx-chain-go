package state

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// StoragePruningManagerStub -
type StoragePruningManagerStub struct {
	MarkForEvictionCalled             func(bytes []byte, bytes2 []byte, hashes common.ModifiedHashes, hashes2 common.ModifiedHashes) error
	PruneTrieCalled                   func(rootHash []byte, identifier state.TriePruningIdentifier, tsm common.StorageManager, handler state.PruningHandler)
	CancelPruneCalled                 func(rootHash []byte, identifier state.TriePruningIdentifier, tsm common.StorageManager)
	ResetCalled                       func()
	EvictionWaitingListCacheLenCalled func() int
	CloseCalled                       func() error
}

// MarkForEviction -
func (stub *StoragePruningManagerStub) MarkForEviction(bytes []byte, bytes2 []byte, hashes common.ModifiedHashes, hashes2 common.ModifiedHashes) error {
	if stub.MarkForEvictionCalled != nil {
		return stub.MarkForEvictionCalled(bytes, bytes2, hashes, hashes2)
	}

	return nil
}

// PruneTrie -
func (stub *StoragePruningManagerStub) PruneTrie(rootHash []byte, identifier state.TriePruningIdentifier, tsm common.StorageManager, handler state.PruningHandler) {
	if stub.PruneTrieCalled != nil {
		stub.PruneTrieCalled(rootHash, identifier, tsm, handler)
	}
}

// CancelPrune -
func (stub *StoragePruningManagerStub) CancelPrune(rootHash []byte, identifier state.TriePruningIdentifier, tsm common.StorageManager) {
	if stub.CancelPruneCalled != nil {
		stub.CancelPruneCalled(rootHash, identifier, tsm)
	}
}

// Reset -
func (stub *StoragePruningManagerStub) Reset() {
	if stub.ResetCalled != nil {
		stub.ResetCalled()
	}
}

// Close -
func (stub *StoragePruningManagerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// EvictionWaitingListCacheLen -
func (stub *StoragePruningManagerStub) EvictionWaitingListCacheLen() int {
	if stub.EvictionWaitingListCacheLenCalled != nil {
		return stub.EvictionWaitingListCacheLenCalled()
	}
	return 0
}

// IsInterfaceNil -
func (stub *StoragePruningManagerStub) IsInterfaceNil() bool {
	return stub == nil
}
