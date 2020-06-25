package factory

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

var _ ComponentHandler = (*managedStateComponents)(nil)
var _ StateComponentsHolder = (*managedStateComponents)(nil)
var _ StateComponentsHandler = (*managedStateComponents)(nil)

// TODO: integrate this in main.go and remove obsolete component from structs.go afterwards

type managedStateComponents struct {
	*stateComponents
	factory              *stateComponentsFactory
	cancelFunc           func()
	mutProcessComponents sync.RWMutex
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

	m.mutProcessComponents.Lock()
	m.stateComponents = sc
	_, m.cancelFunc = context.WithCancel(context.Background())
	m.mutProcessComponents.Unlock()

	return nil
}

// Close will close all underlying sub-components
func (m *managedStateComponents) Close() error {
	m.mutProcessComponents.Lock()
	defer m.mutProcessComponents.Unlock()

	m.cancelFunc()
	//TODO: close underlying components
	m.cancelFunc = nil
	m.stateComponents = nil

	return nil
}

// PeerAccounts returns the accounts adapter for the validators
func (m *managedStateComponents) PeerAccounts() state.AccountsAdapter {
	return m.stateComponents.PeerAccounts
}

// AccountsAdapter returns the accounts adapter for the user accounts
func (m *managedStateComponents) AccountsAdapter() state.AccountsAdapter {
	return m.stateComponents.AccountsAdapter
}

// TriesContainer returns the tries container
func (m *managedStateComponents) TriesContainer() state.TriesHolder {
	return m.stateComponents.triesContainer
}

// TrieStorageManager returns the trie storage manager for the given account type
func (m *managedStateComponents) TrieStorageManager(accType string) data.StorageManager {
	m.mutTrieStorageManagers.RLock()
	defer m.mutTrieStorageManagers.RUnlock()

	return m.stateComponents.trieStorageManagers[accType]
}

// IsInterfaceNil returns true if the interface is nil
func (m *managedStateComponents) IsInterfaceNil() bool {
	return m == nil
}
