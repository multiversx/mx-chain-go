package patriciaMerkleTrie

import (
	"bytes"
	"errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/patriciaMerkleTrie/encoding"
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
	trie  *PatriciaMerkleTree  // Trie being iterated
	stack []*nodeIteratorState // Hierarchy of trie nodes persisting the iteration state
	path  []byte               // Path to the current node
	err   error                // Failure set in case of an internal error in the iterator
}

func newNodeIterator(trie *PatriciaMerkleTree) *nodeIterator {
	it := &nodeIterator{trie: trie}
	return it
}

// Next moves the iterator to the next node, returning whether there are any
// further nodes. In case of an internal error this method returns false and
// sets the Error field to the encountered failure.
func (it *nodeIterator) Next() bool {
	if it.err != nil {
		return false
	}

	// Otherwise step forward with the iterator and report any errors.
	state, parentIndex, path, err := it.peek()
	it.err = err
	if it.err != nil {
		return false
	}
	it.push(state, parentIndex, path)
	return true
}

// peek creates the next state of the iterator.
func (it *nodeIterator) peek() (*nodeIteratorState, *int, []byte, error) {
	if len(it.stack) == 0 {
		// Initialize the iterator if we've just started.
		root := it.trie.Root()
		state := &nodeIteratorState{node: it.trie.root, index: -1}
		if !bytes.Equal(root, emptyHash) {
			state.hash = root
		}
		return state, nil, nil, nil
	}

	// Continue iteration to the next child
	for len(it.stack) > 0 {
		parent := it.stack[len(it.stack)-1]
		ancestor := parent.hash
		if bytes.Equal(ancestor, emptyHash) {
			ancestor = parent.parent
		}
		state, path, ok := it.nextChild(parent, ancestor)
		if ok {
			_, node, err := it.trie.setHash(state.node)
			if err != nil {
				return nil, nil, nil, err
			}
			state.hash = node.getHash()
			return state, &parent.index, path, nil
		}
		// No more child nodes, move back up.
		it.pop()
	}
	return nil, nil, nil, errors.New("end of iteration")
}

func (it *nodeIterator) nextChild(parent *nodeIteratorState, ancestor []byte) (*nodeIteratorState, []byte, bool) {
	switch node := parent.node.(type) {
	case *fullNode:
		// Full node, move to the first non-nil child.
		for i := parent.index + 1; i < len(node.Children); i++ {
			child := node.Children[i]
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
				parent.index = i - 1
				return state, path, true
			}
		}
	case *shortNode:
		// Short node, return the pointer singleton child
		if parent.index < 0 {
			hash := node.Val.getHash()
			state := &nodeIteratorState{
				hash:    hash,
				node:    node.Val,
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

func (it *nodeIterator) Error() error {
	return it.err
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
	return encoding.HasTerm(it.path)
}

// LeafKey returns the key of the leaf. The method panics if the iterator is not
// positioned at a leaf.
func (it *nodeIterator) LeafKey() []byte {
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
			return encoding.HexToKeyBytes(it.path)
		}
	}
	panic("not at leaf")
}

// LeafBlob returns the content of the leaf. The method panics if the iterator
// is not positioned at a leaf.
func (it *nodeIterator) LeafBlob() []byte {
	if len(it.stack) > 0 {
		if node, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
			return []byte(node)
		}
	}
	panic("not at leaf")
}

// LeafProof returns the Merkle proof of the leaf. The method panics if the
// iterator is not positioned at a leaf.
func (it *nodeIterator) LeafProof() [][]byte {
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
			proofs := make([][]byte, 0, len(it.stack))

			for i, item := range it.stack[:len(it.stack)-1] {
				// Gather nodes that end up as hash nodes (or the root)
				node, _, _ := it.trie.hashChildren(item.node)
				hashed, _ := it.trie.hash(node)
				if _, ok := hashed.(hashNode); ok || i == 0 {
					n, err := it.trie.marshalizer.Marshal(node)
					if err != nil {
						return nil
					}
					proofs = append(proofs, n)
				}
			}
			return proofs
		}
	}
	panic("not at leaf")
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
