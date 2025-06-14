package trie

import (
	"sync"
)

type rootManager struct {
	root         node
	rootHash     []byte
	oldHashes    [][]byte
	oldRootHash  []byte
	mutOperation sync.RWMutex
}

// RootData holds information about the current and previous root nodes and their respective hashes in the trie.
type RootData struct {
	root        node
	rootHash    []byte
	oldRootHash []byte
	oldHashes   [][]byte
}

// NewRootManager creates a new instance of rootManager. All operations are protected by a mutex
func NewRootManager() *rootManager {
	return &rootManager{
		root:         nil,
		rootHash:     nil,
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

// GetRootHash returns the root hash
func (rm *rootManager) GetRootHash() []byte {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	return rm.rootHash
}

// SetDataForRootChange sets the new root node, the old root hash and the old hashes
func (rm *rootManager) SetDataForRootChange(rootData RootData) {
	rm.mutOperation.Lock()
	defer rm.mutOperation.Unlock()

	rm.root = rootData.root
	rm.rootHash = rootData.rootHash
	rm.oldRootHash = rootData.oldRootHash
	rm.oldHashes = append(rm.oldHashes, rootData.oldHashes...)
}

// GetRootData returns the collected root data.
func (rm *rootManager) GetRootData() RootData {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	return RootData{
		root:        rm.root,
		rootHash:    rm.rootHash,
		oldRootHash: rm.oldRootHash,
		oldHashes:   rm.oldHashes,
	}
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

	return rm.oldHashes
}

// GetOldRootHash returns the old root hash
func (rm *rootManager) GetOldRootHash() []byte {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	return rm.oldRootHash
}
