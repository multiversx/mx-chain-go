package storage

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

// UnitOpenerStub -
type UnitOpenerStub struct {
	OpenDBCalled                   func(dbConfig config.DBConfig, shardID uint32, epoch uint32) (storage.Storer, error)
	GetMostRecentStorageUnitCalled func(config config.DBConfig) (storage.Storer, error)
}

// OpenDB -
func (uohs *UnitOpenerStub) OpenDB(dbConfig config.DBConfig, shardID uint32, epoch uint32) (storage.Storer, error) {
	if uohs.OpenDBCalled != nil {
		return uohs.OpenDBCalled(dbConfig, shardID, epoch)
	}
	return &StorerStub{}, nil
}

// GetMostRecentStorageUnit -
func (uohs *UnitOpenerStub) GetMostRecentStorageUnit(config config.DBConfig) (storage.Storer, error) {
	if uohs.GetMostRecentStorageUnitCalled != nil {
		return uohs.GetMostRecentStorageUnitCalled(config)
	}
	return &StorerStub{}, nil
}

// IsInterfaceNil -
func (uohs *UnitOpenerStub) IsInterfaceNil() bool {
	return uohs == nil
}
