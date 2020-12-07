package mock

import "github.com/ElrondNetwork/elrond-go/data"

// StorageManagerStub --
type StorageManagerStub struct {
	DatabaseCalled                    func() data.DBWriteCacher
	TakeSnapshotCalled                func([]byte)
	SetCheckpointCalled               func([]byte)
	PruneCalled                       func([]byte)
	CancelPruneCalled                 func([]byte)
	MarkForEvictionCalled             func([]byte, data.ModifiedHashes) error
	GetDbThatContainsHashCalled       func([]byte) data.DBWriteCacher
	GetSnapshotThatContainsHashCalled func(rootHash []byte) data.SnapshotDbHandler
	IsPruningEnabledCalled            func() bool
	EnterPruningBufferingModeCalled   func()
	ExitPruningBufferingModeCalled    func()
	IsInterfaceNilCalled              func() bool
}

// Database --
func (sms *StorageManagerStub) Database() data.DBWriteCacher {
	if sms.DatabaseCalled != nil {
		return sms.DatabaseCalled()
	}
	return nil
}

// TakeSnapshot --
func (sms *StorageManagerStub) TakeSnapshot([]byte) {

}

// SetCheckpoint --
func (sms *StorageManagerStub) SetCheckpoint([]byte) {

}

// Prune --
func (sms *StorageManagerStub) Prune([]byte, data.TriePruningIdentifier) {

}

// CancelPrune --
func (sms *StorageManagerStub) CancelPrune([]byte, data.TriePruningIdentifier) {

}

// MarkForEviction --
func (sms *StorageManagerStub) MarkForEviction(d []byte, m data.ModifiedHashes) error {
	if sms.MarkForEvictionCalled != nil {
		return sms.MarkForEvictionCalled(d, m)
	}
	return nil
}

// GetSnapshotThatContainsHash --
func (sms *StorageManagerStub) GetSnapshotThatContainsHash(d []byte) data.SnapshotDbHandler {
	if sms.GetSnapshotThatContainsHashCalled != nil {
		return sms.GetSnapshotThatContainsHashCalled(d)
	}

	return nil
}

// IsPruningEnabled --
func (sms *StorageManagerStub) IsPruningEnabled() bool {
	if sms.IsPruningEnabledCalled != nil {
		return sms.IsPruningEnabledCalled()
	}
	return false
}

// EnterPruningBufferingMode --
func (sms *StorageManagerStub) EnterPruningBufferingMode() {
	if sms.EnterPruningBufferingModeCalled != nil {
		sms.EnterPruningBufferingModeCalled()
	}
}

// ExitPruningBufferingMode --
func (sms *StorageManagerStub) ExitPruningBufferingMode() {
	if sms.ExitPruningBufferingModeCalled != nil {
		sms.ExitPruningBufferingModeCalled()
	}
}

// GetSnapshotDbBatchDelay -
func (sms *StorageManagerStub) GetSnapshotDbBatchDelay() int {
	return 0
}

// Close -
func (sms *StorageManagerStub) Close() error {
	return nil
}

// IsInterfaceNil --
func (sms *StorageManagerStub) IsInterfaceNil() bool {
	return sms == nil
}
