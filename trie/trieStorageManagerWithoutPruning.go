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
	storage *trieStorageManager
}

// NewTrieStorageManagerWithoutPruning creates a new instance of trieStorageManagerWithoutPruning
func NewTrieStorageManagerWithoutPruning(sm common.StorageManager) (*trieStorageManagerWithoutPruning, error) {
	if check.IfNil(sm) {
		return nil, ErrNilTrieStorage
	}

	tsm, ok := sm.GetBaseTrieStorageManager().(*trieStorageManager)
	if !ok {
		return nil, errors.New("invalid storage manager type" + fmt.Sprintf("%T", sm.GetBaseTrieStorageManager()))
	}

	return &trieStorageManagerWithoutPruning{
		StorageManager: sm,
		storage:        tsm,
	}, nil
}

// IsPruningEnabled returns false if the trie pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) IsPruningEnabled() bool {
	return false
}

// Remove deletes the given hash from checkpointHashesHolder
func (tsm *trieStorageManagerWithoutPruning) Remove(hash []byte) error {
	tsm.storage.RemoveFromCheckpointHashesHolder(hash)
	return nil
}
