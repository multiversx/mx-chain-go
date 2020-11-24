package factory

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/errors"
)

var _ ComponentHandler = (*managedStateComponents)(nil)
var _ StateComponentsHolder = (*managedStateComponents)(nil)
var _ StateComponentsHandler = (*managedStateComponents)(nil)

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
	msc.mutStateComponents.Lock()
	defer msc.mutStateComponents.Unlock()

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

// TriesContainer returns the tries container
func (msc *managedStateComponents) TriesContainer() state.TriesHolder {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	return msc.stateComponents.triesContainer
}

// TrieStorageManagers returns the trie storage manager for the given account type
func (msc *managedStateComponents) TrieStorageManagers() map[string]data.StorageManager {
	msc.mutStateComponents.RLock()
	defer msc.mutStateComponents.RUnlock()

	if msc.stateComponents == nil {
		return nil
	}

	retMap := make(map[string]data.StorageManager)

	// give back a map copy
	for key, val := range msc.stateComponents.trieStorageManagers {
		retMap[key] = val
	}

	return retMap
}

// SetTriesContainer sets the internal tries container to the one given as parameter
func (msc *managedStateComponents) SetTriesContainer(triesContainer state.TriesHolder) error {
	if check.IfNil(triesContainer) {
		return errors.ErrNilTriesContainer
	}

	msc.mutStateComponents.Lock()
	msc.stateComponents.triesContainer = triesContainer
	msc.mutStateComponents.Unlock()

	return nil
}

// SetTriesStorageManagers sets the internal map with the given parameter
func (msc *managedStateComponents) SetTriesStorageManagers(managers map[string]data.StorageManager) error {
	if len(managers) == 0 {
		return errors.ErrNilTriesStorageManagers
	}

	msc.mutStateComponents.Lock()
	msc.stateComponents.trieStorageManagers = managers
	msc.mutStateComponents.Unlock()

	return nil
}

// IsInterfaceNil returns true if the interface is nil
func (msc *managedStateComponents) IsInterfaceNil() bool {
	return msc == nil
}

// String returns the name of the component
func (msc *managedStateComponents) String() string {
	return "managedStateComponents"
}
