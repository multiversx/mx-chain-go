package hashesCollector

import (
	"sync"

	"github.com/multiversx/mx-chain-go/common"
)

type dataTrieHashesCollector struct {
	oldHashes common.ModifiedHashes
	newHashes common.ModifiedHashes

	sync.RWMutex
}

// NewDataTrieHashesCollector creates a new instance of dataTrieHashesCollector.
// This collector is used to collect hashes related to the data trie.
func NewDataTrieHashesCollector() *dataTrieHashesCollector {
	return &dataTrieHashesCollector{
		oldHashes: make(common.ModifiedHashes),
		newHashes: make(common.ModifiedHashes),
	}
}

// AddDirtyHash adds the new hash to the collector.
func (hc *dataTrieHashesCollector) AddDirtyHash(hash []byte) {
	hc.Lock()
	defer hc.Unlock()

	hc.newHashes[string(hash)] = struct{}{}
}

// GetDirtyHashes returns the new hashes.
func (hc *dataTrieHashesCollector) GetDirtyHashes() common.ModifiedHashes {
	hc.RLock()
	defer hc.RUnlock()

	return hc.newHashes
}

// AddObsoleteHashes adds the old hashes to the collector.
func (hc *dataTrieHashesCollector) AddObsoleteHashes(_ []byte, oldHashes [][]byte) {
	hc.Lock()
	defer hc.Unlock()

	for _, hash := range oldHashes {
		hc.oldHashes[string(hash)] = struct{}{}
	}
}

// GetCollectedData returns the old hashes and the new hashes.
func (hc *dataTrieHashesCollector) GetCollectedData() ([]byte, common.ModifiedHashes, common.ModifiedHashes) {
	hc.RLock()
	defer hc.RUnlock()

	return nil, hc.oldHashes, hc.newHashes
}
