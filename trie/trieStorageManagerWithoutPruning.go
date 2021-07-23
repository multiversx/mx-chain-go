package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// trieStorageManagerWithoutPruning manages the storage operations of the trie, but does not prune old values
type trieStorageManagerWithoutPruning struct {
	*trieStorageManager
}

// NewTrieStorageManagerWithoutPruning creates a new instance of trieStorageManagerWithoutPruning
func NewTrieStorageManagerWithoutPruning(db temporary.DBWriteCacher) (*trieStorageManagerWithoutPruning, error) {
	if check.IfNil(db) {
		return nil, ErrNilDatabase
	}

	return &trieStorageManagerWithoutPruning{&trieStorageManager{db: db}}, nil
}

// TakeSnapshot does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) TakeSnapshot(_ []byte, _ bool, _ chan core.KeyValueHolder) {
	log.Trace("trieStorageManagerWithoutPruning - TakeSnapshot:trie storage pruning is disabled")
}

// SetCheckpoint does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) SetCheckpoint(_ []byte, _ chan core.KeyValueHolder) {
	log.Trace("trieStorageManagerWithoutPruning - SetCheckpoint:trie storage pruning is disabled")
}

// Close - closes all underlying components
func (tsm *trieStorageManagerWithoutPruning) Close() error {
	log.Trace("trieStorageManagerWithoutPruning - Close:trie storage pruning is disabled")
	return tsm.db.Close()
}

// IsPruningEnabled returns false if the trie pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) IsPruningEnabled() bool {
	return false
}

// AddDirtyCheckpointHashes does nothing for this implementation
func (tsm *trieStorageManagerWithoutPruning) AddDirtyCheckpointHashes(_ []byte, _ temporary.ModifiedHashes) bool {
	return false
}

// Remove does nothing for this implementation
func (tsm *trieStorageManagerWithoutPruning) Remove(_ []byte) error {
	return nil
}
