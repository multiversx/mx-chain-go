package trie2

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/encoding"
)

var emptyHash []byte

// nodeIteratorState represents the iteration state at one particular node of the
// trie, which can be resumed at a later invocation.
type nodeIteratorState struct {
	hash    []byte // Hash of the node being iterated
	node    Node   // Trie node being iterated
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

func (it *nodeIterator) Error() error {
	return it.err
}

func (it *nodeIterator) Hash() []byte {
	if len(it.stack) == 0 {
		return []byte{}
	}
	return it.stack[len(it.stack)-1].hash
}

func (it *nodeIterator) Parent() []byte {
	if len(it.stack) == 0 {
		return []byte{}
	}
	return it.stack[len(it.stack)-1].parent
}

func (it *nodeIterator) Path() []byte {
	return it.path
}

func (it *nodeIterator) Leaf() bool {
	return encoding.HasTerm(it.path)
}

func (it *nodeIterator) LeafKey() []byte {
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
			return encoding.HexToKeyBytes(it.path)
		}
	}
	panic("not at leaf")
}

func (it *nodeIterator) LeafBlob() []byte {
	if len(it.stack) > 0 {
		if node, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
			return []byte(node)
		}
	}
	panic("not at leaf")
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
			node, err := SetHash(state.node)
			if err != nil {
				return nil, nil, nil, err
			}
			state.hash = node.GetHash()
			return state, &parent.index, path, nil
		}
		// No more child nodes, move back up.
		it.pop()
	}
	return nil, nil, nil, errIteratorEnd
}

func (it *nodeIterator) nextChild(parent *nodeIteratorState, ancestor []byte) (*nodeIteratorState, []byte, bool) {
	switch node := parent.node.(type) {
	case *fullNode:
		// Full node, move to the first non-nil child.
		for i := parent.index + 1; i < len(node.Children); i++ {
			child := node.Children[i]
			if child != nil {
				hash := child.GetHash()
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
			hash := node.Val.GetHash()
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
