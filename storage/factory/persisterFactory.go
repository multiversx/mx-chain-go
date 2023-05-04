package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/disabled"
)

// PersisterFactory is the factory which will handle creating new databases
type PersisterFactory struct {
	dbConfigHandler storage.DBConfigHandler
}

// NewPersisterFactory will return a new instance of a PersisterFactory
func NewPersisterFactory(dbConfigHandler storage.DBConfigHandler) (*PersisterFactory, error) {
	if check.IfNil(dbConfigHandler) {
		return nil, storage.ErrNilDBConfigHandler
	}

	return &PersisterFactory{
		dbConfigHandler: dbConfigHandler,
	}, nil
}

// Create will return a new instance of a DB with a given path
func (pf *PersisterFactory) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, storage.ErrInvalidFilePath
	}

	dbConfig, err := pf.dbConfigHandler.GetDBConfig(path)
	if err != nil {
		return nil, err
	}

	pc := newPersisterCreator(*dbConfig)

	persister, err := pc.Create(path)
	if err != nil {
		return nil, err
	}

	err = pf.dbConfigHandler.SaveDBConfigToFilePath(path, dbConfig)
	if err != nil {
		return nil, err
	}

	return persister, nil
}

// CreateDisabled will return a new disabled persister
func (pf *PersisterFactory) CreateDisabled() storage.Persister {
	return disabled.NewErrorDisabledPersister()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *PersisterFactory) IsInterfaceNil() bool {
	return pf == nil
}
