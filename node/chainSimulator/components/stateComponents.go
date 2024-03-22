package components

import (
	"io"

	chainData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
	factoryState "github.com/multiversx/mx-chain-go/factory/state"
	"github.com/multiversx/mx-chain-go/state"
)

// ArgsStateComponents will hold the components needed for state components
type ArgsStateComponents struct {
	Config         config.Config
	CoreComponents factory.CoreComponentsHolder
	StatusCore     factory.StatusCoreComponentsHolder
	StoreService   dataRetriever.StorageService
	ChainHandler   chainData.ChainHandler
}

type stateComponentsHolder struct {
	peerAccount              state.AccountsAdapter
	accountsAdapter          state.AccountsAdapter
	accountsAdapterAPI       state.AccountsAdapter
	accountsRepository       state.AccountsRepository
	triesContainer           common.TriesHolder
	triesStorageManager      map[string]common.StorageManager
	missingTrieNodesNotifier common.MissingTrieNodesNotifier
	stateComponentsCloser    io.Closer
}

// CreateStateComponents will create the state components holder
func CreateStateComponents(args ArgsStateComponents) (*stateComponentsHolder, error) {
	stateComponentsFactory, err := factoryState.NewStateComponentsFactory(factoryState.StateComponentsFactoryArgs{
		Config:                   args.Config,
		Core:                     args.CoreComponents,
		StatusCore:               args.StatusCore,
		StorageService:           args.StoreService,
		ProcessingMode:           common.Normal,
		ShouldSerializeSnapshots: false,
		ChainHandler:             args.ChainHandler,
	})
	if err != nil {
		return nil, err
	}

	stateComp, err := factoryState.NewManagedStateComponents(stateComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = stateComp.Create()
	if err != nil {
		return nil, err
	}

	err = stateComp.CheckSubcomponents()
	if err != nil {
		return nil, err
	}

	return &stateComponentsHolder{
		peerAccount:              stateComp.PeerAccounts(),
		accountsAdapter:          stateComp.AccountsAdapter(),
		accountsAdapterAPI:       stateComp.AccountsAdapterAPI(),
		accountsRepository:       stateComp.AccountsRepository(),
		triesContainer:           stateComp.TriesContainer(),
		triesStorageManager:      stateComp.TrieStorageManagers(),
		missingTrieNodesNotifier: stateComp.MissingTrieNodesNotifier(),
		stateComponentsCloser:    stateComp,
	}, nil
}

// PeerAccounts will return peer accounts
func (s *stateComponentsHolder) PeerAccounts() state.AccountsAdapter {
	return s.peerAccount
}

// AccountsAdapter will return accounts adapter
func (s *stateComponentsHolder) AccountsAdapter() state.AccountsAdapter {
	return s.accountsAdapter
}

// AccountsAdapterAPI will return accounts adapter api
func (s *stateComponentsHolder) AccountsAdapterAPI() state.AccountsAdapter {
	return s.accountsAdapterAPI
}

// AccountsRepository will return accounts repository
func (s *stateComponentsHolder) AccountsRepository() state.AccountsRepository {
	return s.accountsRepository
}

// TriesContainer will return tries container
func (s *stateComponentsHolder) TriesContainer() common.TriesHolder {
	return s.triesContainer
}

// TrieStorageManagers will return trie storage managers
func (s *stateComponentsHolder) TrieStorageManagers() map[string]common.StorageManager {
	return s.triesStorageManager
}

// MissingTrieNodesNotifier will return missing trie nodes notifier
func (s *stateComponentsHolder) MissingTrieNodesNotifier() common.MissingTrieNodesNotifier {
	return s.missingTrieNodesNotifier
}

// Close will close the state components
func (s *stateComponentsHolder) Close() error {
	return s.stateComponentsCloser.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *stateComponentsHolder) IsInterfaceNil() bool {
	return s == nil
}

// Create will do nothing
func (s *stateComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (s *stateComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (s *stateComponentsHolder) String() string {
	return ""
}
