package mock

import "github.com/ElrondNetwork/elrond-go/data"

// TrieStorageManagerStub -
type TrieStorageManagerStub struct {
	DatabaseCalled                    func() data.DBWriteCacher
	TakeSnapshotCalled                func(bytes []byte)
	SetCheckpointCalled               func(bytes []byte)
	PruneCalled                       func(bytes []byte, identifier data.TriePruningIdentifier)
	CancelPruneCalled                 func(bytes []byte, identifier data.TriePruningIdentifier)
	MarkForEvictionCalled             func(bytes []byte, hashes data.ModifiedHashes) error
	GetSnapshotThatContainsHashCalled func(rootHash []byte) data.SnapshotDbHandler
	IsPruningEnabledCalled            func() bool
	EnterPruningBufferingModeCalled   func()
	ExitPruningBufferingModeCalled    func()
	GetSnapshotDbBatchDelayCalled     func() int
	CloseCalled                       func() error
}

// Database -
func (tms *TrieStorageManagerStub) Database() data.DBWriteCacher {
	if tms.DatabaseCalled != nil {
		return tms.DatabaseCalled()
	}

	return nil
}

// TakeSnapshot -
func (tms *TrieStorageManagerStub) TakeSnapshot(bytes []byte) {
	if tms.TakeSnapshotCalled != nil {
		tms.TakeSnapshotCalled(bytes)
	}
}

// SetCheckpoint -
func (tms *TrieStorageManagerStub) SetCheckpoint(bytes []byte) {
	if tms.SetCheckpointCalled != nil {
		tms.SetCheckpointCalled(bytes)
	}
}

// Prune -
func (tms *TrieStorageManagerStub) Prune(bytes []byte, identifier data.TriePruningIdentifier) {
	if tms.PruneCalled != nil {
		tms.PruneCalled(bytes, identifier)
	}
}

// CancelPrune -
func (tms *TrieStorageManagerStub) CancelPrune(bytes []byte, identifier data.TriePruningIdentifier) {
	if tms.CancelPruneCalled != nil {
		tms.CancelPruneCalled(bytes, identifier)
	}
}

// MarkForEviction -
func (tms *TrieStorageManagerStub) MarkForEviction(bytes []byte, hashes data.ModifiedHashes) error {
	if tms.MarkForEvictionCalled != nil {
		return tms.MarkForEvictionCalled(bytes, hashes)
	}

	return nil
}

// GetSnapshotThatContainsHash -
func (tms *TrieStorageManagerStub) GetSnapshotThatContainsHash(rootHash []byte) data.SnapshotDbHandler {
	if tms.GetSnapshotThatContainsHashCalled != nil {
		return tms.GetSnapshotThatContainsHashCalled(rootHash)
	}

	return nil
}

// IsPruningEnabled -
func (tms *TrieStorageManagerStub) IsPruningEnabled() bool {
	if tms.IsPruningEnabledCalled != nil {
		return tms.IsPruningEnabledCalled()
	}

	return false
}

// EnterPruningBufferingMode -
func (tms *TrieStorageManagerStub) EnterPruningBufferingMode() {
	if tms.EnterPruningBufferingModeCalled != nil {
		tms.EnterPruningBufferingModeCalled()
	}
}

// ExitPruningBufferingMode -
func (tms *TrieStorageManagerStub) ExitPruningBufferingMode() {
	if tms.ExitPruningBufferingModeCalled != nil {
		tms.ExitPruningBufferingModeCalled()
	}
}

// GetSnapshotDbBatchDelay -
func (tms *TrieStorageManagerStub) GetSnapshotDbBatchDelay() int {
	if tms.GetSnapshotDbBatchDelayCalled != nil {
		return tms.GetSnapshotDbBatchDelayCalled()
	}

	return 0
}

// Close -
func (tms *TrieStorageManagerStub) Close() error {
	if tms.CloseCalled != nil {
		return tms.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (tms *TrieStorageManagerStub) IsInterfaceNil() bool {
	return tms == nil
}
