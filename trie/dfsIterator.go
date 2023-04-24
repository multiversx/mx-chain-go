package trie

import "github.com/multiversx/mx-chain-go/common"

type dfsIterator struct {
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
