package factory

import (
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type trieCreator struct {
	evictionWaitingListCfg config.EvictionWaitingListConfig
	snapshotDbCfg          config.DBConfig
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	pathManager            storage.PathManagerHandler
	shardId                string
}

var log = logger.GetOrCreate("trie")

// NewTrieFactory creates a new trie factory
func NewTrieFactory(
	args TrieFactoryArgs,
) (*trieCreator, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, trie.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, trie.ErrNilHasher
	}
	if check.IfNil(args.PathManager) {
		return nil, trie.ErrNilPathManager
	}

	return &trieCreator{
		evictionWaitingListCfg: args.EvictionWaitingListCfg,
		snapshotDbCfg:          args.SnapshotDbCfg,
		marshalizer:            args.Marshalizer,
		hasher:                 args.Hasher,
		pathManager:            args.PathManager,
		shardId:                args.ShardId,
	}, nil
}

// Create creates a new trie
func (tc *trieCreator) Create(trieStorageCfg config.StorageConfig, pruningEnabled bool) (data.Trie, error) {
	trieStoragePath, mainDb := path.Split(tc.pathManager.PathForStatic(tc.shardId, trieStorageCfg.DB.FilePath))

	dbConfig := factory.GetDBFromConfig(trieStorageCfg.DB)
	dbConfig.FilePath = path.Join(trieStoragePath, mainDb)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(trieStorageCfg.Cache),
		dbConfig,
		factory.GetBloomFromConfig(trieStorageCfg.Bloom),
	)
	if err != nil {
		return nil, err
	}

	log.Trace("trie pruning status", "enabled", pruningEnabled)
	if !pruningEnabled {
		trieStorage, errNewTrie := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
		if errNewTrie != nil {
			return nil, errNewTrie
		}

		return trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher)
	}

	evictionDb, err := storageUnit.NewDB(
		storageUnit.DBType(tc.evictionWaitingListCfg.DB.Type),
		filepath.Join(trieStoragePath, tc.evictionWaitingListCfg.DB.FilePath),
		tc.evictionWaitingListCfg.DB.MaxBatchSize,
		tc.evictionWaitingListCfg.DB.BatchDelaySeconds,
		tc.evictionWaitingListCfg.DB.MaxOpenFiles,
	)
	if err != nil {
		return nil, err
	}

	ewl, err := evictionWaitingList.NewEvictionWaitingList(tc.evictionWaitingListCfg.Size, evictionDb, tc.marshalizer)
	if err != nil {
		return nil, err
	}

	snapshotDbCfg := config.DBConfig{
		FilePath:          filepath.Join(trieStoragePath, tc.snapshotDbCfg.FilePath),
		Type:              tc.snapshotDbCfg.Type,
		BatchDelaySeconds: tc.snapshotDbCfg.BatchDelaySeconds,
		MaxBatchSize:      tc.snapshotDbCfg.MaxBatchSize,
		MaxOpenFiles:      tc.snapshotDbCfg.MaxOpenFiles,
	}

	trieStorage, err := trie.NewTrieStorageManager(accountsTrieStorage, snapshotDbCfg, ewl)
	if err != nil {
		return nil, err
	}

	return trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *trieCreator) IsInterfaceNil() bool {
	return tc == nil
}
