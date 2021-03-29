package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// StateComponentsMock -
type StateComponentsMock struct {
	PeersAcc        state.AccountsAdapter
	Accounts        state.AccountsAdapter
	Tries           state.TriesHolder
	StorageManagers map[string]data.StorageManager
}

// Create -
func (scm *StateComponentsMock) Create() error {
	return nil
}

// Close -
func (scm *StateComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (scm *StateComponentsMock) CheckSubcomponents() error {
	return nil
}

// PeerAccounts -
func (scm *StateComponentsMock) PeerAccounts() state.AccountsAdapter {
	return scm.PeersAcc
}

// AccountsAdapter -
func (scm *StateComponentsMock) AccountsAdapter() state.AccountsAdapter {
	return scm.Accounts
}

// TriesContainer -
func (scm *StateComponentsMock) TriesContainer() state.TriesHolder {
	return scm.Tries
}

// TrieStorageManagers -
func (scm *StateComponentsMock) TrieStorageManagers() map[string]data.StorageManager {
	return scm.StorageManagers
}

// IsInterfaceNil -
func (scm *StateComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
