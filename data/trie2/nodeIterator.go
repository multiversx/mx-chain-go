// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package trie2

type nodeIterator struct {
	trie  *patriciaMerkleTrie  // Trie being iterated
	stack []*nodeIteratorState // Hierarchy of trie nodes persisting the iteration state
	path  []byte               // Path to the current node
}

// nodeIteratorState represents the iteration state at one particular node of the
// trie, which can be resumed at a later invocation.
type nodeIteratorState struct {
	hash    []byte // Hash of the node being iterated
	node    node   // Trie node being iterated
	parent  []byte // Hash of the first full ancestor node (nil if current is the root)
	index   int    // Child to be processed next
	pathlen int    // Length of the path to this node
}

func newNodeIterator(trie *patriciaMerkleTrie) *nodeIterator {
	return &nodeIterator{trie: trie}
}

func newIteratorState(hash []byte, node node, parent []byte, pathlen int) *nodeIteratorState {
	return &nodeIteratorState{hash, node, parent, -1, pathlen}
}

// Next moves the iterator to the next node
func (it *nodeIterator) Next() error {
	if it.trie.root == nil {
		return ErrNilNode
	}
	if it.trie.root.getHash() == nil {
		err := it.trie.root.setHash(it.trie.marshalizer, it.trie.hasher)
		if err != nil {
			return err
		}
	}
	err := it.peek()
	if err != nil {
		return err
	}
	return nil
}

// peek creates the next state of the iterator.
func (it *nodeIterator) peek() error {
	if len(it.stack) == 0 {
		err := it.initialise()
		return err
	}
	for len(it.stack) > 0 {
		parent := it.stack[len(it.stack)-1]
		state, path, ok := parent.node.nextChild(parent, it.path)
		if ok {
			it.push(state, &parent.index, path)
			return nil
		}
		it.pop()
	}
	return ErrIterationEnd
}

func (it *nodeIterator) initialise() error {
	root, err := it.trie.Root()
	if err != nil {
		return err
	}
	state := &nodeIteratorState{hash: root, node: it.trie.root, index: -1}
	it.push(state, nil, nil)
	return nil
}

// Hash returns the hash of the current node
func (it *nodeIterator) Hash() []byte {
	if len(it.stack) == 0 {
		return []byte{}
	}
	return it.stack[len(it.stack)-1].hash
}

// Parent returns the hash of the parent of the current node. The hash may be the one grandparent.
func (it *nodeIterator) Parent() []byte {
	if len(it.stack) == 0 {
		return []byte{}
	}
	return it.stack[len(it.stack)-1].parent
}

// Path returns the hex-encoded path to the current node.
// For leaf nodes, the last element of the path is the 'terminator symbol' 0x10.
func (it *nodeIterator) Path() []byte {
	return it.path
}

// Leaf returns true iff the current node is a leaf node.
func (it *nodeIterator) Leaf() bool {
	return hasTerm(it.path)
}

// LeafKey returns the key of the leaf. Iterator must be positioned at leaf
func (it *nodeIterator) LeafKey() ([]byte, error) {
	if it.Leaf() {
		return hexToKeyBytes(it.path), nil
	}
	return nil, ErrNotAtLeaf
}

// LeafBlob returns the content of the leaf. Iterator must be positioned at leaf
func (it *nodeIterator) LeafBlob() ([]byte, error) {
	if len(it.stack) > 0 {
		if node, ok := it.stack[len(it.stack)-1].node.(*leafNode); ok {
			return node.Value, nil
		}
	}
	return nil, ErrNotAtLeaf
}

// LeafProof returns the Merkle proof of the leaf. Iterator must be positioned at leaf
func (it *nodeIterator) LeafProof() ([][]byte, error) {
	if it.trie.root == nil {
		return nil, ErrNilNode
	}
	if it.trie.root.getHash() == nil {
		err := it.trie.root.setHash(it.trie.marshalizer, it.trie.hasher)
		if err != nil {
			return nil, err
		}
	}
	if it.Leaf() {
		proofs := make([][]byte, 0, len(it.stack))
		for _, item := range it.stack {
			node, err := item.node.getCollapsed(it.trie.marshalizer, it.trie.hasher)
			if err != nil {
				return nil, err
			}
			encNode, err := node.getEncodedNode(it.trie.marshalizer)
			if err != nil {
				return nil, err
			}
			proofs = append(proofs, encNode)
		}
		return proofs, nil
	}
	return nil, ErrNotAtLeaf
}

func (it *nodeIterator) push(state *nodeIteratorState, parentIndex *int, path []byte) {
	it.path = path
	it.stack = append(it.stack, state)
	if parentIndex != nil {
		*parentIndex++
	}
}

func (it *nodeIterator) pop() {
	parent := it.stack[len(it.stack)-1]
	it.path = it.path[:parent.pathlen]
	it.stack = it.stack[:len(it.stack)-1]
}
