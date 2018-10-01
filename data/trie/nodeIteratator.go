package eth

import (
	"bytes"
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/rlp"
)

var emptyState = DefaultHasher().Compute(string(""))

// Iterator is a key-value trie iterator that traverses a Trie.
type Iterator struct {
	nodeIt NodeIterator

	Key   []byte // Current data key on which the iterator is positioned on
	Value []byte // Current data value on which the iterator is positioned on
	Err   error
}

// NewIterator creates a new key-value iterator from a node iterator
func NewIterator(it NodeIterator) *Iterator {
	return &Iterator{
		nodeIt: it,
	}
}

// Next moves the iterator forward one key-value entry.
func (it *Iterator) Next() bool {
	for it.nodeIt.Next(true) {
		if it.nodeIt.Leaf() {
			it.Key = it.nodeIt.LeafKey()
			it.Value = it.nodeIt.LeafBlob()
			return true
		}
	}
	it.Key = nil
	it.Value = nil
	it.Err = it.nodeIt.Error()
	return false
}

// Prove generates the Merkle proof for the leaf node the iterator is currently
// positioned on.
func (it *Iterator) Prove() [][]byte {
	return it.nodeIt.LeafProof()
}

// NodeIterator is an iterator to traverse the trie pre-order.
type NodeIterator interface {
	// Next moves the iterator to the next node. If the parameter is false, any child
	// nodes will be skipped.
	Next(bool) bool

	// Error returns the error status of the iterator.
	Error() error

	// Hash returns the hash of the current node.
	Hash() encoding.Hash

	// Parent returns the hash of the parent of the current node. The hash may be the one
	// grandparent if the immediate parent is an internal node with no hash.
	Parent() encoding.Hash

	// Path returns the hex-encoded path to the current node.
	// Callers must not retain references to the return value after calling Next.
	// For leaf nodes, the last element of the path is the 'terminator symbol' 0x10.
	Path() []byte

	// Leaf returns true iff the current node is a leaf node.
	Leaf() bool

	// LeafKey returns the key of the leaf. The method panics if the iterator is not
	// positioned at a leaf. Callers must not retain references to the value after
	// calling Next.
	LeafKey() []byte

	// LeafBlob returns the content of the leaf. The method panics if the iterator
	// is not positioned at a leaf. Callers must not retain references to the value
	// after calling Next.
	LeafBlob() []byte

	// LeafProof returns the Merkle proof of the leaf. The method panics if the
	// iterator is not positioned at a leaf. Callers must not retain references
	// to the value after calling Next.
	LeafProof() [][]byte
}

// nodeIteratorState represents the iteration state at one particular node of the
// trie, which can be resumed at a later invocation.
type nodeIteratorState struct {
	hash    encoding.Hash // Hash of the node being iterated (nil if not standalone)
	node    node          // Trie node being iterated
	parent  encoding.Hash // Hash of the first full ancestor node (nil if current is the root)
	index   int           // Child to be processed next
	pathlen int           // Length of the path to this node
}

type nodeIterator struct {
	trie  *Trie                // Trie being iterated
	stack []*nodeIteratorState // Hierarchy of trie nodes persisting the iteration state
	path  []byte               // Path to the current node
	err   error                // Failure set in case of an internal error in the iterator
}

// errIteratorEnd is stored in nodeIterator.err when iteration is done.
var errIteratorEnd = errors.New("end of iteration")

// seekError is stored in nodeIterator.err if the initial seek has failed.
type seekError struct {
	key []byte
	err error
}

func (e seekError) Error() string {
	return "seek error: " + e.err.Error()
}

func newNodeIterator(trie *Trie, start []byte) NodeIterator {
	h := encoding.Hash{}
	h.SetBytes(emptyState)

	if trie.Hash() == h {
		return new(nodeIterator)
	}
	it := &nodeIterator{trie: trie}
	it.err = it.seek(start)
	return it
}

func (it *nodeIterator) Hash() encoding.Hash {
	if len(it.stack) == 0 {
		return encoding.Hash{}
	}
	return it.stack[len(it.stack)-1].hash
}

func (it *nodeIterator) Parent() encoding.Hash {
	if len(it.stack) == 0 {
		return encoding.Hash{}
	}
	return it.stack[len(it.stack)-1].parent
}

func (it *nodeIterator) Leaf() bool {
	return encoding.HasTerm(it.path)
}

func (it *nodeIterator) LeafKey() []byte {
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
			return encoding.HexToKeybytes(it.path)
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

func (it *nodeIterator) LeafProof() [][]byte {
	if len(it.stack) > 0 {
		if _, ok := it.stack[len(it.stack)-1].node.(valueNode); ok {
			hasher := newHasher(0, 0, nil)
			proofs := make([][]byte, 0, len(it.stack))

			for i, item := range it.stack[:len(it.stack)-1] {
				// Gather nodes that end up as hash nodes (or the root)
				node, _, _ := hasher.hashChildren(item.node, nil)
				hashed, _ := hasher.store(node, nil, false)
				if _, ok := hashed.(hashNode); ok || i == 0 {
					enc, _ := rlp.EncodeToBytes(node)
					proofs = append(proofs, enc)
				}
			}
			return proofs
		}
	}
	panic("not at leaf")
}

func (it *nodeIterator) Path() []byte {
	return it.path
}

func (it *nodeIterator) Error() error {
	if it.err == errIteratorEnd {
		return nil
	}
	if seek, ok := it.err.(seekError); ok {
		return seek.err
	}
	return it.err
}

// Next moves the iterator to the next node, returning whether there are any
// further nodes. In case of an internal error this method returns false and
// sets the Error field to the encountered failure. If `descend` is false,
// skips iterating over any subnodes of the current node.
func (it *nodeIterator) Next(descend bool) bool {
	if it.err == errIteratorEnd {
		return false
	}
	if seek, ok := it.err.(seekError); ok {
		if it.err = it.seek(seek.key); it.err != nil {
			return false
		}
	}
	// Otherwise step forward with the iterator and report any errors.
	state, parentIndex, path, err := it.peek(descend)
	it.err = err
	if it.err != nil {
		return false
	}
	it.push(state, parentIndex, path)
	return true
}

func (it *nodeIterator) seek(prefix []byte) error {
	// The path we're looking for is the hex encoded key without terminator.
	key := encoding.KeybytesToHex(prefix)
	key = key[:len(key)-1]
	// Move forward until we're just before the closest match to key.
	for {
		state, parentIndex, path, err := it.peek(encoding.HasPrefix(key, it.path))
		if err == errIteratorEnd {
			return errIteratorEnd
		} else if err != nil {
			return seekError{prefix, err}
		} else if bytes.Compare(path, key) >= 0 {
			return nil
		}
		it.push(state, parentIndex, path)
	}
}

// peek creates the next state of the iterator.
func (it *nodeIterator) peek(descend bool) (*nodeIteratorState, *int, []byte, error) {
	if len(it.stack) == 0 {
		// Initialize the iterator if we've just started.
		root := it.trie.Hash()
		state := &nodeIteratorState{node: it.trie.root, index: -1}
		if root != emptyRoot {
			state.hash = root
		}
		err := state.resolve(it.trie, nil)
		return state, nil, nil, err
	}
	if !descend {
		// If we're skipping children, pop the current node first
		it.pop()
	}

	// Continue iteration to the next child
	for len(it.stack) > 0 {
		parent := it.stack[len(it.stack)-1]
		ancestor := parent.hash
		if (ancestor == encoding.Hash{}) {
			ancestor = parent.parent
		}
		state, path, ok := it.nextChild(parent, ancestor)
		if ok {
			if err := state.resolve(it.trie, path); err != nil {
				return parent, &parent.index, path, err
			}
			return state, &parent.index, path, nil
		}
		// No more child nodes, move back up.
		it.pop()
	}
	return nil, nil, nil, errIteratorEnd
}

func (st *nodeIteratorState) resolve(tr *Trie, path []byte) error {
	if hash, ok := st.node.(hashNode); ok {
		resolved, err := tr.resolveHash(hash, path)
		if err != nil {
			return err
		}
		st.node = resolved
		st.hash = encoding.BytesToHash(hash)
	}
	return nil
}

func (it *nodeIterator) nextChild(parent *nodeIteratorState, ancestor encoding.Hash) (*nodeIteratorState, []byte, bool) {
	switch node := parent.node.(type) {
	case *fullNode:
		// Full node, move to the first non-nil child.
		for i := parent.index + 1; i < len(node.Children); i++ {
			child := node.Children[i]
			if child != nil {
				hash, _ := child.cache()
				state := &nodeIteratorState{
					hash:    encoding.BytesToHash(hash),
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
			hash, _ := node.Val.cache()
			state := &nodeIteratorState{
				hash:    encoding.BytesToHash(hash),
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
