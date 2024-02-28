package factory

import (
	"os"
	"path"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-storage-go/factory"
)

const minNumShards = 2
const pathSeparator = "/"

// persisterCreator is the factory which will handle creating new persisters
type persisterCreator struct {
	conf config.DBConfig
}

func newPersisterCreator(config config.DBConfig) *persisterCreator {
	return &persisterCreator{
		conf: config,
	}
}

// Create will create the persister for the provided path
func (pc *persisterCreator) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, storage.ErrInvalidFilePath
	}

	if pc.conf.UseTmpAsFilePath {
		filePath, err := getTmpFilePath(path)
		if err != nil {
			return nil, err
		}

		path = filePath
	}

	if pc.conf.NumShards < minNumShards {
		return pc.CreateBasePersister(path)
	}

	shardIDProvider, err := pc.createShardIDProvider()
	if err != nil {
		return nil, err
	}
	return database.NewShardedPersister(path, pc, shardIDProvider)
}

func getTmpFilePath(p string) (string, error) {
	_, file := path.Split(p)
	return os.MkdirTemp("", file)
}

// CreateBasePersister will create base the persister for the provided path
func (pc *persisterCreator) CreateBasePersister(path string) (storage.Persister, error) {
	var dbType = storageunit.DBType(pc.conf.Type)

	argsDB := factory.ArgDB{
		DBType:            dbType,
		Path:              path,
		BatchDelaySeconds: pc.conf.BatchDelaySeconds,
		MaxBatchSize:      pc.conf.MaxBatchSize,
		MaxOpenFiles:      pc.conf.MaxOpenFiles,
	}

	return storageunit.NewDB(argsDB)
}

func (pc *persisterCreator) createShardIDProvider() (storage.ShardIDProvider, error) {
	switch storageunit.ShardIDProviderType(pc.conf.ShardIDProviderType) {
	case storageunit.BinarySplit:
		return database.NewShardIDProvider(pc.conf.NumShards)
	default:
		return nil, storage.ErrNotSupportedShardIDProviderType
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pc *persisterCreator) IsInterfaceNil() bool {
	return pc == nil
}
