package mock

import (
	"context"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/update"
)

// TrieSyncersStub -
type TrieSyncersStub struct {
	GetCalled          func(key string) (update.TrieSyncer, error)
	AddCalled          func(key string, val update.TrieSyncer) error
	AddMultipleCalled  func(keys []string, interceptors []update.TrieSyncer) error
	ReplaceCalled      func(key string, val update.TrieSyncer) error
	RemoveCalled       func(key string)
	LenCalled          func() int
	StartSyncingCalled func(rootHash []byte, ctx context.Context) error
	TrieCalled         func() common.Trie
}

// StartSyncing -
func (tss *TrieSyncersStub) StartSyncing(rootHash []byte, ctx context.Context) error {
	if tss.StartSyncingCalled != nil {
		return tss.StartSyncingCalled(rootHash, ctx)
	}
	return nil
}

// Trie -
func (tss *TrieSyncersStub) Trie() common.Trie {
	if tss.TrieCalled != nil {
		return tss.TrieCalled()
	}
	return nil
}

// Get -
func (tss *TrieSyncersStub) Get(key string) (update.TrieSyncer, error) {
	if tss.GetCalled != nil {
		return tss.GetCalled(key)
	}

	return nil, nil
}

// Add -
func (tss *TrieSyncersStub) Add(key string, val update.TrieSyncer) error {
	if tss.AddCalled != nil {
		return tss.AddCalled(key, val)
	}

	return nil
}

// AddMultiple -
func (tss *TrieSyncersStub) AddMultiple(keys []string, interceptors []update.TrieSyncer) error {
	if tss.AddMultipleCalled != nil {
		return tss.AddMultipleCalled(keys, interceptors)
	}

	return nil
}

// Replace -
func (tss *TrieSyncersStub) Replace(key string, val update.TrieSyncer) error {
	if tss.ReplaceCalled != nil {
		return tss.ReplaceCalled(key, val)
	}

	return nil
}

// Remove -
func (tss *TrieSyncersStub) Remove(key string) {
	if tss.RemoveCalled != nil {
		tss.RemoveCalled(key)
	}

}

// Len -
func (tss *TrieSyncersStub) Len() int {
	if tss.LenCalled != nil {
		return tss.LenCalled()
	}
	return 0
}

// IsInterfaceNil -
func (tss *TrieSyncersStub) IsInterfaceNil() bool {
	return tss == nil
}
