package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-storage-go/factory"
)

const minNumShards = 2

// persisterCreator is the factory which will handle creating new persisters
type persisterCreator struct {
	dbType              string
	batchDelaySeconds   int
	maxBatchSize        int
	maxOpenFiles        int
	shardIDProviderType string
	numShards           int32
}

func newPersisterCreator(config config.DBConfig) *persisterCreator {
	return &persisterCreator{
		dbType:              config.Type,
		batchDelaySeconds:   config.BatchDelaySeconds,
		maxBatchSize:        config.MaxBatchSize,
		maxOpenFiles:        config.MaxOpenFiles,
		shardIDProviderType: config.ShardIDProviderType,
		numShards:           config.NumShards,
	}
}

// Create will create the persister for the provided path
// TODO: refactor to use max tries mechanism
func (pc *persisterCreator) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, storage.ErrInvalidFilePath
	}

	if pc.numShards < minNumShards {
		return pc.CreateBasePersister(path)
	}

	shardIDProvider, err := pc.createShardIDProvider()
	if err != nil {
		return nil, err
	}
	return database.NewShardedPersister(path, pc, shardIDProvider)
}

// CreateBasePersister will create base the persister for the provided path
func (pc *persisterCreator) CreateBasePersister(path string) (storage.Persister, error) {
	var dbType = storageunit.DBType(pc.dbType)

	argsDB := factory.ArgDB{
		DBType:            dbType,
		Path:              path,
		BatchDelaySeconds: pc.batchDelaySeconds,
		MaxBatchSize:      pc.maxBatchSize,
		MaxOpenFiles:      pc.maxOpenFiles,
	}

	return storageunit.NewDB(argsDB)
}

func (pc *persisterCreator) createShardIDProvider() (storage.ShardIDProvider, error) {
	switch storageunit.ShardIDProviderType(pc.shardIDProviderType) {
	case storageunit.BinarySplit:
		return database.NewShardIDProvider(pc.numShards)
	default:
		return nil, storage.ErrNotSupportedShardIDProviderType
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pc *persisterCreator) IsInterfaceNil() bool {
	return pc == nil
}
