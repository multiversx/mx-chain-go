package factory

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	trieFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

//TODO: merge this with data components

// StateComponentsFactoryArgs holds the arguments needed for creating a state components factory
type StateComponentsFactoryArgs struct {
	Config              config.Config
	ShardCoordinator    sharding.Coordinator
	Core                CoreComponentsHolder
	TriesContainer      state.TriesHolder
	TrieStorageManagers map[string]data.StorageManager
}

type stateComponentsFactory struct {
	config              config.Config
	shardCoordinator    sharding.Coordinator
	core                CoreComponentsHolder
	triesContainer      state.TriesHolder
	trieStorageManagers map[string]data.StorageManager
}

// stateComponents struct holds the state components of the Elrond protocol
type stateComponents struct {
	peerAccounts        state.AccountsAdapter
	accountsAdapter     state.AccountsAdapter
	triesContainer      state.TriesHolder
	trieStorageManagers map[string]data.StorageManager
	closeFunc           func()
}

// NewStateComponentsFactory will return a new instance of stateComponentsFactory
func NewStateComponentsFactory(args StateComponentsFactoryArgs) (*stateComponentsFactory, error) {
	if args.Core == nil {
		return nil, errors.ErrNilCoreComponents
	}
	if check.IfNil(args.Core.Hasher()) {
		return nil, errors.ErrNilHasher
	}
	if check.IfNil(args.Core.InternalMarshalizer()) {
		return nil, errors.ErrNilMarshalizer
	}
	if check.IfNil(args.Core.PathHandler()) {
		return nil, errors.ErrNilPathHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.TriesContainer) {
		return nil, errors.ErrNilTriesContainer
	}
	if len(args.TrieStorageManagers) == 0 {
		return nil, errors.ErrNilTriesStorageManagers
	}
	for _, storageManager := range args.TrieStorageManagers {
		if check.IfNil(storageManager) {
			return nil, errors.ErrNilTrieStorageManager
		}
	}

	return &stateComponentsFactory{
		config:              args.Config,
		shardCoordinator:    args.ShardCoordinator,
		core:                args.Core,
		triesContainer:      args.TriesContainer,
		trieStorageManagers: args.TrieStorageManagers,
	}, nil
}

// Create creates the state components
func (scf *stateComponentsFactory) Create() (*stateComponents, error) {
	accountFactory := factoryState.NewAccountCreator()

	merkleTrie := scf.triesContainer.Get([]byte(trieFactory.UserAccountTrie))
	accountsAdapter, err := state.NewAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountFactory = factoryState.NewPeerAccountCreator()
	merkleTrie = scf.triesContainer.Get([]byte(trieFactory.PeerAccountTrie))
	peerAdapter, err := state.NewPeerAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, err
	}

	_, cancelFunc := context.WithCancel(context.Background())

	return &stateComponents{
		peerAccounts:        peerAdapter,
		accountsAdapter:     accountsAdapter,
		triesContainer:      scf.triesContainer,
		trieStorageManagers: scf.trieStorageManagers,
		closeFunc:           cancelFunc,
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

// Close closes all underlying components that need closing
func (pc *stateComponents) Close() error {
	pc.closeFunc()

	// TODO: close all components

	return nil
}
