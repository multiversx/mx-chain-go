//// Copyright 2014 The go-ethereum Authors
//// This file is part of the go-ethereum library.
////
//// The go-ethereum library is free software: you can redistribute it and/or modify
//// it under the terms of the GNU Lesser General Public License as published by
//// the Free Software Foundation, either version 3 of the License, or
//// (at your option) any later version.
////
//// The go-ethereum library is distributed in the hope that it will be useful,
//// but WITHOUT ANY WARRANTY; without even the implied warranty of
//// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
//// GNU Lesser General Public License for more details.
////
//// You should have received a copy of the GNU Lesser General Public License
//// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package state

//
//import (
//	"bytes"
//	"errors"
//)
//
//// nodeIteratorState represents the iteration state at one particular node of the
//// trie, which can be resumed at a later invocation.
//type nodeIteratorState struct {
//	hash    []byte // Hash of the node being iterated (nil if not standalone)
//	//node    node   // Trie node being iterated
//	parent  []byte // Hash of the first full ancestor node (nil if current is the root)
//	index   int    // Child to be processed next
//	pathlen int    // Length of the path to this node
//}
//
//type nodeIterator struct {
//	hasher Hasher
//	trie   Trie                // Trie being iterated
//	stack []*nodeIteratorState // Hierarchy of trie nodes persisting the iteration state
//	path  []byte               // Path to the current node
//	err   error                // Failure set in case of an internal error in the iterator
//}
//
//// errIteratorEnd is stored in nodeIterator.err when iteration is done.
//var errIteratorEnd = errors.New("end of iteration")
//
//func newNodeIterator(trie Trie, start []byte, hasher Hasher) NodeIterator {
//	if bytes.Equal(trie.Hash(), hasher.EmptyHash()) {
//		it := new(nodeIterator)
//		it.hasher = hasher
//
//		return it
//	}
//	it := &nodeIterator{trie: trie}
//	it.err = it.seek(start)
//	it.hasher = hasher
//	return it
//}
//
//func (it *nodeIterator) Hash() []byte {
//	if len(it.stack) == 0 {
//		return it.hasher.EmptyHash()
//	}
//	return it.stack[len(it.stack)-1].hash
//}
//
//func (it *nodeIterator) Parent() []byte {
//	if len(it.stack) == 0 {
//		return it.hasher.EmptyHash()
//	}
//	return it.stack[len(it.stack)-1].parent
//}
//
//func (it *nodeIterator) Leaf() bool {
//	return hasTerm(it.path)
//}
//
//func (it *nodeIterator) LeafKey() []byte {
//	if len(it.stack) > 0 {
//		if _, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
//			return hexToKeybytes(it.path)
//		}
//	}
//	panic("not at leaf")
//}
//
//func (it *nodeIterator) LeafBlob() []byte {
//	if len(it.stack) > 0 {
//		if node, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
//			return []byte(node)
//		}
//	}
//	panic("not at leaf")
//}
//
//func (it *nodeIterator) LeafProof() [][]byte {
//	if len(it.stack) > 0 {
//		if _, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
//			hasher := newHasher(0, 0, nil)
//			proofs := make([][]byte, 0, len(it.stack))
//
//			for i, item := range it.stack[:len(it.stack)-1] {
//				// Gather nodes that end up as hash nodes (or the root)
//				node, _, _ := hasher.hashChildren(item.node, nil)
//				hashed, _ := hasher.store(node, nil, false)
//				if _, ok := hashed.(hashNode); ok || i == 0 {
//					enc, _ := rlp.EncodeToBytes(node)
//					proofs = append(proofs, enc)
//				}
//			}
//			return proofs
//		}
//	}
//	panic("not at leaf")
//}
//
//func (it *nodeIterator) Path() []byte {
//	return it.path
//}
//
//func (it *nodeIterator) Error() error {
//	if it.err == errIteratorEnd {
//		return nil
//	}
//	if seek, ok := it.err.(seekError); ok {
//		return seek.err
//	}
//	return it.err
//}
//
//// Next moves the iterator to the next node, returning whether there are any
//// further nodes. In case of an internal error this method returns false and
//// sets the Error field to the encountered failure. If `descend` is false,
//// skips iterating over any subnodes of the current node.
//func (it *nodeIterator) Next(descend bool) bool {
//	if it.err == errIteratorEnd {
//		return false
//	}
//	if seek, ok := it.err.(seekError); ok {
//		if it.err = it.seek(seek.key); it.err != nil {
//			return false
//		}
//	}
//	// Otherwise step forward with the iterator and report any errors.
//	state, parentIndex, path, err := it.peek(descend)
//	it.err = err
//	if it.err != nil {
//		return false
//	}
//	it.push(state, parentIndex, path)
//	return true
//}
//
//func (it *nodeIterator) seek(prefix []byte) error {
//	// The path we're looking for is the hex encoded key without terminator.
//	key := keybytesToHex(prefix)
//	key = key[:len(key)-1]
//	// Move forward until we're just before the closest match to key.
//	for {
//		state, parentIndex, path, err := it.peek(bytes.HasPrefix(key, it.path))
//		if err == errIteratorEnd {
//			return errIteratorEnd
//		} else if err != nil {
//			return seekError{prefix, err}
//		} else if bytes.Compare(path, key) >= 0 {
//			return nil
//		}
//		it.push(state, parentIndex, path)
//	}
//}
//
//// peek creates the next state of the iterator.
//func (it *nodeIterator) peek(descend bool) (*nodeIteratorState, *int, []byte, error) {
//	if len(it.stack) == 0 {
//		// Initialize the iterator if we've just started.
//		root := it.trie.Hash()
//		state := &nodeIteratorState{node: it.trie.root, index: -1}
//		if root != emptyRoot {
//			state.hash = root
//		}
//		err := state.resolve(it.trie, nil)
//		return state, nil, nil, err
//	}
//	if !descend {
//		// If we're skipping children, pop the current node first
//		it.pop()
//	}
//
//	// Continue iteration to the next child
//	for len(it.stack) > 0 {
//		parent := it.stack[len(it.stack)-1]
//		ancestor := parent.hash
//		if (ancestor == common.Hash{}) {
//			ancestor = parent.parent
//		}
//		state, path, ok := it.nextChild(parent, ancestor)
//		if ok {
//			if err := state.resolve(it.trie, path); err != nil {
//				return parent, &parent.index, path, err
//			}
//			return state, &parent.index, path, nil
//		}
//		// No more child nodes, move back up.
//		it.pop()
//	}
//	return nil, nil, nil, errIteratorEnd
//}
//
//func (st *nodeIteratorState) resolve(tr *Trie, path []byte) error {
//	if hash, ok := st.node.(hashNode); ok {
//		resolved, err := tr.resolveHash(hash, path)
//		if err != nil {
//			return err
//		}
//		st.node = resolved
//		st.hash = hash
//	}
//	return nil
//}
//
//func (it *nodeIterator) nextChild(parent *nodeIteratorState, ancestor common.Hash) (*nodeIteratorState, []byte, bool) {
//	switch node := parent.node.(type) {
//	case *fullNode:
//		// Full node, move to the first non-nil child.
//		for i := parent.index + 1; i < len(node.Children); i++ {
//			child := node.Children[i]
//			if child != nil {
//				hash, _ := child.cache()
//				state := &nodeIteratorState{
//					hash:    hash,
//					node:    child,
//					parent:  ancestor,
//					index:   -1,
//					pathlen: len(it.path),
//				}
//				path := append(it.path, byte(i))
//				parent.index = i - 1
//				return state, path, true
//			}
//		}
//	case *shortNode:
//		// Short node, return the pointer singleton child
//		if parent.index < 0 {
//			hash, _ := node.Val.cache()
//			state := &nodeIteratorState{
//				hash:    hash,
//				node:    node.Val,
//				parent:  ancestor,
//				index:   -1,
//				pathlen: len(it.path),
//			}
//			path := append(it.path, node.Key...)
//			return state, path, true
//		}
//	}
//	return parent, it.path, false
//}
//
//func (it *nodeIterator) push(state *nodeIteratorState, parentIndex *int, path []byte) {
//	it.path = path
//	it.stack = append(it.stack, state)
//	if parentIndex != nil {
//		*parentIndex++
//	}
//}
//
//func (it *nodeIterator) pop() {
//	parent := it.stack[len(it.stack)-1]
//	it.path = it.path[:parent.pathlen]
//	it.stack = it.stack[:len(it.stack)-1]
//}
//
//func compareNodes(a, b NodeIterator) int {
//	if cmp := bytes.Compare(a.Path(), b.Path()); cmp != 0 {
//		return cmp
//	}
//	if a.Leaf() && !b.Leaf() {
//		return -1
//	} else if b.Leaf() && !a.Leaf() {
//		return 1
//	}
//	if cmp := bytes.Compare(a.Hash(), b.Hash()); cmp != 0 {
//		return cmp
//	}
//	if a.Leaf() && b.Leaf() {
//		return bytes.Compare(a.LeafBlob(), b.LeafBlob())
//	}
//	return 0
//}
