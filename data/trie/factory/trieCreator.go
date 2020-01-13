package factory

import (
	"path"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

type trieCreator struct {
	trieStorage data.StorageManager
	msh         marshal.Marshalizer
	hsh         hashing.Hasher
}

// NewTrieFactory creates a new trie factory
func NewTrieFactory(
	args *trieFactoryArgs,
) (*trieCreator, error) {
	if check.IfNil(args.marshalizer) {
		return nil, trie.ErrNilMarshalizer
	}
	if check.IfNil(args.hasher) {
		return nil, trie.ErrNilHasher
	}
	if check.IfNil(args.pathManager) {
		return nil, trie.ErrNilPathManager
	}

	trieStoragePath, mainDb := path.Split(args.pathManager.PathForStatic(args.shardId, args.cfg.DB.FilePath))

	dbConfig := storageFactory.GetDBFromConfig(args.cfg.DB)
	dbConfig.FilePath = path.Join(trieStoragePath, mainDb)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(args.cfg.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(args.cfg.Bloom),
	)
	if err != nil {
		return nil, err
	}

	if !args.pruningEnabled {
		trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
		if err != nil {
			return nil, err
		}

		return &trieCreator{
			trieStorage: trieStorage,
			msh:         args.marshalizer,
			hsh:         args.hasher,
		}, nil
	}

	evictionDb, err := storageUnit.NewDB(
		storageUnit.DBType(args.evictionWaitingListCfg.DB.Type),
		filepath.Join(trieStoragePath, args.evictionWaitingListCfg.DB.FilePath),
		args.evictionWaitingListCfg.DB.MaxBatchSize,
		args.evictionWaitingListCfg.DB.BatchDelaySeconds,
		args.evictionWaitingListCfg.DB.MaxOpenFiles,
	)
	if err != nil {
		return nil, err
	}

	ewl, err := evictionWaitingList.NewEvictionWaitingList(args.evictionWaitingListCfg.Size, evictionDb, args.marshalizer)
	if err != nil {
		return nil, err
	}

	args.snapshotDbCfg.FilePath = filepath.Join(trieStoragePath, args.snapshotDbCfg.FilePath)

	trieStorage, err := trie.NewTrieStorageManager(accountsTrieStorage, &args.snapshotDbCfg, ewl)
	if err != nil {
		return nil, err
	}

	return &trieCreator{
		trieStorage: trieStorage,
		msh:         args.marshalizer,
		hsh:         args.hasher,
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
