package factory

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

var _ ComponentHandler = (*managedStateComponents)(nil)
var _ StateComponentsHolder = (*managedStateComponents)(nil)
var _ StateComponentsHandler = (*managedStateComponents)(nil)

// TODO: integrate this in main.go and remove obsolete component from structs.go afterwards

type managedStateComponents struct {
	*stateComponents
	factory            *stateComponentsFactory
	mutStateComponents sync.RWMutex
}

// NewManagedStateComponents returns a news instance of managedStateComponents
func NewManagedStateComponents(args StateComponentsFactoryArgs) (*managedStateComponents, error) {
	pcf, err := NewStateComponentsFactory(args)
	if err != nil {
		return nil, err
	}

	return &managedStateComponents{
		stateComponents: nil,
		factory:         pcf,
	}, nil
}

// Create will create the managed components
func (m *managedStateComponents) Create() error {
	sc, err := m.factory.Create()
	if err != nil {
		return err
	}

	m.mutStateComponents.Lock()
	m.stateComponents = sc
	m.mutStateComponents.Unlock()

	return nil
}

// Close will close all underlying sub-components
func (m *managedStateComponents) Close() error {
	m.mutStateComponents.Lock()
	defer m.mutStateComponents.Unlock()

	if m.stateComponents != nil {
		err := m.stateComponents.Close()
		if err != nil {
			return err
		}
		m.stateComponents = nil
	}

	return nil
}

// PeerAccounts returns the accounts adapter for the validators
func (m *managedStateComponents) PeerAccounts() state.AccountsAdapter {
	return m.stateComponents.peerAccounts
}

// AccountsAdapter returns the accounts adapter for the user accounts
func (m *managedStateComponents) AccountsAdapter() state.AccountsAdapter {
	return m.stateComponents.accountsAdapter
}

// TriesContainer returns the tries container
func (m *managedStateComponents) TriesContainer() state.TriesHolder {
	return m.stateComponents.triesContainer
}

// TrieStorageManagers returns the trie storage manager for the given account type
func (m *managedStateComponents) TrieStorageManagers() map[string]data.StorageManager {
	retMap := make(map[string]data.StorageManager)

	// give back a map copy
	m.mutStateComponents.RLock()
	for key, val := range m.stateComponents.trieStorageManagers {
		retMap[key] = val
	}
	m.mutStateComponents.RUnlock()

	return retMap
}

// SetTriesContainer sets the internal tries container to the one given as parameter
func (m *managedStateComponents) SetTriesContainer(triesContainer state.TriesHolder) error {
	if check.IfNil(triesContainer) {
		return ErrNilTriesContainer
	}

	m.mutStateComponents.Lock()
	m.stateComponents.triesContainer = triesContainer
	m.mutStateComponents.Unlock()

	return nil
}

// SetTriesStorageManagers sets the internal map with the given parameter
func (m *managedStateComponents) SetTriesStorageManagers(managers map[string]data.StorageManager) error {
	if len(managers) == 0 {
		return ErrNilTriesStorageManagers
	}

	m.mutStateComponents.Lock()
	m.stateComponents.trieStorageManagers = managers
	m.mutStateComponents.Unlock()

	return nil
}

// IsInterfaceNil returns true if the interface is nil
func (m *managedStateComponents) IsInterfaceNil() bool {
	return m == nil
}
