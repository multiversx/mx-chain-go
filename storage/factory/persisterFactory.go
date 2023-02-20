package factory

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/pelletier/go-toml"
)

// PersisterFactory is the factory which will handle creating new databases
type PersisterFactory struct {
	dbType              string
	batchDelaySeconds   int
	maxBatchSize        int
	maxOpenFiles        int
	shardIDProvider     storage.ShardIDProvider
	shardIDProviderType string
	numShards           uint32
}

// NewPersisterFactory will return a new instance of a PersisterFactory
func NewPersisterFactory(config config.DBConfig, shardIDProvider storage.ShardIDProvider) (*PersisterFactory, error) {
	if check.IfNil(shardIDProvider) {
		return nil, storage.ErrNilShardIDProvider
	}

	return &PersisterFactory{
		dbType:              config.Type,
		batchDelaySeconds:   config.BatchDelaySeconds,
		maxBatchSize:        config.MaxBatchSize,
		maxOpenFiles:        config.MaxOpenFiles,
		shardIDProvider:     shardIDProvider,
		shardIDProviderType: config.ShardIDProviderType,
		numShards:           config.NumShards,
	}, nil
}

// Create will return a new instance of a DB with a given path
func (pf *PersisterFactory) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, errors.New("invalid file path")
	}

	dbConfig, err := pf.getDBConfig(path)
	if err != nil {
		return nil, err
	}

	persister, err := pf.createDB(path, dbConfig)
	if err != nil {
		return nil, err
	}

	err = pf.createPersisterConfigFile(path, dbConfig)
	if err != nil {
		return nil, err
	}

	return persister, nil
}

func (pf *PersisterFactory) getDBConfig(path string) (*config.DBConfig, error) {
	dbConfigFromFile := &config.DBConfig{}
	err := core.LoadTomlFile(dbConfigFromFile, pf.getPersisterConfigFilePath(path))
	if err == nil {
		return dbConfigFromFile, nil
	}

	dbConfig := &config.DBConfig{
		Type:                pf.dbType,
		BatchDelaySeconds:   pf.batchDelaySeconds,
		MaxBatchSize:        pf.maxBatchSize,
		MaxOpenFiles:        pf.maxOpenFiles,
		ShardIDProviderType: pf.shardIDProviderType,
		NumShards:           pf.numShards,
	}

	return dbConfig, nil
}

func (pf *PersisterFactory) createDB(path string, dbConfig *config.DBConfig) (storage.Persister, error) {
	dbType := storageunit.DBType(dbConfig.Type)
	switch dbType {
	case storageunit.LvlDB:
		return database.NewLevelDB(path, dbConfig.BatchDelaySeconds, dbConfig.MaxBatchSize, dbConfig.MaxOpenFiles)
	case storageunit.LvlDBSerial:
		return database.NewSerialDB(path, dbConfig.BatchDelaySeconds, dbConfig.MaxBatchSize, dbConfig.MaxOpenFiles)
	case storageunit.ShardedLvlDBSerial:
		shardIDProvider, err := pf.createShardIDProvider()
		if err != nil {
			return nil, err
		}
		return database.NewShardedDB(storageunit.LvlDBSerial, path, dbConfig.BatchDelaySeconds, dbConfig.MaxBatchSize, dbConfig.MaxOpenFiles, shardIDProvider)
	case storageunit.MemoryDB:
		return database.NewMemDB(), nil
	default:
		return nil, storage.ErrNotSupportedDBType
	}
}

func (pf *PersisterFactory) createPersisterConfigFile(path string, dbConfig *config.DBConfig) error {
	configFilePath := pf.getPersisterConfigFilePath(path)
	f, err := core.OpenFile(configFilePath)
	if err == nil {
		// config file already exists, no need to create config
		return nil
	}

	defer func() {
		_ = f.Close()
	}()

	err = SaveTomlFile(dbConfig, configFilePath)
	if err != nil {
		return err
	}

	return nil
}

// SaveTomlFile will open and save data to toml file
// TODO: move to core
func SaveTomlFile(dest interface{}, relativePath string) error {
	f, err := os.Create(relativePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return toml.NewEncoder(f).Encode(dest)
}

func (pf *PersisterFactory) getPersisterConfigFilePath(path string) string {
	return filepath.Join(
		path,
		"dbConfig.toml",
	)
}

func (pf *PersisterFactory) createShardIDProvider() (storage.ShardIDProvider, error) {
	switch storageunit.ShardIDProviderType(pf.shardIDProviderType) {
	case storageunit.BinarySplit:
		return database.NewShardIDProvider(pf.numShards)
	default:
		return nil, storage.ErrNotSupportedShardIDProviderType
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
