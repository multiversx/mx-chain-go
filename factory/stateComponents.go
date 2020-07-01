package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	trieFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

//TODO: merge this with data components

// StateComponentsFactoryArgs holds the arguments needed for creating a state components factory
type StateComponentsFactoryArgs struct {
	Config           config.Config
	ShardCoordinator sharding.Coordinator
	Core             CoreComponentsHolder
}

type stateComponentsFactory struct {
	config           config.Config
	shardCoordinator sharding.Coordinator
	core             CoreComponentsHolder
}

// stateComponents struct holds the state components of the Elrond protocol
type stateComponents struct {
	peerAccounts        state.AccountsAdapter
	accountsAdapter     state.AccountsAdapter
	triesContainer      state.TriesHolder
	trieStorageManagers map[string]data.StorageManager
}

// NewStateComponentsFactory will return a new instance of stateComponentsFactory
func NewStateComponentsFactory(args StateComponentsFactoryArgs) (*stateComponentsFactory, error) {
	if args.Core == nil {
		return nil, ErrNilCoreComponents
	}
	if check.IfNil(args.Core.Hasher()) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.Core.InternalMarshalizer()) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Core.PathHandler()) {
		return nil, ErrNilPathManager
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}

	return &stateComponentsFactory{
		config:           args.Config,
		core:             args.Core,
		shardCoordinator: args.ShardCoordinator,
	}, nil
}

// Create creates the state components
func (scf *stateComponentsFactory) Create() (*stateComponents, error) {
	accountFactory := factoryState.NewAccountCreator()
	triesContainer, triesStorageManagers, err := scf.createTries()
	if err != nil {
		return nil, err
	}

	merkleTrie := triesContainer.Get([]byte(trieFactory.UserAccountTrie))
	accountsAdapter, err := state.NewAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrAccountsAdapterCreation, err.Error())
	}

	accountFactory = factoryState.NewPeerAccountCreator()
	merkleTrie = triesContainer.Get([]byte(trieFactory.PeerAccountTrie))
	peerAdapter, err := state.NewPeerAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, err
	}

	return &stateComponents{
		peerAccounts:        peerAdapter,
		accountsAdapter:     accountsAdapter,
		triesContainer:      triesContainer,
		trieStorageManagers: triesStorageManagers,
	}, nil
}

func (scf *stateComponentsFactory) createTries() (state.TriesHolder, map[string]data.StorageManager, error) {
	trieContainer := state.NewDataTriesHolder()
	trieFactoryArgs := trieFactory.TrieFactoryArgs{
		EvictionWaitingListCfg:   scf.config.EvictionWaitingList,
		SnapshotDbCfg:            scf.config.TrieSnapshotDB,
		Marshalizer:              scf.core.InternalMarshalizer(),
		Hasher:                   scf.core.Hasher(),
		PathManager:              scf.core.PathHandler(),
		TrieStorageManagerConfig: scf.config.TrieStorageManagerConfig,
	}
	shardIDString := convertShardIDToString(scf.shardCoordinator.SelfId())

	trieFactoryObj, err := trieFactory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	trieStorageManagers := make(map[string]data.StorageManager)
	userStorageManager, userAccountTrie, err := trieFactoryObj.Create(
		scf.config.AccountsTrieStorage,
		shardIDString,
		scf.config.StateTriesConfig.AccountsStatePruningEnabled,
		scf.config.StateTriesConfig.MaxStateTrieLevelInMemory,
	)
	if err != nil {
		return nil, nil, err
	}
	trieContainer.Put([]byte(trieFactory.UserAccountTrie), userAccountTrie)
	trieStorageManagers[trieFactory.UserAccountTrie] = userStorageManager

	peerStorageManager, peerAccountsTrie, err := trieFactoryObj.Create(
		scf.config.PeerAccountsTrieStorage,
		shardIDString,
		scf.config.StateTriesConfig.PeerStatePruningEnabled,
		scf.config.StateTriesConfig.MaxPeerTrieLevelInMemory,
	)
	if err != nil {
		return nil, nil, err
	}
	trieContainer.Put([]byte(trieFactory.PeerAccountTrie), peerAccountsTrie)
	trieStorageManagers[trieFactory.PeerAccountTrie] = peerStorageManager

	return trieContainer, trieStorageManagers, nil
}

func convertShardIDToString(shardID uint32) string {
	if shardID == core.MetachainShardId {
		return "metachain"
	}

	return fmt.Sprintf("%d", shardID)
}
