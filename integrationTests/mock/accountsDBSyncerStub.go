package mock

import (
	"context"
	"github.com/multiversx/mx-chain-go/common"
)

// AccountsDBSyncerStub -
type AccountsDBSyncerStub struct {
	SyncAccountsWithContextCalled func(context.Context, []byte, common.StorageMarker) error
	GetSyncedTriesCalled          func() map[string]common.Trie
	SyncAccountsCalled            func(rootHash []byte, storageMarker common.StorageMarker) error
}

// SyncAccountsWithContext -
func (a *AccountsDBSyncerStub) SyncAccountsWithContext(ctx context.Context, rootHash []byte, marker common.StorageMarker) error {
	if a.SyncAccountsWithContextCalled != nil {
		return a.SyncAccountsWithContextCalled(ctx, rootHash, marker)
	}
	return a.SyncAccounts(rootHash, marker)
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
