package dfsTrieIterator

import (
	"context"

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

func NewIterator(rootHash []byte, db common.TrieStorageInteractor, marshaller marshal.Marshalizer, hasher hashing.Hasher) (*dfsIterator, error) {
	data, err := trie.GetNodeDataFromHash(rootHash, keyBuilder.NewKeyBuilder(), db, marshaller, hasher)
	if err != nil {
		return nil, err
	}

	return &dfsIterator{
		nextNodes:  data,
		rootHash:   rootHash,
		db:         db,
		marshaller: marshaller,
		hasher:     hasher,
		size:       0,
	}, nil
}

func (it *dfsIterator) GetLeaves(numLeaves int, ctx context.Context) (map[string]string, error) {
	retrievedLeaves := make(map[string]string)
	nextNodes := make([]common.TrieNodeData, 0)
	for {
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
			} else {
				nextNodes = append(nextNodes, childNode)
				childrenSize += childNode.Size()
			}
		}

		it.size += childrenSize
		it.size -= it.nextNodes[0].Size()
		it.nextNodes = append(nextNodes, it.nextNodes[1:]...)
	}
}

func (it *dfsIterator) GetIteratorId() []byte {
	if len(it.nextNodes) == 0 {
		return nil
	}

	nextNodeHash := it.nextNodes[0].GetData()
	iteratorID := it.hasher.Compute(string(append(it.rootHash, nextNodeHash...)))
	return iteratorID
}

func (it *dfsIterator) Clone() common.DfsIterator {
	nextNodes := make([]common.TrieNodeData, len(it.nextNodes))
	copy(nextNodes, it.nextNodes)

	return &dfsIterator{
		nextNodes:  nextNodes,
		rootHash:   it.rootHash,
		db:         it.db,
		marshaller: it.marshaller,
		hasher:     it.hasher,
	}
}

func (it *dfsIterator) FinishedIteration() bool {
	return len(it.nextNodes) == 0
}

func (it *dfsIterator) Size() uint64 {
	return it.size + uint64(len(it.rootHash))
}

func (it *dfsIterator) IsInterfaceNil() bool {
	return it == nil
}

func checkContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
