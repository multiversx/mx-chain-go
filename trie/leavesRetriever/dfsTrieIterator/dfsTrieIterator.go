package dfsTrieIterator

import (
	"context"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever/trieNodeData"
)

// numNibbleValues bounds the symbols of a path in the iterator state: the nibbles of a key, and the terminator
// (the child of a branch holding the value of the key ending there) as the last symbol only
const (
	numNibbleValues = 16
	terminator      = 16
)

type dfsIterator struct {
	nextNodes  []common.TrieNodeData
	db         common.TrieStorageInteractor
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
}

// NewIterator creates a new DFS iterator for the trie.
func NewIterator(initialState [][]byte, db common.TrieStorageInteractor, marshaller marshal.Marshalizer, hasher hashing.Hasher) (*dfsIterator, error) {
	if check.IfNil(db) {
		return nil, trie.ErrNilDatabase
	}
	if check.IfNil(marshaller) {
		return nil, trie.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, trie.ErrNilHasher
	}
	if len(initialState) == 0 {
		return nil, trie.ErrEmptyInitialIteratorState
	}

	nextNodes, err := getNextNodesFromInitialState(initialState, uint(hasher.Size()))
	if err != nil {
		return nil, err
	}

	return &dfsIterator{
		nextNodes:  nextNodes,
		db:         db,
		marshaller: marshaller,
		hasher:     hasher,
	}, nil
}

func getNextNodesFromInitialState(initialState [][]byte, hashSize uint) ([]common.TrieNodeData, error) {
	nextNodes := make([]common.TrieNodeData, len(initialState))
	for i, state := range initialState {
		if len(state) < int(hashSize) {
			return nil, trie.ErrInvalidIteratorState
		}

		nodeHash := state[:hashSize]
		key := state[hashSize:]
		for j, symbol := range key {
			isTerminator := symbol == terminator && j == len(key)-1
			if symbol >= numNibbleValues && !isTerminator {
				return nil, trie.ErrInvalidIteratorState
			}
		}

		kb := keyBuilder.NewKeyBuilder()
		kb.BuildKey(key)
		nodeData, err := trieNodeData.NewIntermediaryNodeData(kb, nodeHash)
		if err != nil {
			return nil, err
		}
		nextNodes[i] = nodeData
	}

	return nextNodes, nil
}

func getIteratorStateFromNextNodes(nextNodes []common.TrieNodeData) [][]byte {
	iteratorState := make([][]byte, len(nextNodes))
	for i, node := range nextNodes {
		nodeHash := node.GetData()
		key := node.GetKeyBuilder().GetRawKey()

		iteratorState[i] = append(nodeHash, key...)
	}

	return iteratorState
}

// IterateLeaves walks the trie from the current iterator state and calls visit for every leaf. It stops between
// nodes when numLeaves accepted leaves (0 means no limit) or maxSize bytes, as reported by the visitor, were reached,
// when the context is done or when there are no more nodes to visit. A visitor error aborts the iteration.
func (it *dfsIterator) IterateLeaves(numLeaves int, maxSize uint64, leavesParser common.TrieLeafParser, ctx context.Context, visit common.LeafVisitor) error {
	numAccepted := 0
	size := uint64(0)
	for {
		if size >= maxSize {
			return nil
		}

		if numAccepted >= numLeaves && numLeaves != 0 {
			return nil
		}

		if it.FinishedIteration() {
			return nil
		}

		if common.IsContextDone(ctx) {
			return nil
		}

		nextNode := it.nextNodes[0]
		nodeHash := nextNode.GetData()
		childrenNodes, err := trie.GetNodeDataFromHash(nodeHash, nextNode.GetKeyBuilder(), it.db, it.marshaller, it.hasher)
		if err != nil {
			return err
		}

		nextNodes := make([]common.TrieNodeData, 0)
		for _, childNode := range childrenNodes {
			if !childNode.IsLeaf() {
				nextNodes = append(nextNodes, childNode)
				continue
			}

			key, err := childNode.GetKeyBuilder().GetKey()
			if err != nil {
				// the key of a leaf is built from the path of the state: a path that does not fit is not a state
				return fmt.Errorf("%w: %v", trie.ErrInvalidIteratorState, err)
			}

			leaf, err := leavesParser.ParseLeaf(key, childNode.GetData(), childNode.GetVersion())
			if err != nil {
				return err
			}

			accepted, leafSize, err := visit(leaf)
			if err != nil {
				return err
			}
			if accepted {
				numAccepted++
				size += leafSize
			}
		}

		it.nextNodes = append(nextNodes, it.nextNodes[1:]...)
	}
}

// GetIteratorState returns the state of the iterator from which it can be resumed by another call.
func (it *dfsIterator) GetIteratorState() [][]byte {
	if it.FinishedIteration() {
		return nil
	}

	return getIteratorStateFromNextNodes(it.nextNodes)
}

// FinishedIteration checks if the iterator has finished the iteration.
func (it *dfsIterator) FinishedIteration() bool {
	return len(it.nextNodes) == 0
}
