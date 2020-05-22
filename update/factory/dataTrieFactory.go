package factory

import (
	"path"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

// ArgsNewDataTrieFactory is the argument structure for the new data trie factory
type ArgsNewDataTrieFactory struct {
	StorageConfig        config.StorageConfig
	SyncFolder           string
	Marshalizer          marshal.Marshalizer
	Hasher               hashing.Hasher
	ShardCoordinator     sharding.Coordinator
	MaxTrieLevelInMemory uint
}

type dataTrieFactory struct {
	shardCoordinator     sharding.Coordinator
	trieStorage          data.StorageManager
	marshalizer          marshal.Marshalizer
	hasher               hashing.Hasher
	maxTrieLevelInMemory uint
}

// NewDataTrieFactory creates a data trie factory
func NewDataTrieFactory(args ArgsNewDataTrieFactory) (*dataTrieFactory, error) {
	if len(args.SyncFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}

	dbConfig := storageFactory.GetDBFromConfig(args.StorageConfig.DB)
	dbConfig.FilePath = path.Join(args.SyncFolder, args.StorageConfig.DB.FilePath)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(args.StorageConfig.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(args.StorageConfig.Bloom),
	)
	if err != nil {
		return nil, err
	}

	trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
	if err != nil {
		return nil, err
	}

	d := &dataTrieFactory{
		shardCoordinator:     args.ShardCoordinator,
		trieStorage:          trieStorage,
		marshalizer:          args.Marshalizer,
		hasher:               args.Hasher,
		maxTrieLevelInMemory: args.MaxTrieLevelInMemory,
	}

	return d, nil
}

// TrieStorageManager returns trie storage manager
func (d *dataTrieFactory) TrieStorageManager() data.StorageManager {
	return d.trieStorage
}

// Create creates a TriesHolder container to hold all the states
func (d *dataTrieFactory) Create() (state.TriesHolder, error) {
	container := state.NewDataTriesHolder()

	for i := uint32(0); i < d.shardCoordinator.NumberOfShards(); i++ {
		err := d.createAndAddOneTrie(i, genesis.UserAccount, container)
		if err != nil {
			return nil, err
		}
	}

	err := d.createAndAddOneTrie(core.MetachainShardId, genesis.UserAccount, container)
	if err != nil {
		return nil, err
	}

	err = d.createAndAddOneTrie(core.MetachainShardId, genesis.ValidatorAccount, container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (d *dataTrieFactory) createAndAddOneTrie(shId uint32, accType genesis.Type, container state.TriesHolder) error {
	dataTrie, err := trie.NewTrie(d.trieStorage, d.marshalizer, d.hasher, d.maxTrieLevelInMemory)
	if err != nil {
		return err
	}

	trieId := genesis.CreateTrieIdentifier(shId, accType)
	container.Put([]byte(trieId), dataTrie)

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (d *dataTrieFactory) IsInterfaceNil() bool {
	return d == nil
}
