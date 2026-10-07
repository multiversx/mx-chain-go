package mock

import (
	"github.com/multiversx/mx-chain-go/common"
)

// AccountsDBSyncerStub -
type AccountsDBSyncerStub struct {
	GetSyncedTriesCalled func() map[string]common.Trie
	SyncAccountsCalled   func(rootHash []byte, storageMarker common.StorageMarker) error
}

// GetSyncedTries -
func (a *AccountsDBSyncerStub) GetSyncedTries() map[string]common.Trie {
	if a.GetSyncedTriesCalled != nil {
		return a.GetSyncedTriesCalled()
	}
	return nil
}

// SyncAccounts -
func (a *AccountsDBSyncerStub) SyncAccounts(rootHash []byte, storageMarker common.StorageMarker) error {
	if a.SyncAccountsCalled != nil {
		return a.SyncAccountsCalled(rootHash, storageMarker)
	}
	return nil
}

// IsInterfaceNil -
func (a *AccountsDBSyncerStub) IsInterfaceNil() bool {
	return a == nil
}
