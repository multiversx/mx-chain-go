package trie

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

// TODO: add unit tests

type rootManager struct {
	root         node
	oldHashes    [][]byte
	oldRootHash  []byte
	mutOperation sync.RWMutex
}

// NewRootManager creates a new instance of rootManager. All operations are protected by a mutex
func NewRootManager() *rootManager {
	return &rootManager{
		root:         nil,
		oldHashes:    make([][]byte, 0),
		oldRootHash:  make([]byte, 0),
		mutOperation: sync.RWMutex{},
	}
}

// GetRootNode returns the root node
func (rm *rootManager) GetRootNode() node {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	return rm.root
}

// SetNewRootNode sets the given node as the new root node
func (rm *rootManager) SetNewRootNode(newRoot node) {
	rm.mutOperation.Lock()
	defer rm.mutOperation.Unlock()

	rm.root = newRoot
}

// SetDataForRootChange sets the new root node, the old root hash and the old hashes
func (rm *rootManager) SetDataForRootChange(newRoot node, oldRootHash []byte, oldHashes [][]byte) {
	rm.mutOperation.Lock()
	defer rm.mutOperation.Unlock()

	rm.root = newRoot
	if len(oldRootHash) > 0 {
		rm.oldRootHash = oldRootHash
	}
	rm.oldHashes = append(rm.oldHashes, oldHashes...)
}

// ResetCollectedHashes resets the old root hash and the old hashes
func (rm *rootManager) ResetCollectedHashes() {
	rm.mutOperation.Lock()
	defer rm.mutOperation.Unlock()

	rm.oldRootHash = make([]byte, 0)
	rm.oldHashes = make([][]byte, 0)
}

// GetOldHashes returns the old hashes
func (rm *rootManager) GetOldHashes() [][]byte {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	oldHashes := make([][]byte, len(rm.oldHashes))
	copy(oldHashes, rm.oldHashes)
	return oldHashes
}

// GetOldRootHash returns the old root hash
func (rm *rootManager) GetOldRootHash() []byte {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	if !check.IfNil(rm.root) && !rm.root.isDirty() {
		return rm.root.getHash()
	}

	return rm.oldRootHash
}
