package trie

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
)

// trieStorageManagerWithoutPruning manages the storage operations of the trie, but does not prune old values
type trieStorageManagerWithoutPruning struct {
	common.StorageManager
	storageManagerExtension
}

// NewTrieStorageManagerWithoutPruning creates a new instance of trieStorageManagerWithoutPruning
func NewTrieStorageManagerWithoutPruning(tsm common.StorageManager) (*trieStorageManagerWithoutPruning, error) {
	if check.IfNil(tsm) {
		return nil, ErrNilTrieStorage
	}

	sm, ok := tsm.(storageManagerExtension)
	if !ok {
		return nil, errors.New("invalid storage manager type" + fmt.Sprintf("%T", tsm))
	}

	return &trieStorageManagerWithoutPruning{
		StorageManager:          tsm,
		storageManagerExtension: sm,
	}, nil
}

// IsPruningEnabled returns false if the trie pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) IsPruningEnabled() bool {
	return false
}

// Remove deletes the given hash from checkpointHashesHolder
func (tsm *trieStorageManagerWithoutPruning) Remove(hash []byte) error {
	tsm.storageManagerExtension.removeFromCheckpointHashesHolder(hash)
	return nil
}
