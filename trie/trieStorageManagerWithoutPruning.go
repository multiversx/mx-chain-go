package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
)

// trieStorageManagerWithoutPruning manages the storage operations of the trie, but does not prune old values
type trieStorageManagerWithoutPruning struct {
	*trieStorageManager
}

// NewTrieStorageManagerWithoutPruning creates a new instance of trieStorageManagerWithoutPruning
func NewTrieStorageManagerWithoutPruning(db common.DBWriteCacher) (*trieStorageManagerWithoutPruning, error) {
	if check.IfNil(db) {
		return nil, ErrNilDatabase
	}

	return &trieStorageManagerWithoutPruning{&trieStorageManager{db: db}}, nil
}

// Put adds only to the db
func (tsm *trieStorageManagerWithoutPruning) Put(key []byte, val []byte) error {
	return tsm.db.Put(key, val)
}

// Get checks only the db
func (tsm *trieStorageManagerWithoutPruning) Get(key []byte) ([]byte, error) {
	return tsm.db.Get(key)
}

// TakeSnapshot does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) TakeSnapshot(_ []byte, _ bool, chLeaves chan core.KeyValueHolder) {
	if chLeaves != nil {
		close(chLeaves)
	}

	log.Trace("trieStorageManagerWithoutPruning - TakeSnapshot:trie storage pruning is disabled")
}

// SetCheckpoint does nothing if pruning is disabled
func (tsm *trieStorageManagerWithoutPruning) SetCheckpoint(_ []byte, chLeaves chan core.KeyValueHolder) {
	if chLeaves != nil {
		close(chLeaves)
	}

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
func (tsm *trieStorageManagerWithoutPruning) AddDirtyCheckpointHashes(_ []byte, _ common.ModifiedHashes) bool {
	return false
}

// Remove does nothing for this implementation
func (tsm *trieStorageManagerWithoutPruning) Remove(_ []byte) error {
	return nil
}
