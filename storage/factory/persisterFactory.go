package factory

import (
	"errors"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/disabled"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// PersisterFactory is the factory which will handle creating new databases
type PersisterFactory struct {
	dbType            string
	batchDelaySeconds int
	maxBatchSize      int
	maxOpenFiles      int
}

// NewPersisterFactory will return a new instance of a PersisterFactory
func NewPersisterFactory(config config.DBConfig) *PersisterFactory {
	return &PersisterFactory{
		dbType:            config.Type,
		batchDelaySeconds: config.BatchDelaySeconds,
		maxBatchSize:      config.MaxBatchSize,
		maxOpenFiles:      config.MaxOpenFiles,
	}
}

// Create will return a new instance of a DB with a given path
func (pf *PersisterFactory) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, errors.New("invalid file path")
	}

	switch storageunit.DBType(pf.dbType) {
	case storageunit.LvlDB:
		return database.NewLevelDB(path, pf.batchDelaySeconds, pf.maxBatchSize, pf.maxOpenFiles)
	case storageunit.LvlDBSerial:
		return database.NewSerialDB(path, pf.batchDelaySeconds, pf.maxBatchSize, pf.maxOpenFiles)
	case storageunit.MemoryDB:
		return database.NewMemDB(), nil
	default:
		return nil, storage.ErrNotSupportedDBType
	}
}

// CreateDisabled will return a new disabled persister
func (pf *PersisterFactory) CreateDisabled() storage.Persister {
	return disabled.NewErrorDisabledPersister()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *PersisterFactory) IsInterfaceNil() bool {
	return pf == nil
}
