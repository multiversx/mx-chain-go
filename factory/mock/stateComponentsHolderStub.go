package mock

import (
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// StateComponentsHolderStub -
type StateComponentsHolderStub struct {
	PeerAccountsCalled        func() state.AccountsAdapter
	AccountsAdapterCalled     func() state.AccountsAdapter
	AccountsAdapterAPICalled  func() state.AccountsAdapter
	TriesContainerCalled      func() state.TriesHolder
	TrieStorageManagersCalled func() map[string]temporary.StorageManager
}

// PeerAccounts -
func (s *StateComponentsHolderStub) PeerAccounts() state.AccountsAdapter {
	if s.PeerAccountsCalled != nil {
		return s.PeerAccountsCalled()
	}

	return nil
}

// AccountsAdapter -
func (s *StateComponentsHolderStub) AccountsAdapter() state.AccountsAdapter {
	if s.AccountsAdapterCalled != nil {
		return s.AccountsAdapterCalled()
	}

	return nil
}

// AccountsAdapterAPI -
func (s *StateComponentsHolderStub) AccountsAdapterAPI() state.AccountsAdapter {
	if s.AccountsAdapterAPICalled != nil {
		return s.AccountsAdapterAPICalled()
	}

	return nil
}

// TriesContainer -
func (s *StateComponentsHolderStub) TriesContainer() state.TriesHolder {
	if s.TriesContainerCalled != nil {
		return s.TriesContainerCalled()
	}

	return nil
}

// TrieStorageManagers -
func (s *StateComponentsHolderStub) TrieStorageManagers() map[string]temporary.StorageManager {
	if s.TrieStorageManagersCalled != nil {
		return s.TrieStorageManagersCalled()
	}

	return nil
}

// IsInterfaceNil -
func (s *StateComponentsHolderStub) IsInterfaceNil() bool {
	return s == nil
}
