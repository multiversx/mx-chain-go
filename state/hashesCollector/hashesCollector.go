package hashesCollector

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type hashesCollector struct {
	common.TrieHashesCollector

	oldRootHash []byte
}

// ErrNilTrieHashesCollector is returned when the trie hashes collector is nil.
var ErrNilTrieHashesCollector = errors.New("nil trie hashes collector")

// NewHashesCollector creates a new instance of hashesCollector.
// This collector is used to collect hashes related to the main trie.
func NewHashesCollector(collector common.TrieHashesCollector) (*hashesCollector, error) {
	if check.IfNil(collector) {
		return nil, ErrNilTrieHashesCollector
	}
	return &hashesCollector{
		TrieHashesCollector: collector,
		oldRootHash:         nil,
	}, nil
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

// Clean initializes the old root hash and the old and new hashes collectors.
func (hc *hashesCollector) Clean() {
	hc.TrieHashesCollector.Clean()
	hc.oldRootHash = nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hc *hashesCollector) IsInterfaceNil() bool {
	return hc == nil
}
