package factory

import (
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/trie"
)

// TODO remove this when it is no longer needed
type oldTrieStorageCreator struct {
	pathManager   storage.PathManagerHandler
	generalConfig config.Config

	createdStorages map[string]common.DBWriteCacher
}

// NewOldTrieStorageCreator creates a new trieStorageCreator
func NewOldTrieStorageCreator(
	pathManager storage.PathManagerHandler,
	generalConfig config.Config,
) (*oldTrieStorageCreator, error) {
	if check.IfNil(pathManager) {
		return nil, trie.ErrNilPathManager
	}

	return &oldTrieStorageCreator{
		pathManager:     pathManager,
		generalConfig:   generalConfig,
		createdStorages: make(map[string]common.DBWriteCacher),
	}, nil
}

// GetStorageForShard creates a new storage for the given shard, or returns an existing one
// if it was created previously
func (tsc *oldTrieStorageCreator) GetStorageForShard(shardId string, trieType string) (common.DBWriteCacher, error) {
	switch trieType {
	case UserAccountTrie:
		return tsc.getStorage(shardId, tsc.generalConfig.AccountsTrieStorageOld)
	case PeerAccountTrie:
		return tsc.getStorage(shardId, tsc.generalConfig.PeerAccountsTrieStorageOld)
	default:
		return nil, trie.ErrInvalidTrieType
	}
}

// GetSnapshotsConfig returns the snapshot config for the given shard and trie type
func (tsc *oldTrieStorageCreator) GetSnapshotsConfig(shardId string, trieType string) config.DBConfig {
	switch trieType {
	case UserAccountTrie:
		return tsc.getSnapshotsConfig(shardId, tsc.generalConfig.AccountsTrieStorageOld)
	case PeerAccountTrie:
		return tsc.getSnapshotsConfig(shardId, tsc.generalConfig.PeerAccountsTrieStorageOld)
	default:
		return config.DBConfig{}
	}
}

func (tsc *oldTrieStorageCreator) getSnapshotsConfig(shardId string, storageConfig config.StorageConfig) config.DBConfig {
	storagePath, _ := path.Split(tsc.pathManager.PathForStatic(shardId, storageConfig.DB.FilePath))
	return config.DBConfig{
		FilePath:          filepath.Join(storagePath, tsc.generalConfig.TrieSnapshotDB.FilePath),
		Type:              tsc.generalConfig.TrieSnapshotDB.Type,
		BatchDelaySeconds: tsc.generalConfig.TrieSnapshotDB.BatchDelaySeconds,
		MaxBatchSize:      tsc.generalConfig.TrieSnapshotDB.MaxBatchSize,
		MaxOpenFiles:      tsc.generalConfig.TrieSnapshotDB.MaxOpenFiles,
	}
}

func (tsc *oldTrieStorageCreator) getStorage(
	shardId string,
	storageConfig config.StorageConfig,
) (common.DBWriteCacher, error) {
	storagePath, mainDb := path.Split(tsc.pathManager.PathForStatic(shardId, storageConfig.DB.FilePath))

	existingStorage, ok := tsc.createdStorages[storagePath]
	if ok {
		return existingStorage, nil
	}

	dbConfig := factory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(storagePath, mainDb)
	trieStorage, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(storageConfig.Cache),
		dbConfig,
		factory.GetBloomFromConfig(storageConfig.Bloom),
	)
	if err != nil {
		return nil, err
	}

	tsc.createdStorages[storagePath] = trieStorage

	return trieStorage, nil
}

// Close closes all the underlying storages
func (tsc *oldTrieStorageCreator) Close() error {
	for _, trStorage := range tsc.createdStorages {
		err := trStorage.Close()
		if err != nil {
			log.Warn("could not close trie storage", "error", err.Error())
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tsc *oldTrieStorageCreator) IsInterfaceNil() bool {
	return tsc == nil
}
