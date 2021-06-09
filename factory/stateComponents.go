package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
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
	accountsAdapterAPI  state.AccountsAdapter
	triesContainer      state.TriesHolder
	trieStorageManagers map[string]data.StorageManager
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

	accountsAdapterAPI, err := state.NewAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, fmt.Errorf("accounts adapter API: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountFactory = factoryState.NewPeerAccountCreator()
	merkleTrie = scf.triesContainer.Get([]byte(trieFactory.PeerAccountTrie))
	peerAdapter, err := state.NewPeerAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, err
	}

	return &stateComponents{
		peerAccounts:        peerAdapter,
		accountsAdapter:     accountsAdapter,
		accountsAdapterAPI:  accountsAdapterAPI,
		triesContainer:      scf.triesContainer,
		trieStorageManagers: scf.trieStorageManagers,
	}, nil
}

// Close closes all underlying components that need closing
func (pc *stateComponents) Close() error {
	return nil
}
