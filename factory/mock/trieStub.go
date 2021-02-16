package mock

import (
	"context"
	"errors"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

// TrieStub -
type TrieStub struct {
	GetCalled                   func(key []byte) ([]byte, error)
	UpdateCalled                func(key, value []byte) error
	DeleteCalled                func(key []byte) error
	CommitCalled                func() error
	RecreateCalled              func(root []byte) (data.Trie, error)
	CancelPruneCalled           func(rootHash []byte, identifier data.TriePruningIdentifier)
	PruneCalled                 func(rootHash []byte, identifier data.TriePruningIdentifier)
	ResetOldHashesCalled        func() [][]byte
	AppendToOldHashesCalled     func([][]byte)
	GetSerializedNodesCalled    func([]byte, uint64) ([][]byte, uint64, error)
	GetAllHashesCalled          func() ([][]byte, error)
	DatabaseCalled              func() data.DBWriteCacher
	GetAllLeavesOnChannelCalled func(rootHash []byte) (chan core.KeyValueHolder, error)
	GetProofCalled              func(key []byte) ([][]byte, error)
	VerifyProofCalled           func(key []byte, proof [][]byte) (bool, error)
	GetStorageManagerCalled     func() data.StorageManager
	RootHashCalled              func() ([]byte, error)
}

// EnterSnapshotMode -
func (ts *TrieStub) EnterSnapshotMode() {
}

// ExitSnapshotMode -
func (ts *TrieStub) ExitSnapshotMode() {
}

// TakeSnapshot -
func (ts *TrieStub) TakeSnapshot(_ []byte) {
}

// SetCheckpoint -
func (ts *TrieStub) SetCheckpoint(_ []byte) {
}

// GetAllLeaves -
func (ts *TrieStub) GetAllLeaves() (map[string][]byte, error) {
	return make(map[string][]byte), nil
}

// GetAllLeavesOnChannel -
func (ts *TrieStub) GetAllLeavesOnChannel(rootHash []byte, _ context.Context) (chan core.KeyValueHolder, error) {
	if ts.GetAllLeavesOnChannelCalled != nil {
		return ts.GetAllLeavesOnChannelCalled(rootHash)
	}

	ch := make(chan core.KeyValueHolder)
	close(ch)

	return ch, nil
}

// IsPruningEnabled -
func (ts *TrieStub) IsPruningEnabled() bool {
	return false
}

// Get -
func (ts *TrieStub) Get(key []byte) ([]byte, error) {
	if ts.GetCalled != nil {
		return ts.GetCalled(key)
	}

	return nil, nil
}

// Update -
func (ts *TrieStub) Update(key, value []byte) error {
	if ts.UpdateCalled != nil {
		return ts.UpdateCalled(key, value)
	}

	return nil
}

// Delete -
func (ts *TrieStub) Delete(key []byte) error {
	if ts.DeleteCalled != nil {
		return ts.DeleteCalled(key)
	}

	return nil
}

// Commit -
func (ts *TrieStub) Commit() error {
	if ts.CommitCalled != nil {
		return ts.CommitCalled()
	}

	return nil
}

// Recreate -
func (ts *TrieStub) Recreate(root []byte) (data.Trie, error) {
	if ts.RecreateCalled != nil {
		return ts.RecreateCalled(root)
	}

	return nil, nil
}

// String -
func (ts *TrieStub) String() string {
	return "stub trie"
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *TrieStub) IsInterfaceNil() bool {
	return ts == nil
}

// CancelPrune invalidates the hashes that correspond to the given root hash from the eviction waiting list
func (ts *TrieStub) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	if ts.CancelPruneCalled != nil {
		ts.CancelPruneCalled(rootHash, identifier)
	}
}

// Prune removes from the database all the old hashes that correspond to the given root hash
func (ts *TrieStub) Prune(rootHash []byte, identifier data.TriePruningIdentifier) {
	if ts.PruneCalled != nil {
		ts.PruneCalled(rootHash, identifier)
	}
}

// ResetOldHashes resets the oldHashes and oldRoot variables and returns the old hashes
func (ts *TrieStub) ResetOldHashes() [][]byte {
	if ts.ResetOldHashesCalled != nil {
		return ts.ResetOldHashesCalled()
	}

	return nil
}

// AppendToOldHashes appends the given hashes to the trie's oldHashes variable
func (ts *TrieStub) AppendToOldHashes(hashes [][]byte) {
	if ts.AppendToOldHashesCalled != nil {
		ts.AppendToOldHashesCalled(hashes)
	}
}

// GetSerializedNodes -
func (ts *TrieStub) GetSerializedNodes(hash []byte, maxBuffToSend uint64) ([][]byte, uint64, error) {
	if ts.GetSerializedNodesCalled != nil {
		return ts.GetSerializedNodesCalled(hash, maxBuffToSend)
	}
	return nil, 0, nil
}

// Database -
func (ts *TrieStub) Database() data.DBWriteCacher {
	if ts.DatabaseCalled != nil {
		return ts.DatabaseCalled()
	}

	return &testscommon.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return nil, errors.New("key not found")
		},
		PutCalled: func(key, data []byte) error {
			return nil
		},
	}
}

// GetDirtyHashes -
func (ts *TrieStub) GetDirtyHashes() (data.ModifiedHashes, error) {
	return nil, nil
}

// SetNewHashes -
func (ts *TrieStub) SetNewHashes(_ data.ModifiedHashes) {
}

// GetAllHashes -
func (ts *TrieStub) GetAllHashes() ([][]byte, error) {
	if ts.GetAllHashesCalled != nil {
		return ts.GetAllHashesCalled()
	}

	return nil, nil
}

// EnterPruningBufferingMode -
func (ts *TrieStub) EnterPruningBufferingMode() {
}

// ExitPruningBufferingMode -
func (ts *TrieStub) ExitPruningBufferingMode() {
}

// GetSnapshotDbBatchDelay -
func (ts *TrieStub) GetSnapshotDbBatchDelay() int {
	return 0
}

// GetProofCalled -
func (ts *TrieStub) GetProof(key []byte) ([][]byte, error) {
	if ts.GetProofCalled != nil {
		return ts.GetProofCalled(key)
	}
	return nil, nil
}

// VerifyProofCalled
func (ts *TrieStub) VerifyProof(key []byte, proof [][]byte) (bool, error) {
	if ts.VerifyProofCalled != nil {
		return ts.VerifyProofCalled(key, proof)
	}

	return false, nil
}

// GetStorageManager -
func (ts *TrieStub) GetStorageManager() data.StorageManager {
	if ts.GetStorageManagerCalled != nil {
		return ts.GetStorageManagerCalled()
	}
	return nil
}

// RootHash -
func (ts *TrieStub) RootHash() ([]byte, error) {
	if ts.RootHashCalled != nil {
		return ts.RootHashCalled()
	}
	return nil, nil
}
