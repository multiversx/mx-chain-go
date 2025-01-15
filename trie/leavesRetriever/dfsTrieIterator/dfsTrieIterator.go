package dfsTrieIterator

import (
	"context"
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever/trieNodeData"
)

type dfsIterator struct {
	nextNodes  []common.TrieNodeData
	db         common.TrieStorageInteractor
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
	size       uint64
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

	size := uint64(0)
	for _, node := range nextNodes {
		size += node.Size()
	}

	return &dfsIterator{
		nextNodes:  nextNodes,
		db:         db,
		marshaller: marshaller,
		hasher:     hasher,
		size:       size,
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

// GetLeaves retrieves leaves from the trie. It stops either when the number of leaves is reached or the context is done.
func (it *dfsIterator) GetLeaves(numLeaves int, maxSize uint64, ctx context.Context) (map[string]string, error) {
	retrievedLeaves := make(map[string]string)
	for {
		nextNodes := make([]common.TrieNodeData, 0)
		if it.size >= maxSize {
			return retrievedLeaves, nil
		}

		if len(retrievedLeaves) >= numLeaves {
			return retrievedLeaves, nil
		}

		if it.FinishedIteration() {
			return retrievedLeaves, nil
		}

		if checkContextDone(ctx) {
			return retrievedLeaves, nil
		}

		nextNode := it.nextNodes[0]
		nodeHash := nextNode.GetData()
		childrenNodes, err := trie.GetNodeDataFromHash(nodeHash, nextNode.GetKeyBuilder(), it.db, it.marshaller, it.hasher)
		if err != nil {
			return nil, err
		}

		childrenSize := uint64(0)
		for _, childNode := range childrenNodes {
			if childNode.IsLeaf() {
				key, err := childNode.GetKeyBuilder().GetKey()
				if err != nil {
					return nil, err
				}

				retrievedLeaves[hex.EncodeToString(key)] = hex.EncodeToString(childNode.GetData())
				continue
			}

			nextNodes = append(nextNodes, childNode)
			childrenSize += childNode.Size()
		}

		it.size += childrenSize
		it.size -= it.nextNodes[0].Size()
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

// Size returns the size of the iterator.
func (it *dfsIterator) Size() uint64 {
	return it.size
}

// IsInterfaceNil returns true if there is no value under the interface
func (it *dfsIterator) IsInterfaceNil() bool {
	return it == nil
}

func checkContextDone(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
