package state

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/state"
)

var _ factory.ComponentHandler = (*managedStateComponents)(nil)
var _ factory.StateComponentsHolder = (*managedStateComponents)(nil)
var _ factory.StateComponentsHandler = (*managedStateComponents)(nil)

type managedStateComponents struct {
	*stateComponents
	factory            *stateComponentsFactory
	mutStateComponents sync.RWMutex
}

// NewManagedStateComponents returns a news instance of managedStateComponents
func NewManagedStateComponents(scf *stateComponentsFactory) (*managedStateComponents, error) {
	if scf == nil {
		return nil, errors.ErrNilStateComponentsFactory
	}

	return &managedStateComponents{
		stateComponents: nil,
		factory:         scf,
	}, nil
}

// Create will create the managed components
func (msc *managedStateComponents) Create() error {
	sc, err := msc.factory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrStateComponentsFactoryCreate, err)
	}

	msc.mutStateComponents.Lock()
	msc.stateComponents = sc
	msc.mutStateComponents.Unlock()

	return nil
}

// Close will close all underlying sub-components
func (msc *managedStateComponents) Close() error {
	msc.mutStateComponents.Lock()
	defer msc.mutStateComponents.Unlock()

	if msc.stateComponents == nil {
		return nil
	}

	err := msc.stateComponents.Close()
	if err != nil {
		return err
	}
	msc.stateComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (msc *managedStateComponents) CheckSubcomponents() error {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return errors.ErrNilStateComponents
	}
	if check.IfNil(msc.peerAccounts) {
		return errors.ErrNilPeerAccounts
	}
	if check.IfNil(msc.accountsAdapter) {
		return errors.ErrNilAccountsAdapter
	}
	if check.IfNil(msc.triesContainer) {
		return errors.ErrNilTriesContainer
	}
	if len(msc.trieStorageManagers) == 0 {
		return errors.ErrNilStorageManagers
	}
	for _, trieStorageManager := range msc.trieStorageManagers {
		if check.IfNil(trieStorageManager) {
			return errors.ErrNilTrieStorageManager
		}
	}
	if check.IfNil(msc.missingTrieNodesNotifier) {
		return errors.ErrNilMissingTrieNodesNotifier
	}
	if check.IfNil(msc.trieLeavesRetriever) {
		return errors.ErrNilTrieLeavesRetriever
	}

	return nil
}

// PeerAccounts returns the accounts adapter for the validators
func (msc *managedStateComponents) PeerAccounts() state.AccountsAdapter {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.peerAccounts
}

// AccountsAdapter returns the accounts adapter for the user accounts
func (msc *managedStateComponents) AccountsAdapter() state.AccountsAdapter {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.accountsAdapter
}

// AccountsAdapterAPI returns the accounts adapter for the user accounts to be used in REST API
func (msc *managedStateComponents) AccountsAdapterAPI() state.AccountsAdapter {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.accountsAdapterAPI
}

// AccountsRepository returns the accounts adapter for the user accounts to be used in REST API
func (msc *managedStateComponents) AccountsRepository() state.AccountsRepository {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.accountsRepository
}

// TriesContainer returns the tries container
func (msc *managedStateComponents) TriesContainer() common.TriesHolder {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.triesContainer
}

// TrieStorageManagers returns the trie storage manager for the given account type
func (msc *managedStateComponents) TrieStorageManagers() map[string]common.StorageManager {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	retMap := make(map[string]common.StorageManager)

	// give back a map copy
	for key, val := range msc.stateComponents.trieStorageManagers {
		retMap[key] = val
	}

	return retMap
}

// SetTriesContainer sets the internal tries container to the one given as parameter
func (msc *managedStateComponents) SetTriesContainer(triesContainer common.TriesHolder) error {
	if check.IfNil(triesContainer) {
		return errors.ErrNilTriesContainer
	}

	msc.mutStateComponents.Lock()
	msc.stateComponents.triesContainer = triesContainer
	msc.mutStateComponents.Unlock()

	return nil
}

// SetTriesStorageManagers sets the internal map with the given parameter
func (msc *managedStateComponents) SetTriesStorageManagers(managers map[string]common.StorageManager) error {
	if len(managers) == 0 {
		return errors.ErrNilTriesStorageManagers
	}

	msc.mutStateComponents.Lock()
	msc.stateComponents.trieStorageManagers = managers
	msc.mutStateComponents.Unlock()

	return nil
}

// MissingTrieNodesNotifier returns the missing trie nodes notifier
func (msc *managedStateComponents) MissingTrieNodesNotifier() common.MissingTrieNodesNotifier {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.missingTrieNodesNotifier
}

// TrieLeavesRetriever returns the trie leaves retriever
func (msc *managedStateComponents) TrieLeavesRetriever() common.TrieLeavesRetriever {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.trieLeavesRetriever
}

// IsInterfaceNil returns true if the interface is nil
func (msc *managedStateComponents) IsInterfaceNil() bool {
	return msc == nil
}

// String returns the name of the component
func (msc *managedStateComponents) String() string {
	return factory.StateComponentsName
}
