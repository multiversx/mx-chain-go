package trie

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

type bfsIterator struct {
	*baseIterator
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
