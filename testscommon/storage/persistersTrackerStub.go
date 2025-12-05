package storage

import "github.com/multiversx/mx-chain-go/storage"

// PersistersTrackerStub -
type PersistersTrackerStub struct {
	HasInitializedEnoughPersistersCalled func(epoch int64) bool
	ShouldClosePersisterCalled           func(epoch int64) bool
	CollectPersisterDataCalled           func(p storage.Persister)
}

// HasInitializedEnoughPersisters -
func (p *PersistersTrackerStub) HasInitializedEnoughPersisters(epoch int64) bool {
	if p.HasInitializedEnoughPersistersCalled != nil {
		return p.HasInitializedEnoughPersistersCalled(epoch)
	}

	return false
}

// ShouldClosePersister -
func (p *PersistersTrackerStub) ShouldClosePersister(epoch int64) bool {
	if p.ShouldClosePersisterCalled != nil {
		return p.ShouldClosePersisterCalled(epoch)
	}

	return false
}

// CollectPersisterData -
func (p *PersistersTrackerStub) CollectPersisterData(persister storage.Persister) {
	if p.CollectPersisterDataCalled != nil {
		p.CollectPersisterDataCalled(persister)
	}
}

// IsInterfaceNil -
func (p *PersistersTrackerStub) IsInterfaceNil() bool {
	return p == nil
}
