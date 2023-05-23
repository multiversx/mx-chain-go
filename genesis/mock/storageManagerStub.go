package mock

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// StorageManagerStub -
type StorageManagerStub struct {
	DatabaseCalled                    func() common.BaseStorer
	TakeSnapshotCalled                func([]byte)
	SetCheckpointCalled               func([]byte)
	PruneCalled                       func([]byte)
	CancelPruneCalled                 func([]byte)
	MarkForEvictionCalled             func([]byte, common.ModifiedHashes) error
	GetDbThatContainsHashCalled       func([]byte) common.BaseStorer
	GetSnapshotThatContainsHashCalled func(rootHash []byte) common.SnapshotDbHandler
	IsPruningEnabledCalled            func() bool
	EnterSnapshotModeCalled           func()
	ExitSnapshotModeCalled            func()
	IsInterfaceNilCalled              func() bool
}

// Database -
func (sms *StorageManagerStub) Database() common.BaseStorer {
	if sms.DatabaseCalled != nil {
		return sms.DatabaseCalled()
	}
	return nil
}

// TakeSnapshot -
func (sms *StorageManagerStub) TakeSnapshot([]byte) {

}

// SetCheckpoint -
func (sms *StorageManagerStub) SetCheckpoint([]byte) {

}

// Prune -
func (sms *StorageManagerStub) Prune([]byte, state.TriePruningIdentifier) {

}

// CancelPrune -
func (sms *StorageManagerStub) CancelPrune([]byte, state.TriePruningIdentifier) {

}

// MarkForEviction -
func (sms *StorageManagerStub) MarkForEviction(d []byte, m common.ModifiedHashes) error {
	if sms.MarkForEvictionCalled != nil {
		return sms.MarkForEvictionCalled(d, m)
	}
	return nil
}

// GetSnapshotThatContainsHash -
func (sms *StorageManagerStub) GetSnapshotThatContainsHash(d []byte) common.SnapshotDbHandler {
	if sms.GetSnapshotThatContainsHashCalled != nil {
		return sms.GetSnapshotThatContainsHashCalled(d)
	}

	return nil
}

// IsPruningEnabled -
func (sms *StorageManagerStub) IsPruningEnabled() bool {
	if sms.IsPruningEnabledCalled != nil {
		return sms.IsPruningEnabledCalled()
	}
	return false
}

// EnterSnapshotMode -
func (sms *StorageManagerStub) EnterSnapshotMode() {
	if sms.EnterSnapshotModeCalled != nil {
		sms.EnterSnapshotModeCalled()
	}
}

// ExitSnapshotMode -
func (sms *StorageManagerStub) ExitSnapshotMode() {
	if sms.ExitSnapshotModeCalled != nil {
		sms.ExitSnapshotModeCalled()
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

// IsInterfaceNil -
func (sms *StorageManagerStub) IsInterfaceNil() bool {
	return sms == nil
}
