package trie3

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
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
	state, parentIndex, path, err := it.peek()
	if err != nil {
		return false, err
	}
	it.push(state, parentIndex, path)
	return true, nil
}

// peek creates the next state of the iterator.
func (it *nodeIterator) peek() (*nodeIteratorState, *int, []byte, error) {
	if len(it.stack) == 0 {
		// Initialize the iterator if we've just started.
		root, err := it.trie.Root()
		if err != nil {
			return nil, nil, nil, err
		}
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
			node, err := it.trie.setHash(state.node)
			if err != nil {
				return nil, nil, nil, err
			}
			state.hash = node.getHash()
			return state, &parent.index, path, nil
		}
		// No more child nodes, move back up.
		it.pop()
	}
	return nil, nil, nil, trie2.ErrIterationEnd
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
	return nil, trie2.ErrNotAtLeaf
}

// LeafBlob returns the content of the leaf. Iterator must be positioned at leaf
func (it *nodeIterator) LeafBlob() ([]byte, error) {
	if len(it.stack) > 0 {
		if node, ok := it.stack[len(it.stack)-1].node.(*leafNode); ok {
			return node.Value, nil
		}
	}
	return nil, trie2.ErrNotAtLeaf
}

// LeafProof returns the Merkle proof of the leaf. Iterator must be positioned at leaf
func (it *nodeIterator) LeafProof() ([][]byte, error) {
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(*leafNode); ok {
			proofs := make([][]byte, 0, len(it.stack))

			for _, item := range it.stack[:len(it.stack)] {
				// Gather nodes that end up as hash nodes (or the root)
				node, _ := it.trie.setHash(item.node)
				collapsed, _ := it.trie.collapseNode(node)

				encNode, err := it.trie.encodeNode(collapsed)
				if err != nil {
					return nil, err
				}
				proofs = append(proofs, encNode)

			}
			return proofs, nil
		}
	}
	return nil, trie2.ErrNotAtLeaf
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
