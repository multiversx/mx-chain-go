package mock

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// StateComponentsHolderStub -
type StateComponentsHolderStub struct {
	PeerAccountsCalled             func() state.AccountsAdapter
	AccountsAdapterCalled          func() state.AccountsAdapter
	AccountsAdapterAPICalled       func() state.AccountsAdapter
	AccountsRepositoryCalled       func() state.AccountsRepository
	TriesContainerCalled           func() common.TriesHolder
	TrieStorageManagersCalled      func() map[string]common.StorageManager
	MissingTrieNodesNotifierCalled func() common.MissingTrieNodesNotifier
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

// AccountsRepository -
func (s *StateComponentsHolderStub) AccountsRepository() state.AccountsRepository {
	if s.AccountsRepositoryCalled != nil {
		return s.AccountsRepositoryCalled()
	}

	return nil
}

// TriesContainer -
func (s *StateComponentsHolderStub) TriesContainer() common.TriesHolder {
	if s.TriesContainerCalled != nil {
		return s.TriesContainerCalled()
	}

	return nil
}

// TrieStorageManagers -
func (s *StateComponentsHolderStub) TrieStorageManagers() map[string]common.StorageManager {
	if s.TrieStorageManagersCalled != nil {
		return s.TrieStorageManagersCalled()
	}

	return nil
}

// MissingTrieNodesNotifier -
func (s *StateComponentsHolderStub) MissingTrieNodesNotifier() common.MissingTrieNodesNotifier {
	if s.MissingTrieNodesNotifierCalled != nil {
		return s.MissingTrieNodesNotifierCalled()
	}

	return nil
}

// Close -
func (s *StateComponentsHolderStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (s *StateComponentsHolderStub) IsInterfaceNil() bool {
	return s == nil
}
