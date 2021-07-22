package mock

import (
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// StorageManagerStub -
type StorageManagerStub struct {
	DatabaseCalled                    func() temporary.DBWriteCacher
	TakeSnapshotCalled                func([]byte)
	SetCheckpointCalled               func([]byte)
	GetDbThatContainsHashCalled       func([]byte) temporary.DBWriteCacher
	GetSnapshotThatContainsHashCalled func(rootHash []byte) temporary.SnapshotDbHandler
	IsPruningEnabledCalled            func() bool
	IsPruningBlockedCalled            func() bool
	EnterPruningBufferingModeCalled   func()
	ExitPruningBufferingModeCalled    func()
	IsInterfaceNilCalled              func() bool
}

// Database -
func (sms *StorageManagerStub) Database() temporary.DBWriteCacher {
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

// GetSnapshotThatContainsHash -
func (sms *StorageManagerStub) GetSnapshotThatContainsHash(d []byte) temporary.SnapshotDbHandler {
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

// IsPruningBlocked -
func (sms *StorageManagerStub) IsPruningBlocked() bool {
	if sms.IsPruningBlockedCalled != nil {
		return sms.IsPruningBlockedCalled()
	}
	return false
}

// GetSnapshotDbBatchDelay -
func (sms *StorageManagerStub) GetSnapshotDbBatchDelay() int {
	return 0
}

// Close -
func (sms *StorageManagerStub) Close() error {
	return nil
}

// EnterPruningBufferingMode -
func (sms *StorageManagerStub) EnterPruningBufferingMode() {
	if sms.EnterPruningBufferingModeCalled != nil {
		sms.EnterPruningBufferingModeCalled()
	}
}

// ExitPruningBufferingMode -
func (sms *StorageManagerStub) ExitPruningBufferingMode() {
	if sms.ExitPruningBufferingModeCalled != nil {
		sms.ExitPruningBufferingModeCalled()
	}
}

// IsInterfaceNil -
func (sms *StorageManagerStub) IsInterfaceNil() bool {
	return sms == nil
}
