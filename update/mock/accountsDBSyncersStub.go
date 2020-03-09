package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/update"
)

type AccountsDBSyncersStub struct {
	GetCalled          func(key string) (update.AccountsDBSyncer, error)
	AddCalled          func(key string, val update.AccountsDBSyncer) error
	AddMultipleCalled  func(keys []string, interceptors []update.AccountsDBSyncer) error
	ReplaceCalled      func(key string, val update.AccountsDBSyncer) error
	RemoveCalled       func(key string)
	LenCalled          func() int
	StartSyncingCalled func(rootHash []byte) error
	TrieCalled         func() data.Trie
}

func (a *AccountsDBSyncersStub) Get(key string) (update.AccountsDBSyncer, error) {
	if a.GetCalled != nil {
		return a.GetCalled(key)
	}

	return nil, nil
}
func (a *AccountsDBSyncersStub) Add(key string, val update.AccountsDBSyncer) error {
	if a.AddCalled != nil {
		return a.AddCalled(key, val)
	}

	return nil
}
func (a *AccountsDBSyncersStub) AddMultiple(keys []string, interceptors []update.AccountsDBSyncer) error {
	if a.AddMultipleCalled != nil {
		return a.AddMultipleCalled(keys, interceptors)
	}

	return nil
}
func (a *AccountsDBSyncersStub) Replace(key string, val update.AccountsDBSyncer) error {
	if a.ReplaceCalled != nil {
		return a.ReplaceCalled(key, val)
	}

	return nil
}
func (a *AccountsDBSyncersStub) Remove(key string) {
	if a.RemoveCalled != nil {
		a.RemoveCalled(key)
	}

}
func (a *AccountsDBSyncersStub) Len() int {
	if a.LenCalled != nil {
		return a.LenCalled()
	}
	return 0
}
func (a *AccountsDBSyncersStub) IsInterfaceNil() bool {
	return a == nil
}
