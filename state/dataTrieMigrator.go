package state

import "github.com/multiversx/mx-chain-core-go/core"

type DataTrieMigratorMock struct {
	leavesToBeMigrated []core.TrieData
}

func NewDataTrieMigratorMock() *DataTrieMigratorMock {
	return &DataTrieMigratorMock{
		leavesToBeMigrated: make([]core.TrieData, 0),
	}
}

func (d *DataTrieMigratorMock) ConsumeStorageLoadGas() bool {
	return true
}

func (d *DataTrieMigratorMock) AddLeafToMigrationQueue(leafData core.TrieData, _ core.TrieNodeVersion) (bool, error) {
	d.leavesToBeMigrated = append(d.leavesToBeMigrated, leafData)
	return true, nil
}

func (d *DataTrieMigratorMock) GetLeavesToBeMigrated() []core.TrieData {
	defer func() {
		d.leavesToBeMigrated = make([]core.TrieData, 0)
	}()

	return d.leavesToBeMigrated
}

func (d *DataTrieMigratorMock) IsInterfaceNil() bool {
	return d == nil
}
