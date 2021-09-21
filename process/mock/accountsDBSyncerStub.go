package mock

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

// AccountsDBSyncerStub -
type AccountsDBSyncerStub struct {
	GetSyncedTriesCalled func() map[string]common.Trie
	SyncAccountsCalled   func(rootHash []byte) error
}

// GetSyncedTries -
func (a *AccountsDBSyncerStub) GetSyncedTries() map[string]common.Trie {
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
