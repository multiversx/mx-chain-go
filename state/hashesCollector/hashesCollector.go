package hashesCollector

import (
	"github.com/multiversx/mx-chain-go/common"
)

type hashesCollector struct {
	common.TrieHashesCollector

	oldRootHash []byte
}

// NewHashesCollector creates a new instance of hashesCollector.
// This collector is used to collect hashes related to the main trie.
func NewHashesCollector(collector common.TrieHashesCollector) *hashesCollector {
	return &hashesCollector{
		TrieHashesCollector: collector,
		oldRootHash:         nil,
	}
}

// AddObsoleteHashes adds the old root hash and the old hashes to the collector.
func (hc *hashesCollector) AddObsoleteHashes(oldRootHash []byte, oldHashes [][]byte) {
	hc.TrieHashesCollector.AddObsoleteHashes(oldRootHash, oldHashes)
	hc.oldRootHash = oldRootHash
}

// GetCollectedData returns the old root hash, the old hashes and the new hashes.
func (hc *hashesCollector) GetCollectedData() ([]byte, common.ModifiedHashes, common.ModifiedHashes) {
	_, oldHashes, newHashes := hc.TrieHashesCollector.GetCollectedData()
	return hc.oldRootHash, oldHashes, newHashes
}
