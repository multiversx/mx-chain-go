package trie

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
)

type iterator struct {
	currentNode node
	nextNodes   []node
	db          common.DBWriteCacher
	priority    common.StorageAccessType
}

// NewIterator creates a new instance of trie iterator
func NewIterator(trie common.Trie, priority common.StorageAccessType) (*iterator, error) {
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}

	pmt, ok := trie.(*patriciaMerkleTrie)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	trieStorage := trie.GetStorageManager()
	nextNodes, err := pmt.root.getChildren(trieStorage, priority)
	if err != nil {
		return nil, err
	}

	return &iterator{
		currentNode: pmt.root,
		nextNodes:   nextNodes,
		db:          trieStorage,
		priority:    priority,
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
	nextChildren, err := it.currentNode.getChildren(it.db, it.priority)
	if err != nil {
		return err
	}

	it.nextNodes = append(it.nextNodes, nextChildren...)
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
