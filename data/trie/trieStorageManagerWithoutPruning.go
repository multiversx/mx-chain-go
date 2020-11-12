package trie

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
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
func (tsm *trieStorageManagerWithoutPruning) TakeSnapshot(_ []byte) {
	log.Trace("trieStorageManagerWithoutPruning - TakeSnapshot:trie storage pruning is disabled")
}

// SetCheckpoint does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) SetCheckpoint(_ []byte) {
	log.Trace("trieStorageManagerWithoutPruning - SetCheckpoint:trie storage pruning is disabled")
}

// Prune does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) Prune(_ []byte, _ data.TriePruningIdentifier) {
	log.Trace("trieStorageManagerWithoutPruning - Prune:trie storage pruning is disabled")
}

// CancelPrune does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) CancelPrune(_ []byte, _ data.TriePruningIdentifier) {
	log.Trace("trieStorageManagerWithoutPruning - CancelPrune:trie storage pruning is disabled")
}

// MarkForEviction does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) MarkForEviction(_ []byte, _ data.ModifiedHashes) error {
	log.Trace("trieStorageManagerWithoutPruning - MarkForEviction:trie storage pruning is disabled")
	return nil
}

// Close - closes all underlying components
func (tsm *trieStorageManagerWithoutPruning) Close() error {
	log.Trace("trieStorageManagerWithoutPruning - Close:trie storage pruning is disabled")
	return nil
}

// IsPruningEnabled returns false if the trie pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) IsPruningEnabled() bool {
	return false
}
