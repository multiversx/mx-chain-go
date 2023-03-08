package trie

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// DataTrieMigratorStub -
type DataTrieMigratorStub struct {
	ConsumeStorageLoadGasCalled   func() bool
	AddLeafToMigrationQueueCalled func(leafData core.TrieData, newLeafVersion core.TrieNodeVersion) (bool, error)
	GetLeavesToBeMigratedCalled   func() []core.TrieData
}

// ConsumeStorageLoadGas -
func (d *DataTrieMigratorStub) ConsumeStorageLoadGas() bool {
	if d.ConsumeStorageLoadGasCalled != nil {
		return d.ConsumeStorageLoadGasCalled()
	}

	return true
}

// AddLeafToMigrationQueue -
func (d *DataTrieMigratorStub) AddLeafToMigrationQueue(leafData core.TrieData, newLeafVersion core.TrieNodeVersion) (bool, error) {
	if d.AddLeafToMigrationQueueCalled != nil {
		return d.AddLeafToMigrationQueueCalled(leafData, newLeafVersion)
	}

	return true, nil
}

// GetLeavesToBeMigrated -
func (d *DataTrieMigratorStub) GetLeavesToBeMigrated() []core.TrieData {
	if d.GetLeavesToBeMigratedCalled != nil {
		return d.GetLeavesToBeMigratedCalled()
	}

	return nil
}

// IsInterfaceNil -
func (d *DataTrieMigratorStub) IsInterfaceNil() bool {
	return d == nil
}
