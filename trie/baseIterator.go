package trie

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
)

type baseIterator struct {
	currentNode nodeWithHash
	nextNodes   []nodeWithHash
	db          common.TrieStorageInteractor
}

// newBaseIterator creates a new instance of trie iterator
func newBaseIterator(trie common.Trie, rootHash []byte) (*baseIterator, error) {
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}

	trie, err := trie.Recreate(holders.NewDefaultRootHashesHolder(rootHash), "")
	if err != nil {
		return nil, err
	}

	pmt, ok := trie.(*patriciaMerkleTrie)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	trieStorage := trie.GetStorageManager()
	rootNode := pmt.GetRootNode()
	nextNodes, err := rootNode.getChildren(trieStorage)
	if err != nil {
		return nil, err
	}

	return &baseIterator{
		currentNode: nodeWithHash{
			node: rootNode,
			hash: rootHash,
		},
		nextNodes: nextNodes,
		db:        trieStorage,
	}, nil
}

// HasNext returns true if there is a next node
func (it *baseIterator) HasNext() bool {
	return len(it.nextNodes) > 0
}

// next moves the iterator to the next node
func (it *baseIterator) next() ([]nodeWithHash, error) {
	n := it.nextNodes[0]

	if n.node == nil {
		return nil, ErrNilNode
	}

	err := n.node.isEmptyOrNil()
	if err != nil {
		return nil, ErrNilNode
	}

	it.currentNode = n
	return it.currentNode.node.getChildren(it.db)
}

// MarshalizedNode marshalizes the current node, and then returns the serialized node
func (it *baseIterator) MarshalizedNode() ([]byte, error) {
	return it.currentNode.node.getEncodedNode()
}

// GetHash returns the current node hash
func (it *baseIterator) GetHash() []byte {
	return it.currentNode.hash
}
