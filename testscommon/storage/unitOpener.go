package storage

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// UnitOpenerStub -
type UnitOpenerStub struct {
	OpenDBCalled                   func(dbConfig config.DBConfig, shardID uint32, epoch uint32, processingMode common.NodeProcessingMode) (storage.Storer, error)
	GetMostRecentStorageUnitCalled func(config config.DBConfig, processingMode common.NodeProcessingMode) (storage.Storer, error)
}

// OpenDB -
func (uohs *UnitOpenerStub) OpenDB(dbConfig config.DBConfig, shardID uint32, epoch uint32, processingMode common.NodeProcessingMode) (storage.Storer, error) {
	if uohs.OpenDBCalled != nil {
		return uohs.OpenDBCalled(dbConfig, shardID, epoch, processingMode)
	}
	return &StorerStub{}, nil
}

// GetMostRecentStorageUnit -
func (uohs *UnitOpenerStub) GetMostRecentStorageUnit(config config.DBConfig, processingMode common.NodeProcessingMode) (storage.Storer, error) {
	if uohs.GetMostRecentStorageUnitCalled != nil {
		return uohs.GetMostRecentStorageUnitCalled(config, processingMode)
	}
	return &StorerStub{}, nil
}

// IsInterfaceNil -
func (uohs *UnitOpenerStub) IsInterfaceNil() bool {
	return uohs == nil
}
