package mock

import (
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// AccountsDBSyncerStub -
type AccountsDBSyncerStub struct {
	GetSyncedTriesCalled func() map[string]temporary.Trie
	SyncAccountsCalled   func(rootHash []byte, shardId uint32) error
}

// GetSyncedTries -
func (a *AccountsDBSyncerStub) GetSyncedTries() map[string]temporary.Trie {
	if a.GetSyncedTriesCalled != nil {
		return a.GetSyncedTriesCalled()
	}
	return nil
}

// SyncAccounts -
func (a *AccountsDBSyncerStub) SyncAccounts(rootHash []byte, shardId uint32) error {
	if a.SyncAccountsCalled != nil {
		return a.SyncAccountsCalled(rootHash, shardId)
	}
	return nil
}

// IsInterfaceNil -
func (a *AccountsDBSyncerStub) IsInterfaceNil() bool {
	return a == nil
}
