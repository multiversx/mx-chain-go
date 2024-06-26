package trie

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

// trieStorageManagerWithoutPruning manages the storage operations of the trie, but does not prune old values
type trieStorageManagerWithoutPruning struct {
	common.StorageManager
}

// NewTrieStorageManagerWithoutPruning creates a new instance of trieStorageManagerWithoutPruning
func NewTrieStorageManagerWithoutPruning(sm common.StorageManager) (*trieStorageManagerWithoutPruning, error) {
	if check.IfNil(sm) {
		return nil, ErrNilTrieStorage
	}

	return &trieStorageManagerWithoutPruning{
		StorageManager: sm,
	}, nil
}

// IsPruningEnabled returns false if the trie pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) IsPruningEnabled() bool {
	return false
}

// Remove deletes the given hash from checkpointHashesHolder
func (tsm *trieStorageManagerWithoutPruning) Remove(_ []byte) error {
	return nil
}
