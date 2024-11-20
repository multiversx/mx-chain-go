package trie

import (
	"sync"
	
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type rootManager struct {
	root         node
	oldHashes    [][]byte
	oldRootHash  []byte
	mutOperation sync.RWMutex
}

func NewRootManager() *rootManager {
	return &rootManager{
		root:         nil,
		oldHashes:    make([][]byte, 0),
		oldRootHash:  make([]byte, 0),
		mutOperation: sync.RWMutex{},
	}
}

func (rm *rootManager) GetRootNode() node {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	return rm.root
}

func (rm *rootManager) SetNewRootNode(newRoot node) {
	rm.mutOperation.Lock()
	defer rm.mutOperation.Unlock()

	rm.root = newRoot
}

func (rm *rootManager) SetDataForRootChange(newRoot node, oldRootHash []byte, oldHashes [][]byte) {
	rm.mutOperation.Lock()
	defer rm.mutOperation.Unlock()

	rm.root = newRoot
	if len(oldRootHash) > 0 {
		rm.oldRootHash = oldRootHash
	}
	rm.oldHashes = append(rm.oldHashes, oldHashes...)
}

func (rm *rootManager) ResetCollectedHashes() {
	rm.mutOperation.Lock()
	defer rm.mutOperation.Unlock()

	rm.oldRootHash = make([]byte, 0)
	rm.oldHashes = make([][]byte, 0)
}

func (rm *rootManager) GetOldHashes() [][]byte {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	return rm.oldHashes
}

func (rm *rootManager) GetOldRootHash() []byte {
	rm.mutOperation.RLock()
	defer rm.mutOperation.RUnlock()

	if !check.IfNil(rm.root) && !rm.root.isDirty() {
		return rm.root.getHash()
	}

	return rm.oldRootHash
}
