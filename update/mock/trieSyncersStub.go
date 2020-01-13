package mock

import (
	"github.com/ElrondNetwork/elrond-go/update"
)

type TrieSyncersStub struct {
	GetCalled         func(key string) (update.TrieSyncer, error)
	AddCalled         func(key string, val update.TrieSyncer) error
	AddMultipleCalled func(keys []string, interceptors []update.TrieSyncer) error
	ReplaceCalled     func(key string, val update.TrieSyncer) error
	RemoveCalled      func(key string)
	LenCalled         func() int
}

func (tss *TrieSyncersStub) Get(key string) (update.TrieSyncer, error) {
	if tss.GetCalled != nil {
		return tss.GetCalled(key)
	}

	return nil, nil
}
func (tss *TrieSyncersStub) Add(key string, val update.TrieSyncer) error {
	if tss.AddCalled != nil {
		return tss.AddCalled(key, val)
	}

	return nil
}
func (tss *TrieSyncersStub) AddMultiple(keys []string, interceptors []update.TrieSyncer) error {
	if tss.AddMultipleCalled != nil {
		return tss.AddMultipleCalled(keys, interceptors)
	}

	return nil
}
func (tss *TrieSyncersStub) Replace(key string, val update.TrieSyncer) error {
	if tss.ReplaceCalled != nil {
		return tss.ReplaceCalled(key, val)
	}

	return nil
}
func (tss *TrieSyncersStub) Remove(key string) {
	if tss.RemoveCalled != nil {
		tss.RemoveCalled(key)
	}

}
func (tss *TrieSyncersStub) Len() int {
	if tss.LenCalled != nil {
		return tss.LenCalled()
	}
	return 0
}
func (tss *TrieSyncersStub) IsInterfaceNil() bool {
	return tss == nil
}
