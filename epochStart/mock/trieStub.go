package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

var errNotImplemented = errors.New("not implemented")

// TrieStub -
type TrieStub struct {
	GetCalled                   func(key []byte) ([]byte, error)
	UpdateCalled                func(key, value []byte) error
	DeleteCalled                func(key []byte) error
	RootCalled                  func() ([]byte, error)
	CommitCalled                func() error
	RecreateCalled              func(root []byte) (data.Trie, error)
	CancelPruneCalled           func(rootHash []byte, identifier data.TriePruningIdentifier)
	PruneCalled                 func(rootHash []byte, identifier data.TriePruningIdentifier)
	ResetOldHashesCalled        func() [][]byte
	AppendToOldHashesCalled     func([][]byte)
	TakeSnapshotCalled          func(rootHash []byte)
	SetCheckpointCalled         func(rootHash []byte)
	GetSerializedNodesCalled    func([]byte, uint64) ([][]byte, uint64, error)
	DatabaseCalled              func() data.DBWriteCacher
	GetAllLeavesCalled          func() (map[string][]byte, error)
	GetAllHashesCalled          func() ([][]byte, error)
	IsPruningEnabledCalled      func() bool
	ClosePersisterCalled        func() error
	GetAllLeavesOnChannelCalled func() chan core.KeyValueHolder
}

// EnterSnapshotMode -
func (ts *TrieStub) EnterSnapshotMode() {
}

// ExitSnapshotMode -
func (ts *TrieStub) ExitSnapshotMode() {
}

// Get -
func (ts *TrieStub) Get(key []byte) ([]byte, error) {
	if ts.GetCalled != nil {
		return ts.GetCalled(key)
	}

	return nil, errNotImplemented
}

// Update -
func (ts *TrieStub) Update(key, value []byte) error {
	if ts.UpdateCalled != nil {
		return ts.UpdateCalled(key, value)
	}

	return errNotImplemented
}

// Delete -
func (ts *TrieStub) Delete(key []byte) error {
	if ts.DeleteCalled != nil {
		return ts.DeleteCalled(key)
	}

	return errNotImplemented
}

// Root -
func (ts *TrieStub) Root() ([]byte, error) {
	if ts.RootCalled != nil {
		return ts.RootCalled()
	}

	return nil, errNotImplemented
}

// Commit -
func (ts *TrieStub) Commit() error {
	if ts != nil {
		return ts.CommitCalled()
	}

	return errNotImplemented
}

// Recreate -
func (ts *TrieStub) Recreate(root []byte) (data.Trie, error) {
	if ts.RecreateCalled != nil {
		return ts.RecreateCalled(root)
	}

	return nil, errNotImplemented
}

// String -
func (ts *TrieStub) String() string {
	return "stub trie"
}

// GetAllLeaves -
func (ts *TrieStub) GetAllLeaves() (map[string][]byte, error) {
	if ts.GetAllLeavesCalled != nil {
		return ts.GetAllLeavesCalled()
	}

	return nil, errNotImplemented
}

// GetAllLeavesOnChannel -
func (ts *TrieStub) GetAllLeavesOnChannel() chan core.KeyValueHolder {
	if ts.GetAllLeavesOnChannelCalled != nil {
		return ts.GetAllLeavesOnChannelCalled()
	}

	ch := make(chan core.KeyValueHolder)
	close(ch)

	return ch
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

// TakeSnapshot -
func (ts *TrieStub) TakeSnapshot(rootHash []byte) {
	if ts.TakeSnapshotCalled != nil {
		ts.TakeSnapshotCalled(rootHash)
	}
}

// SetCheckpoint -
func (ts *TrieStub) SetCheckpoint(rootHash []byte) {
	if ts.SetCheckpointCalled != nil {
		ts.SetCheckpointCalled(rootHash)
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
	return nil
}

// IsPruningEnabled -
func (ts *TrieStub) IsPruningEnabled() bool {
	if ts.IsPruningEnabledCalled != nil {
		return ts.IsPruningEnabledCalled()
	}
	return false
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

// GetSnapshotDbBatchDelay -
func (ts *TrieStub) GetSnapshotDbBatchDelay() int {
	return 0
}
