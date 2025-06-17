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
	db         common.TrieStorageInteractor
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
	maxSize    uint64
}

// NewLeavesRetriever creates a new leaves retriever
func NewLeavesRetriever(db common.TrieStorageInteractor, marshaller marshal.Marshalizer, hasher hashing.Hasher, maxSize uint64) (*leavesRetriever, error) {
	if check.IfNil(db) {
		return nil, ErrNilDB
	}
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshaller
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	return &leavesRetriever{
		db:         db,
		marshaller: marshaller,
		hasher:     hasher,
		maxSize:    maxSize,
	}, nil
}

// GetLeaves retrieves leaves from the trie starting from the iterator state. It will also return the new iterator state
// from which one can continue the iteration.
func (lr *leavesRetriever) GetLeaves(numLeaves int, iteratorState [][]byte, leavesParser common.TrieLeafParser, ctx context.Context) (map[string]string, [][]byte, error) {
	if check.IfNil(leavesParser) {
		return nil, nil, fmt.Errorf("nil leaves parser")
	}

	iterator, err := dfsTrieIterator.NewIterator(iteratorState, lr.db, lr.marshaller, lr.hasher)
	if err != nil {
		return nil, nil, err
	}

	leavesData, err := iterator.GetLeaves(numLeaves, lr.maxSize, leavesParser, ctx)
	if err != nil {
		return nil, nil, err
	}

	return leavesData, iterator.GetIteratorState(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (lr *leavesRetriever) IsInterfaceNil() bool {
	return lr == nil
}
