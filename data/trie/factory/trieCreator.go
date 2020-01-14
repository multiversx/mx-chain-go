package factory

import (
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type trieCreator struct {
	trieStorage data.StorageManager
	msh         marshal.Marshalizer
	hsh         hashing.Hasher
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

	trieStoragePath, mainDb := path.Split(args.PathManager.PathForStatic(args.ShardId, args.Cfg.DB.FilePath))

	dbConfig := storageFactory.GetDBFromConfig(args.Cfg.DB)
	dbConfig.FilePath = path.Join(trieStoragePath, mainDb)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(args.Cfg.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(args.Cfg.Bloom),
	)
	if err != nil {
		return nil, err
	}

	log.Trace("trie pruning status", "enabled", args.PruningEnabled)
	if !args.PruningEnabled {
		trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
		if err != nil {
			return nil, err
		}

		return &trieCreator{
			trieStorage: trieStorage,
			msh:         args.Marshalizer,
			hsh:         args.Hasher,
		}, nil
	}

	evictionDb, err := storageUnit.NewDB(
		storageUnit.DBType(args.EvictionWaitingListCfg.DB.Type),
		filepath.Join(trieStoragePath, args.EvictionWaitingListCfg.DB.FilePath),
		args.EvictionWaitingListCfg.DB.MaxBatchSize,
		args.EvictionWaitingListCfg.DB.BatchDelaySeconds,
		args.EvictionWaitingListCfg.DB.MaxOpenFiles,
	)
	if err != nil {
		return nil, err
	}

	ewl, err := evictionWaitingList.NewEvictionWaitingList(args.EvictionWaitingListCfg.Size, evictionDb, args.Marshalizer)
	if err != nil {
		return nil, err
	}

	args.SnapshotDbCfg.FilePath = filepath.Join(trieStoragePath, args.SnapshotDbCfg.FilePath)

	trieStorage, err := trie.NewTrieStorageManager(accountsTrieStorage, &args.SnapshotDbCfg, ewl)
	if err != nil {
		return nil, err
	}

	return &trieCreator{
		trieStorage: trieStorage,
		msh:         args.Marshalizer,
		hsh:         args.Hasher,
	}, nil
}

// Create creates a new trie
func (tc *trieCreator) Create() (data.Trie, error) {
	return trie.NewTrie(tc.trieStorage, tc.msh, tc.hsh)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *trieCreator) IsInterfaceNil() bool {
	return tc == nil
}
