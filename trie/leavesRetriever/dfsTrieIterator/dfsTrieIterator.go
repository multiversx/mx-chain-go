package dfsTrieIterator

import (
	"context"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
)

type dfsIterator struct {
	nextNodes  []common.TrieNodeData
	rootHash   []byte
	db         common.TrieStorageInteractor
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
	size       uint64
}

// NewIterator creates a new DFS iterator for the trie.
func NewIterator(rootHash []byte, db common.TrieStorageInteractor, marshaller marshal.Marshalizer, hasher hashing.Hasher) (*dfsIterator, error) {
	if check.IfNil(db) {
		return nil, trie.ErrNilDatabase
	}
	if check.IfNil(marshaller) {
		return nil, trie.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, trie.ErrNilHasher
	}

	data, err := trie.GetNodeDataFromHash(rootHash, keyBuilder.NewKeyBuilder(), db, marshaller, hasher)
	if err != nil {
		return nil, err
	}

	size := uint64(0)
	for _, node := range data {
		size += node.Size()
	}

	return &dfsIterator{
		nextNodes:  data,
		rootHash:   rootHash,
		db:         db,
		marshaller: marshaller,
		hasher:     hasher,
		size:       size,
	}, nil
}

// GetLeaves retrieves leaves from the trie. It stops either when the number of leaves is reached or the context is done.
func (it *dfsIterator) GetLeaves(numLeaves int, ctx context.Context) (map[string]string, error) {
	retrievedLeaves := make(map[string]string)
	for {
		nextNodes := make([]common.TrieNodeData, 0)
		if len(retrievedLeaves) >= numLeaves {
			return retrievedLeaves, nil
		}

		if len(it.nextNodes) == 0 {
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

				retrievedLeaves[string(key)] = string(childNode.GetData())
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

// GetIteratorId returns the ID of the iterator.
func (it *dfsIterator) GetIteratorId() []byte {
	if len(it.nextNodes) == 0 {
		return nil
	}

	nextNodeHash := it.nextNodes[0].GetData()
	iteratorID := it.hasher.Compute(string(append(it.rootHash, nextNodeHash...)))
	return iteratorID
}

// Clone creates a copy of the iterator.
func (it *dfsIterator) Clone() common.DfsIterator {
	nextNodes := make([]common.TrieNodeData, len(it.nextNodes))
	copy(nextNodes, it.nextNodes)

	return &dfsIterator{
		nextNodes:  nextNodes,
		rootHash:   it.rootHash,
		db:         it.db,
		marshaller: it.marshaller,
		hasher:     it.hasher,
		size:       it.size,
	}
}

// FinishedIteration checks if the iterator has finished the iteration.
func (it *dfsIterator) FinishedIteration() bool {
	return len(it.nextNodes) == 0
}

// Size returns the size of the iterator.
func (it *dfsIterator) Size() uint64 {
	return it.size + uint64(len(it.rootHash))
}

// IsInterfaceNil returns true if there is no value under the interface
func (it *dfsIterator) IsInterfaceNil() bool {
	return it == nil
}

// TODO add context nil test
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
