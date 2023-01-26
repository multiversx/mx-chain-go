package state

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// StoragePruningManagerStub -
type StoragePruningManagerStub struct {
	MarkForEvictionCalled func(bytes []byte, bytes2 []byte, hashes common.ModifiedHashes, hashes2 common.ModifiedHashes) error
	PruneTrieCalled       func(rootHash []byte, identifier state.TriePruningIdentifier, tsm common.StorageManager, handler state.PruningHandler)
	CancelPruneCalled     func(rootHash []byte, identifier state.TriePruningIdentifier, tsm common.StorageManager)
	CloseCalled           func() error
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

// Close -
func (stub *StoragePruningManagerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *StoragePruningManagerStub) IsInterfaceNil() bool {
	return stub == nil
}
