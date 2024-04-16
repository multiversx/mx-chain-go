package leavesRetriever

import "github.com/multiversx/mx-chain-go/common"

// GetIterators -
func (lr *leavesRetriever) GetIterators() map[string]common.DfsIterator {
	return lr.iterators
}

// GetLruIteratorIDs -
func (lr *leavesRetriever) GetLruIteratorIDs() [][]byte {
	return lr.lruIteratorIDs
}

func (lr *leavesRetriever) Size() uint64 {
	return lr.size
}
