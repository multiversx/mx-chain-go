package trie

import "github.com/ElrondNetwork/elrond-go/data"

type iterator struct {
	currentNode node
	nextNodes   []node
}

// NewIterator creates a new instance of trie iterator
func NewIterator(trie data.Trie) (*iterator, error) {
	if trie == nil || trie.IsInterfaceNil() {
		return nil, ErrNilTrie
	}

	pmt, ok := trie.(*patriciaMerkleTrie)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	nextNodes, err := pmt.root.getChildren()
	if err != nil {
		return nil, err
	}

	return &iterator{
		currentNode: pmt.root,
		nextNodes:   nextNodes,
	}, nil
}

// HasNext returns true if there is a next node
func (it *iterator) HasNext() bool {
	if len(it.nextNodes) == 0 {
		return false
	}

	return true
}

// Next moves the iterator to the next node
func (it *iterator) Next() error {
	it.currentNode = it.nextNodes[0]

	nextChildren, err := it.currentNode.getChildren()
	if err != nil {
		return err
	}

	it.nextNodes = append(it.nextNodes, nextChildren...)
	it.nextNodes = it.nextNodes[1:]
	return nil
}

// GetMarshalizedNode marshalizes the current node, and then returns the serialized node
func (it *iterator) GetMarshalizedNode() ([]byte, error) {
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
