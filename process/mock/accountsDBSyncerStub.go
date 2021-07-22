package mock

import (
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// AccountsDBSyncerStub -
type AccountsDBSyncerStub struct {
	GetSyncedTriesCalled func() map[string]temporary.Trie
	SyncAccountsCalled   func(rootHash []byte) error
}

// GetSyncedTries -
func (a *AccountsDBSyncerStub) GetSyncedTries() map[string]temporary.Trie {
	if a.GetSyncedTriesCalled != nil {
		return a.GetSyncedTriesCalled()
	}
	return nil
}

// SyncAccounts -
func (a *AccountsDBSyncerStub) SyncAccounts(rootHash []byte) error {
	if a.SyncAccountsCalled != nil {
		return a.SyncAccountsCalled(rootHash)
	}
	return nil
}

// IsInterfaceNil -
func (a *AccountsDBSyncerStub) IsInterfaceNil() bool {
	return a == nil
}
