package factory

import (
	"os"
	"path"
	"time"

	"github.com/multiversx/mx-chain-storage-go/pebbledb"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/disabled"
)

// persisterFactory is the factory which will handle creating new databases
type persisterFactory struct {
	dbConfigHandler storage.DBConfigHandler
	pebbleResources *pebbledb.SharedResources
}

// NewPersisterFactory will return a new instance of persister factory; pebble persisters get private caches
func NewPersisterFactory(config config.DBConfig) (*persisterFactory, error) {
	return NewPersisterFactoryWithResources(config, nil)
}

// NewPersisterFactoryWithResources will return a new instance of persister factory whose pebble persisters
// share the provided caches
func NewPersisterFactoryWithResources(config config.DBConfig, pebbleResources *pebbledb.SharedResources) (*persisterFactory, error) {
	return &persisterFactory{
		dbConfigHandler: NewDBConfigHandler(config),
		pebbleResources: pebbleResources,
	}, nil
}

// CreateWithRetries will return a new instance of a DB with a given path
// It will try to create db multiple times
func (pf *persisterFactory) CreateWithRetries(path string) (storage.Persister, error) {
	var persister storage.Persister
	var err error

	for i := 0; i < storage.MaxRetriesToCreateDB; i++ {
		persister, err = pf.Create(path)
		if err == nil {
			return persister, nil
		}
		log.Warn("Create Persister failed", "path", path, "error", err)

		// TODO: extract this in a parameter and inject it
		time.Sleep(storage.SleepTimeBetweenCreateDBRetries)
	}

	return nil, err
}

// Create will return a new instance of a DB with a given path
func (pf *persisterFactory) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, storage.ErrInvalidFilePath
	}

	dbConfig, err := pf.dbConfigHandler.GetDBConfig(path)
	if err != nil {
		return nil, err
	}

	if dbConfig.UseTmpAsFilePath {
		filePath, err := getTmpFilePath(path)
		if err != nil {
			return nil, err
		}

		path = filePath
	}

	pc := newPersisterCreator(*dbConfig, pf.pebbleResources)
	if !isKnownPersistentDBType(dbConfig.Type) {
		return pc.Create(path)
	}

	// written before the engine creates any file, so a crash cannot leave an engine-less directory
	err = pf.dbConfigHandler.SaveDBConfigToFilePath(path, dbConfig)
	if err != nil {
		return nil, err
	}

	persister, err := pc.Create(path)
	if err != nil {
		removeConfigIfAlone(path)
		return nil, err
	}

	return persister, nil
}

// CreateDisabled will return a new disabled persister
func (pf *persisterFactory) CreateDisabled() storage.Persister {
	return disabled.NewErrorDisabledPersister()
}

func getTmpFilePath(p string) (string, error) {
	_, file := path.Split(p)
	return os.MkdirTemp("", file)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *persisterFactory) IsInterfaceNil() bool {
	return pf == nil
}
