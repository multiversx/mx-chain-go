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

package trie3

import (
	"bytes"
)

var emptyHash []byte

// nodeIteratorState represents the iteration state at one particular node of the
// trie, which can be resumed at a later invocation.
type nodeIteratorState struct {
	hash    []byte // Hash of the node being iterated
	node    node   // Trie node being iterated
	parent  []byte // Hash of the first full ancestor node (nil if current is the root)
	index   int    // Child to be processed next
	pathlen int    // Length of the path to this node
}

type nodeIterator struct {
	trie  *patriciaMerkleTree  // Trie being iterated
	stack []*nodeIteratorState // Hierarchy of trie nodes persisting the iteration state
	path  []byte               // Path to the current node
}

func newNodeIterator(trie *patriciaMerkleTree) *nodeIterator {
	return &nodeIterator{trie: trie}
}

// Next moves the iterator to the next node, returning whether there are any
// further nodes.
func (it *nodeIterator) Next() (bool, error) {

	// Otherwise step forward with the iterator and report any errors.
	err := it.peek()
	if err != nil {
		return false, err
	}
	return true, nil
}

// peek creates the next state of the iterator.
func (it *nodeIterator) peek() error {
	var state *nodeIteratorState
	if len(it.stack) == 0 {
		root, err := it.trie.Root()
		if err != nil {
			return err
		}
		state = &nodeIteratorState{node: it.trie.root, index: -1}
		if !bytes.Equal(root, emptyHash) {
			state.hash = root
		}
		it.push(state, nil, nil)
		return nil
	}
	// Continue iteration to the next child
	var parent *nodeIteratorState
	for len(it.stack) > 0 {
		parent = it.stack[len(it.stack)-1]
		ancestor := parent.hash
		if bytes.Equal(ancestor, emptyHash) {
			ancestor = parent.parent
		}
		state, path, ok := it.nextChild(parent, ancestor)
		if ok {
			err := state.node.setHash(it.trie.marshalizer, it.trie.hasher)
			if err != nil {
				return err
			}
			state.hash = state.node.getHash()
			it.push(state, &parent.index, path)
			return nil
		}
		// No more child nodes, move back up.
		it.pop()
	}

	return ErrIterationEnd
}

func (it *nodeIterator) nextChild(parent *nodeIteratorState, ancestor []byte) (*nodeIteratorState, []byte, bool) {
	switch node := parent.node.(type) {
	case *branchNode:
		// Full node, move to the first non-nil child.
		for i := parent.index + 1; i < len(node.children); i++ {
			child := node.children[i]

			if child != nil {
				hash := child.getHash()
				state := &nodeIteratorState{
					hash:    hash,
					node:    child,
					parent:  ancestor,
					index:   -1,
					pathlen: len(it.path),
				}
				path := append(it.path, byte(i))
				if child, ok := child.(*leafNode); ok {
					path = append(path, child.Key...)
				}
				parent.index = i - 1
				return state, path, true
			}
		}
	case *extensionNode:
		// Short node, return the pointer singleton child
		if parent.index < 0 {
			hash := node.child.getHash()
			state := &nodeIteratorState{
				hash:    hash,
				node:    node.child,
				parent:  ancestor,
				index:   -1,
				pathlen: len(it.path),
			}
			path := append(it.path, node.Key...)
			return state, path, true
		}
	}
	return parent, it.path, false
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
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(*leafNode); ok {
			return hexToKeyBytes(it.path), nil
		}
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
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(*leafNode); ok {
			proofs := make([][]byte, 0, len(it.stack))

			for _, item := range it.stack[:len(it.stack)] {
				// Gather nodes that end up as hash nodes (or the root)
				err := item.node.setHash(it.trie.marshalizer, it.trie.hasher)
				if err != nil {
					return nil, err
				}
				node := item.node.getCollapsed()

				encNode, err := node.getEncodedNodeUsing(it.trie.marshalizer)
				if err != nil {
					return nil, err
				}
				proofs = append(proofs, encNode)

			}
			return proofs, nil
		}
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
