package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	factoryTrie "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	containers "github.com/ElrondNetwork/elrond-go/update/container"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

// ArgsNewTrieSyncersContainerFactory defines the arguments needed to create trie syncers container
type ArgsNewTrieSyncersContainerFactory struct {
	TrieCacher        storage.Cacher
	SyncFolder        string
	RequestHandler    update.RequestHandler
	DataTrieContainer state.TriesHolder
	ShardCoordinator  sharding.Coordinator
}

type trieSyncersContainerFactory struct {
	shardCoordinator sharding.Coordinator
	trieCacher       storage.Cacher
	trieContainer    state.TriesHolder
	requestHandler   update.RequestHandler
}

// NewTrieSyncersContainerFactory creates a factory for trie syncers container
func NewTrieSyncersContainerFactory(args ArgsNewTrieSyncersContainerFactory) (*trieSyncersContainerFactory, error) {
	if len(args.SyncFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.RequestHandler) {
		return nil, update.ErrNilRequestHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.DataTrieContainer) {
		return nil, update.ErrNilDataTrieContainer
	}
	if check.IfNil(args.TrieCacher) {
		return nil, update.ErrNilCacher
	}

	t := &trieSyncersContainerFactory{
		shardCoordinator: args.ShardCoordinator,
		trieCacher:       args.TrieCacher,
		requestHandler:   args.RequestHandler,
		trieContainer:    args.DataTrieContainer,
	}

	return t, nil
}

// Create creates all the needed syncers and returns the container
func (t *trieSyncersContainerFactory) Create() (update.TrieSyncContainer, error) {
	container := containers.NewTrieSyncersContainer()

	for i := uint32(0); i < t.shardCoordinator.NumberOfShards(); i++ {
		err := t.createOneTrieSyncer(i, state.UserAccount, container)
		if err != nil {
			return nil, err
		}
	}

	err := t.createOneTrieSyncer(core.MetachainShardId, state.UserAccount, container)
	if err != nil {
		return nil, err
	}

	err = t.createOneTrieSyncer(core.MetachainShardId, state.ValidatorAccount, container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (t *trieSyncersContainerFactory) createOneTrieSyncer(
	shId uint32,
	accType state.Type,
	container update.TrieSyncContainer,
) error {
	trieId := genesis.CreateTrieIdentifier(shId, accType)

	dataTrie := t.trieContainer.Get([]byte(trieId))
	if check.IfNil(dataTrie) {
		return update.ErrNilDataTrieContainer
	}

	trieSyncer, err := trie.NewTrieSyncer(t.requestHandler, t.trieCacher, dataTrie, time.Minute, shId, trieTopicFromAccountType(accType))
	if err != nil {
		return err
	}

	err = container.Add(trieId, trieSyncer)
	if err != nil {
		return err
	}

	return nil
}

func trieTopicFromAccountType(accType state.Type) string {
	switch accType {
	case state.UserAccount:
		return factoryTrie.AccountTrieNodesTopic
	case state.ValidatorAccount:
		return factoryTrie.ValidatorTrieNodesTopic
	}
	return ""
}

// IsInterfaceNil returns true if the underlying object is nil
func (t *trieSyncersContainerFactory) IsInterfaceNil() bool {
	return t == nil
}
