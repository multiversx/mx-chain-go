package factory

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// PersisterFactory is the factory which will handle creating new databases
type PersisterFactory struct {
	dbType            string
	batchDelaySeconds int
	maxBatchSize      int
	maxOpenFiles      int
	processingMode    common.NodeProcessingMode
}

// NewPersisterFactory will return a new instance of a PersisterFactory
func NewPersisterFactory(config config.DBConfig, processingMode common.NodeProcessingMode) *PersisterFactory {
	return &PersisterFactory{
		dbType:            config.Type,
		batchDelaySeconds: config.BatchDelaySeconds,
		maxBatchSize:      config.MaxBatchSize,
		maxOpenFiles:      config.MaxOpenFiles,
		processingMode:    processingMode,
	}
}

// Create will return a new instance of a DB with a given path
func (pf *PersisterFactory) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, errors.New("invalid file path")
	}

	switch storageUnit.DBType(pf.dbType) {
	case storageUnit.LvlDB:
		return leveldb.NewDB(path, pf.batchDelaySeconds, pf.maxBatchSize, pf.maxOpenFiles, pf.processingMode)
	case storageUnit.LvlDBSerial:
		return leveldb.NewSerialDB(path, pf.batchDelaySeconds, pf.maxBatchSize, pf.maxOpenFiles, pf.processingMode)
	case storageUnit.MemoryDB:
		return memorydb.New(), nil
	default:
		return nil, storage.ErrNotSupportedDBType
	}
}

// CreateDisabled will return a new disabled persister
func (pf *PersisterFactory) CreateDisabled() storage.Persister {
	return &disabledPersister{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *PersisterFactory) IsInterfaceNil() bool {
	return pf == nil
}
