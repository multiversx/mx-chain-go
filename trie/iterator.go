package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
)

type baseIterator struct {
	currentNode node
	nextNodes   []node
	db          common.DBWriteCacher
}

type dfsIterator struct {
	*baseIterator
}

type bfsIterator struct {
	*baseIterator
}

// NewDFSIterator creates a new depth first traversal iterator
func NewDFSIterator(trie common.Trie) (*dfsIterator, error) {
	bit, err := newBaseIterator(trie)
	if err != nil {
		return nil, err
	}

	return &dfsIterator{
		baseIterator: bit,
	}, nil
}

// Next moves the iterator to the next node
func (it *dfsIterator) Next() error {
	nextChildren, err := it.next()
	if err != nil {
		return err
	}

	it.nextNodes = append(nextChildren, it.nextNodes[1:]...)
	return nil
}

// NewBFSIterator creates a new level order traversal iterator
func NewBFSIterator(trie common.Trie) (*bfsIterator, error) {
	bit, err := newBaseIterator(trie)
	if err != nil {
		return nil, err
	}

	return &bfsIterator{
		baseIterator: bit,
	}, nil
}

// Next moves the iterator to the next node
func (it *bfsIterator) Next() error {
	nextChildren, err := it.next()
	if err != nil {
		return err
	}

	it.nextNodes = append(it.nextNodes, nextChildren...)
	it.nextNodes = it.nextNodes[1:]
	return nil
}

// NewIterator creates a new instance of trie iterator
func newBaseIterator(trie common.Trie) (*baseIterator, error) {
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}

	pmt, ok := trie.(*patriciaMerkleTrie)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	trieStorage := trie.GetStorageManager()
	nextNodes, err := pmt.root.getChildren(trieStorage)
	if err != nil {
		return nil, err
	}

	return &baseIterator{
		currentNode: pmt.root,
		nextNodes:   nextNodes,
		db:          trieStorage,
	}, nil
}

// HasNext returns true if there is a next node
func (it *baseIterator) HasNext() bool {
	return len(it.nextNodes) > 0
}

// next moves the iterator to the next node
func (it *baseIterator) next() ([]node, error) {
	n := it.nextNodes[0]

	err := n.isEmptyOrNil()
	if err != nil {
		return nil, ErrNilNode
	}

	it.currentNode = n
	return it.currentNode.getChildren(it.db)
}

// MarshalizedNode marshalizes the current node, and then returns the serialized node
func (it *baseIterator) MarshalizedNode() ([]byte, error) {
	err := it.currentNode.setHash()
	if err != nil {
		return nil, err
	}

	return it.currentNode.getEncodedNode()
}

// GetHash returns the current node hash
func (it *baseIterator) GetHash() ([]byte, error) {
	err := it.currentNode.setHash()
	if err != nil {
		return nil, err
	}

	return it.currentNode.getHash(), nil
}
