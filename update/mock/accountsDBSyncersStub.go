package mock

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/update"
)

// AccountsDBSyncersStub -
type AccountsDBSyncersStub struct {
	GetCalled          func(key string) (update.AccountsDBSyncer, error)
	AddCalled          func(key string, val update.AccountsDBSyncer) error
	AddMultipleCalled  func(keys []string, interceptors []update.AccountsDBSyncer) error
	ReplaceCalled      func(key string, val update.AccountsDBSyncer) error
	RemoveCalled       func(key string)
	LenCalled          func() int
	StartSyncingCalled func(rootHash []byte) error
	TrieCalled         func() common.Trie
}

// Get -
func (a *AccountsDBSyncersStub) Get(key string) (update.AccountsDBSyncer, error) {
	if a.GetCalled != nil {
		return a.GetCalled(key)
	}

	return nil, nil
}

// Add -
func (a *AccountsDBSyncersStub) Add(key string, val update.AccountsDBSyncer) error {
	if a.AddCalled != nil {
		return a.AddCalled(key, val)
	}

	return nil
}

// AddMultiple -
func (a *AccountsDBSyncersStub) AddMultiple(keys []string, interceptors []update.AccountsDBSyncer) error {
	if a.AddMultipleCalled != nil {
		return a.AddMultipleCalled(keys, interceptors)
	}

	return nil
}

// Replace -
func (a *AccountsDBSyncersStub) Replace(key string, val update.AccountsDBSyncer) error {
	if a.ReplaceCalled != nil {
		return a.ReplaceCalled(key, val)
	}

	return nil
}

// Remove -
func (a *AccountsDBSyncersStub) Remove(key string) {
	if a.RemoveCalled != nil {
		a.RemoveCalled(key)
	}
}

// Len -
func (a *AccountsDBSyncersStub) Len() int {
	if a.LenCalled != nil {
		return a.LenCalled()
	}
	return 0
}

// IsInterfaceNil -
func (a *AccountsDBSyncersStub) IsInterfaceNil() bool {
	return a == nil
}
