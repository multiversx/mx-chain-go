package factory

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

const (
	dbConfigFileName         = "config.toml"
	defaultType              = "LvlDBSerial"
	defaultBatchDelaySeconds = 2
	defaultMaxBatchSize      = 100
	defaultMaxOpenFiles      = 10
)

// PersisterFactory is the factory which will handle creating new databases
type PersisterFactory struct {
	dbType              string
	batchDelaySeconds   int
	maxBatchSize        int
	maxOpenFiles        int
	shardIDProviderType string
	numShards           int32
}

// NewPersisterFactory will return a new instance of a PersisterFactory
func NewPersisterFactory(config config.DBConfig) (*PersisterFactory, error) {
	return &PersisterFactory{
		dbType:              config.Type,
		batchDelaySeconds:   config.BatchDelaySeconds,
		maxBatchSize:        config.MaxBatchSize,
		maxOpenFiles:        config.MaxOpenFiles,
		shardIDProviderType: config.ShardIDProviderType,
		numShards:           config.NumShards,
	}, nil
}

// Create will return a new instance of a DB with a given path
func (pf *PersisterFactory) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, storage.ErrInvalidFilePath
	}

	dbConfig, err := pf.getDBConfig(path)
	if err != nil {
		return nil, err
	}

	pc, err := newPersisterCreator(*dbConfig)
	if err != nil {
		return nil, err
	}

	persister, err := pc.Create(path)
	if err != nil {
		return nil, err
	}

	err = createPersisterConfigFile(path, dbConfig)
	if err != nil {
		return nil, err
	}

	return persister, nil
}

func (pf *PersisterFactory) getDBConfig(path string) (*config.DBConfig, error) {
	dbConfigFromFile := &config.DBConfig{}
	err := core.LoadTomlFile(dbConfigFromFile, getPersisterConfigFilePath(path))
	if err == nil {
		log.Debug("getDBConfig: loaded db config from toml config file", "path", dbConfigFromFile)
		return dbConfigFromFile, nil
	}

	empty := checkIfDirIsEmpty(path)
	if !empty {
		dbConfig := &config.DBConfig{
			Type:              defaultType,
			BatchDelaySeconds: defaultBatchDelaySeconds,
			MaxBatchSize:      defaultMaxBatchSize,
			MaxOpenFiles:      defaultMaxOpenFiles,
		}

		log.Debug("getDBConfig: loaded default db config")
		return dbConfig, nil
	}

	dbConfig := &config.DBConfig{
		Type:                pf.dbType,
		BatchDelaySeconds:   pf.batchDelaySeconds,
		MaxBatchSize:        pf.maxBatchSize,
		MaxOpenFiles:        pf.maxOpenFiles,
		ShardIDProviderType: pf.shardIDProviderType,
		NumShards:           pf.numShards,
	}

	log.Debug("getDBConfig: loaded db config from main config file")
	return dbConfig, nil
}

func checkIfDirIsEmpty(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Debug("getDBConfig: failed to check if dir is empty", "path", path, "error", err.Error())
		return true
	}

	if len(files) == 0 {
		return true
	}

	return false
}

func createPersisterConfigFile(path string, dbConfig *config.DBConfig) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// in memory db, no files available
			log.Debug("createPersisterConfigFile: provided path not available, config file will not be created")
			return nil
		}

		return err
	}

	configFilePath := getPersisterConfigFilePath(path)
	f, err := core.OpenFile(configFilePath)
	if err == nil {
		// config file already exists, no need to create config
		return nil
	}

	defer func() {
		_ = f.Close()
	}()

	err = core.SaveTomlFile(dbConfig, configFilePath)
	if err != nil {
		return err
	}

	return nil
}

func getPersisterConfigFilePath(path string) string {
	return filepath.Join(
		path,
		dbConfigFileName,
	)
}

// CreateDisabled will return a new disabled persister
func (pf *PersisterFactory) CreateDisabled() storage.Persister {
	return &disabledPersister{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *PersisterFactory) IsInterfaceNil() bool {
	return pf == nil
}
