package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
)

// StateComponentsMock -
type StateComponentsMock struct {
	PeersAcc        state.AccountsAdapter
	Accounts        state.AccountsAdapter
	AccountsAPI     state.AccountsAdapter
	Tries           common.TriesHolder
	StorageManagers map[string]common.StorageManager
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

// AccountsAdapterAPI -
func (scm *StateComponentsMock) AccountsAdapterAPI() state.AccountsAdapter {
	return scm.AccountsAPI
}

// TriesContainer -
func (scm *StateComponentsMock) TriesContainer() common.TriesHolder {
	return scm.Tries
}

// TrieStorageManagers -
func (scm *StateComponentsMock) TrieStorageManagers() map[string]common.StorageManager {
	return scm.StorageManagers
}

// String -
func (scm *StateComponentsMock) String() string {
	return "StateComponentsMock"
}

// IsInterfaceNil -
func (scm *StateComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
