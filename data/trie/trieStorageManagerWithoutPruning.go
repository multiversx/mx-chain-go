package trie

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// trieStorageManagerWithoutPruning manages the storage operations of the trie, but does not prune old values
type trieStorageManagerWithoutPruning struct {
	*trieStorageManager
}

// NewTrieStorageManagerWithoutPruning creates a new instance of trieStorageManagerWithoutPruning
func NewTrieStorageManagerWithoutPruning(db data.DBWriteCacher) (*trieStorageManagerWithoutPruning, error) {
	if check.IfNil(db) {
		return nil, ErrNilDatabase
	}

	return &trieStorageManagerWithoutPruning{&trieStorageManager{db: db}}, nil
}

// TakeSnapshot does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) TakeSnapshot([]byte, marshal.Marshalizer, hashing.Hasher) {
	log.Trace("trieStorageManagerWithoutPruning - TakeSnapshot:trie storage pruning is disabled")
}

// SetCheckpoint does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) SetCheckpoint([]byte, marshal.Marshalizer, hashing.Hasher) {
	log.Trace("trieStorageManagerWithoutPruning - SetCheckpoint:trie storage pruning is disabled")
}

// Prune does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) Prune([]byte) error {
	log.Trace("trieStorageManagerWithoutPruning - Prune:trie storage pruning is disabled")
	return nil
}

// CancelPrune does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) CancelPrune([]byte) {
	log.Trace("trieStorageManagerWithoutPruning - CancelPrune:trie storage pruning is disabled")
}

// MarkForEviction does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) MarkForEviction([]byte, [][]byte) error {
	log.Trace("trieStorageManagerWithoutPruning - MarkForEviction:trie storage pruning is disabled")
	return nil
}

// Clone returns a new instance of trieStorageManagerWithoutPruning
func (tsm *trieStorageManagerWithoutPruning) Clone() data.StorageManager {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	return &trieStorageManagerWithoutPruning{
		&trieStorageManager{
			db: tsm.db,
		},
	}
}

// IsPruningEnabled returns false if the trie pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) IsPruningEnabled() bool {
	return false
}
