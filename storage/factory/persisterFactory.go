package factory

import (
	"os"
	"path"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/disabled"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// persisterFactory is the factory which will handle creating new databases
type persisterFactory struct {
	dbConfigHandler storage.DBConfigHandler
}

// NewPersisterFactory will return a new instance of persister factory
func NewPersisterFactory(config config.DBConfig) (*persisterFactory, error) {
	dbConfigHandler := NewDBConfigHandler(config)

	return &persisterFactory{
		dbConfigHandler: dbConfigHandler,
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
			if i > 0 {
				log.Debug("Create Persister succeeded after retrying", "path", path, "attempts", i+1)
			}

			return persister, nil
		}

		logLevel := logger.LogDebug
		if i == 0 {
			logLevel = logger.LogWarning
		}
		log.Log(logLevel, "Create Persister failed, will retry",
			"path", path,
			"attempt", i+1,
			"maxAttempts", storage.MaxRetriesToCreateDB,
			"error", err)

		// TODO: extract this in a parameter and inject it
		time.Sleep(storage.SleepTimeBetweenCreateDBRetries)
	}

	log.Warn("Create Persister failed on all attempts",
		"path", path,
		"maxAttempts", storage.MaxRetriesToCreateDB,
		"error", err)

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
