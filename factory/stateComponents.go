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
)

//TODO: merge this with data components

// StateComponentsFactoryArgs holds the arguments needed for creating a state components factory
type StateComponentsFactoryArgs struct {
	Config              config.Config
	BootstrapComponents BootstrapComponentsHolder
	Core                CoreComponentsHolder
}

type stateComponentsFactory struct {
	config         config.Config
	bootComponents BootstrapComponentsHolder
	core           CoreComponentsHolder
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
	if check.IfNil(args.BootstrapComponents) {
		return nil, errors.ErrNilBootstrapComponentsHolder
	}
	if check.IfNil(args.BootstrapComponents.ShardCoordinator()) {
		return nil, errors.ErrNilShardCoordinator
	}
	if check.IfNil(args.BootstrapComponents.EpochStartBootstrapper()) {
		return nil, errors.ErrNilTriesContainer
	}

	return &stateComponentsFactory{
		config:         args.Config,
		core:           args.Core,
		bootComponents: args.BootstrapComponents,
	}, nil
}

// Create creates the state components
func (scf *stateComponentsFactory) Create() (*stateComponents, error) {
	accountFactory := factoryState.NewAccountCreator()
	triesComponents, trieStorageManagers := scf.bootComponents.EpochStartBootstrapper().GetTriesComponents()
	if len(trieStorageManagers) == 0 {
		return nil, errors.ErrNilTriesStorageManagers
	}
	for _, storageManager := range trieStorageManagers {
		if check.IfNil(storageManager) {
			return nil, errors.ErrNilTrieStorageManager
		}
	}

	merkleTrie := triesComponents.Get([]byte(trieFactory.UserAccountTrie))
	accountsAdapter, err := state.NewAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountFactory = factoryState.NewPeerAccountCreator()
	merkleTrie = triesComponents.Get([]byte(trieFactory.PeerAccountTrie))
	peerAdapter, err := state.NewPeerAccountsDB(merkleTrie, scf.core.Hasher(), scf.core.InternalMarshalizer(), accountFactory)
	if err != nil {
		return nil, err
	}

	return &stateComponents{
		peerAccounts:        peerAdapter,
		accountsAdapter:     accountsAdapter,
		triesContainer:      triesComponents,
		trieStorageManagers: trieStorageManagers,
	}, nil
}

// Close closes all underlying components that need closing
func (pc *stateComponents) Close() error {
	return nil
}
