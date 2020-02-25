package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
	containers "github.com/ElrondNetwork/elrond-go/update/container"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

// ArgsNewTrieSyncersContainerFactory defines the arguments needed to create trie syncers container
type ArgsNewTrieSyncersContainerFactory struct {
	CacheConfig        config.CacheConfig
	SyncFolder         string
	ResolversContainer dataRetriever.ResolversContainer
	DataTrieContainer  state.TriesHolder
	ShardCoordinator   sharding.Coordinator
}

type trieSyncersContainerFactory struct {
	shardCoordinator   sharding.Coordinator
	trieCacher         storage.Cacher
	trieContainer      state.TriesHolder
	resolversContainer dataRetriever.ResolversContainer
}

// NewTrieSyncersContainerFactory creates a factory for trie syncers container
func NewTrieSyncersContainerFactory(args ArgsNewTrieSyncersContainerFactory) (*trieSyncersContainerFactory, error) {
	if len(args.SyncFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.ResolversContainer) {
		return nil, update.ErrNilResolverContainer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.DataTrieContainer) {
		return nil, update.ErrNilDataTrieContainer
	}

	trieCacher, err := storageUnit.NewCache(
		storageUnit.CacheType(args.CacheConfig.Type),
		args.CacheConfig.Size,
		args.CacheConfig.Shards,
	)
	if err != nil {
		return nil, err
	}

	t := &trieSyncersContainerFactory{
		shardCoordinator:   args.ShardCoordinator,
		trieCacher:         trieCacher,
		resolversContainer: args.ResolversContainer,
		trieContainer:      args.DataTrieContainer,
	}

	return t, nil
}

// Create creates all the needed syncers and returns the container
func (t *trieSyncersContainerFactory) Create() (update.TrieSyncContainer, error) {
	container := containers.NewTrieSyncersContainer()

	for i := uint32(0); i < t.shardCoordinator.NumberOfShards(); i++ {
		err := t.createOneTrieSyncer(i, factory.UserAccount, container)
		if err != nil {
			return nil, err
		}
	}

	err := t.createOneTrieSyncer(core.MetachainShardId, factory.UserAccount, container)
	if err != nil {
		return nil, err
	}

	err = t.createOneTrieSyncer(core.MetachainShardId, factory.ValidatorAccount, container)
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
	trieId := genesis.CreateTrieIdentifier(shId, accType)
	resolver, err := t.resolversContainer.Get(trieId)
	if err != nil {
		return err
	}

	dataTrie := t.trieContainer.Get([]byte(trieId))
	if check.IfNil(dataTrie) {
		return update.ErrNilDataTrieContainer
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

// IsInterfaceNil returns true if the underlying object is nil
func (t *trieSyncersContainerFactory) IsInterfaceNil() bool {
	return t == nil
}
