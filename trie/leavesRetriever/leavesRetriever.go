package leavesRetriever

import (
	"context"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever/dfsTrieIterator"
)

type leavesRetriever struct {
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
	maxSize    uint64
}

// NewLeavesRetriever creates a new leaves retriever
func NewLeavesRetriever(marshaller marshal.Marshalizer, hasher hashing.Hasher, maxSize uint64) (*leavesRetriever, error) {
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshaller
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &leavesRetriever{
		marshaller: marshaller,
		hasher:     hasher,
		maxSize:    maxSize,
	}, nil
}

// IterateLeaves walks the trie found in the given storage, starting from the iterator state, and calls visit for
// every leaf. It returns the iterator state from which the iteration can be resumed, or nil when it finished.
// The size limit of the walk is maxSize, or the retriever's one when maxSize is zero or larger.
// New pruning requests of the given storage are buffered while the leaves are iterated; between two calls the
// nodes referenced by the iterator state can be pruned.
func (lr *leavesRetriever) IterateLeaves(numLeaves int, maxSize uint64, iteratorState [][]byte, storage common.StorageManager, leavesParser common.TrieLeafParser, ctx context.Context, visit common.LeafVisitor) ([][]byte, error) {
	if check.IfNil(leavesParser) {
		return nil, fmt.Errorf("nil leaves parser")
	}
	if maxSize == 0 || maxSize > lr.maxSize {
		maxSize = lr.maxSize
	}

	iterator, err := dfsTrieIterator.NewIterator(iteratorState, storage, lr.marshaller, lr.hasher)
	if err != nil {
		return nil, err
	}

	storage.EnterPruningBufferingMode()
	defer storage.ExitPruningBufferingMode()

	err = iterator.IterateLeaves(numLeaves, maxSize, leavesParser, ctx, visit)
	if err != nil {
		return nil, err
	}

	return iterator.GetIteratorState(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (lr *leavesRetriever) IsInterfaceNil() bool {
	return lr == nil
}
