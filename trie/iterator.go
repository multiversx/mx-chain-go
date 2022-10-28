package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
)

type iterator struct {
	currentNode node
	nextNodes   []node
	db          common.DBWriteCacher
}

// NewIterator creates a new instance of trie iterator
func NewIterator(trie common.Trie) (*iterator, error) {
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

	return &iterator{
		currentNode: pmt.root,
		nextNodes:   nextNodes,
		db:          trieStorage,
	}, nil
}

// HasNext returns true if there is a next node
func (it *iterator) HasNext() bool {
	return len(it.nextNodes) > 0
}

// Next moves the iterator to the next node
func (it *iterator) Next() error {
	n := it.nextNodes[0]

	err := n.isEmptyOrNil()
	if err != nil {
		return ErrNilNode
	}

	it.currentNode = n
	nextChildren, err := it.currentNode.getChildren(it.db)
	if err != nil {
		return err
	}

	it.nextNodes = insert(it.nextNodes, 1, nextChildren...)
	it.nextNodes = it.nextNodes[1:]
	return nil
}

// MarshalizedNode marshalizes the current node, and then returns the serialized node
func (it *iterator) MarshalizedNode() ([]byte, error) {
	err := it.currentNode.setHash()
	if err != nil {
		return nil, err
	}

	return it.currentNode.getEncodedNode()
}

// GetHash returns the current node hash
func (it *iterator) GetHash() ([]byte, error) {
	err := it.currentNode.setHash()
	if err != nil {
		return nil, err
	}

	return it.currentNode.getHash(), nil
}

// TODO: move this to core sliceUtil, when generics will be adopted
func insert(s []node, k int, vs ...node) []node {
	n := len(s) + len(vs)
	if n <= cap(s) {
		s2 := s[:n]
		copy(s2[k+len(vs):], s[k:])
		copy(s2[k:], vs)
		return s2
	}

	s2 := make([]node, len(s)+len(vs))
	copy(s2, s[:k])
	copy(s2[k:], vs)
	copy(s2[k+len(vs):], s[k:])

	return s2
}
