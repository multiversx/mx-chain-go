package factory

import (
	"path"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
	containers "github.com/ElrondNetwork/elrond-go/update/container"
)

type ArgsNewTrieSyncersContainerFactory struct {
	StorageConfig      config.StorageConfig
	CacheConfig        config.CacheConfig
	SyncFolder         string
	ResolversContainer dataRetriever.ResolversContainer
	Marshalizer        marshal.Marshalizer
	Hasher             hashing.Hasher
	ShardCoordinator   sharding.Coordinator
}

type trieSyncersContainerFactory struct {
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	shardCoordinator   sharding.Coordinator
	trieStorage        data.StorageManager
	trieCacher         storage.Cacher
	resolversContainer dataRetriever.ResolversContainer
}

func NewTrieSyncersContainerFactory(args ArgsNewTrieSyncersContainerFactory) (*trieSyncersContainerFactory, error) {
	if len(args.SyncFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.ResolversContainer) {
		return nil, dataRetriever.ErrNilResolverContainer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, sharding.ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, data.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, sharding.ErrNilHasher
	}

	dbConfig := storageFactory.GetDBFromConfig(args.StorageConfig.DB)
	dbConfig.FilePath = path.Join(args.SyncFolder, "syncTries")
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

	trieCacher, err := storageUnit.NewCache(storageUnit.CacheType(args.CacheConfig.Type),
		args.CacheConfig.Size,
		args.CacheConfig.Shards,
	)
	if err != nil {
		return nil, err
	}

	t := &trieSyncersContainerFactory{
		marshalizer:        args.Marshalizer,
		hasher:             args.Hasher,
		shardCoordinator:   args.ShardCoordinator,
		trieStorage:        trieStorage,
		trieCacher:         trieCacher,
		resolversContainer: args.ResolversContainer,
	}

	return t, nil
}

func (t *trieSyncersContainerFactory) Create() (update.TrieSyncContainer, error) {
	container := containers.NewTrieSyncersContainer()

	for i := uint32(0); i < t.shardCoordinator.NumberOfShards(); i++ {
		err := t.createOneTrieSyncer(i, factory.UserAccount, container)
		if err != nil {
			return nil, err
		}
	}

	err := t.createOneTrieSyncer(sharding.MetachainShardId, factory.UserAccount, container)
	if err != nil {
		return nil, err
	}

	err = t.createOneTrieSyncer(sharding.MetachainShardId, factory.ValidatorAccount, container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (t *trieSyncersContainerFactory) createOneTrieSyncer(
	shId uint32,
	accType factory.Type,
	container update.TrieSyncContainer,
) error {
	trieId := update.CreateTrieIdentifier(shId, accType)
	resolver, err := t.resolversContainer.Get(trieId)
	if err != nil {
		return err
	}

	dataTrie, err := trie.NewTrie(t.trieStorage, t.marshalizer, t.hasher)
	if err != nil {
		return err
	}

	trieSyncer, err := trie.NewTrieSyncer(resolver, t.trieCacher, dataTrie, time.Minute)
	if err != nil {
		return err
	}

	err = container.Add(trieId, trieSyncer)
	if err != nil {
		return err
	}

	return nil
}

func (t *trieSyncersContainerFactory) IsInterfaceNil() bool {
	return t == nil
}
