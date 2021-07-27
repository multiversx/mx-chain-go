package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	factoryState "github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
)

//TODO: merge this with data components

// StateComponentsFactoryArgs holds the arguments needed for creating a state components factory
type StateComponentsFactoryArgs struct {
	Config              config.Config
	ShardCoordinator    sharding.Coordinator
	Core                CoreComponentsHolder
	TriesContainer      state.TriesHolder
	TrieStorageManagers map[string]temporary.StorageManager
}

type stateComponentsFactory struct {
	config              config.Config
	shardCoordinator    sharding.Coordinator
	core                CoreComponentsHolder
	triesContainer      state.TriesHolder
	trieStorageManagers map[string]temporary.StorageManager
}

// stateComponents struct holds the state components of the Elrond protocol
type stateComponents struct {
	peerAccounts        state.AccountsAdapter
	accountsAdapter     state.AccountsAdapter
	accountsAdapterAPI  state.AccountsAdapter
	triesContainer      state.TriesHolder
	trieStorageManagers map[string]temporary.StorageManager
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
	accountsAdapter, accountsAdapterAPI, err := scf.createAccountsAdapters()
	if err != nil {
		return nil, err
	}

	peerAdapter, err := scf.createPeerAdapter()
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

func (scf *stateComponentsFactory) createAccountsAdapters() (state.AccountsAdapter, state.AccountsAdapter, error) {
	accountFactory := factoryState.NewAccountCreator()
	merkleTrie := scf.triesContainer.Get([]byte(trieFactory.UserAccountTrie))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, nil, err
	}

	accountsAdapter, err := state.NewAccountsDB(
		merkleTrie,
		scf.core.Hasher(),
		scf.core.InternalMarshalizer(),
		accountFactory,
		storagePruning,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	accountsAdapterAPI, err := state.NewAccountsDB(
		merkleTrie,
		scf.core.Hasher(),
		scf.core.InternalMarshalizer(),
		accountFactory,
		storagePruning,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("accounts adapter API: %w: %s", errors.ErrAccountsAdapterCreation, err.Error())
	}

	return accountsAdapter, accountsAdapterAPI, nil
}

func (scf *stateComponentsFactory) createPeerAdapter() (state.AccountsAdapter, error) {
	accountFactory := factoryState.NewPeerAccountCreator()
	merkleTrie := scf.triesContainer.Get([]byte(trieFactory.PeerAccountTrie))
	storagePruning, err := scf.newStoragePruningManager()
	if err != nil {
		return nil, err
	}

	peerAdapter, err := state.NewPeerAccountsDB(
		merkleTrie,
		scf.core.Hasher(),
		scf.core.InternalMarshalizer(),
		accountFactory,
		storagePruning,
	)
	if err != nil {
		return nil, err
	}

	return peerAdapter, nil
}

func (scf *stateComponentsFactory) newStoragePruningManager() (state.StoragePruningManager, error) {
	args := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: scf.config.EvictionWaitingList.RootHashesSize,
		HashesSize:     scf.config.EvictionWaitingList.HashesSize,
	}
	trieEvictionWaitingList, err := evictionWaitingList.NewMemoryEvictionWaitingList(args)
	if err != nil {
		return nil, err
	}

	storagePruning, err := storagePruningManager.NewStoragePruningManager(
		trieEvictionWaitingList,
		scf.config.TrieStorageManagerConfig.PruningBufferLen,
	)
	if err != nil {
		return nil, err
	}

	return storagePruning, nil
}

// Close closes all underlying components that need closing
func (pc *stateComponents) Close() error {
	errString := ""

	err := pc.accountsAdapter.Close()
	if err != nil {
		errString += fmt.Errorf("accountsAdapter close failed: %w ", err).Error()
	}

	err = pc.accountsAdapterAPI.Close()
	if err != nil {
		errString += fmt.Errorf("accountsAdapterAPI close failed: %w ", err).Error()
	}

	err = pc.peerAccounts.Close()
	if err != nil {
		errString += fmt.Errorf("peerAccounts close failed: %w ", err).Error()
	}

	if len(errString) != 0 {
		return fmt.Errorf("state components close failed: %s", errString)
	}
	return nil
}
