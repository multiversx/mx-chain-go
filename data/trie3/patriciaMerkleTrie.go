package trie3

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

const (
	extension = iota
	leaf
	branch
)
const firstByte = 0
const moreThanOneChildren = -2

type patriciaMerkleTree struct {
	root        node
	dbw         DBWriteCacher
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

func NewTrie(hsh hashing.Hasher, msh marshal.Marshalizer, dbw DBWriteCacher) (*patriciaMerkleTree, error) {

	if hsh == nil {
		return nil, ErrNilHasher
	}

	if msh == nil {
		return nil, ErrNilMarshalizer
	}

	return &patriciaMerkleTree{dbw: dbw, hasher: hsh, marshalizer: msh}, nil
}

// NewNodeIterator returns a new node iterator for the current trie
func (tr *patriciaMerkleTree) NewNodeIterator() NodeIterator {
	return newNodeIterator(tr)
}

func (tr *patriciaMerkleTree) Get(key []byte) ([]byte, error) {
	hexKey := keyBytesToHex(key)
	value, err := tr.tryGet(tr.root, hexKey)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (tr *patriciaMerkleTree) tryGet(n node, key []byte) (value []byte, err error) {
	switch n := n.(type) {
	case *extensionNode:
		keyTooShort := len(key) < len(n.Key)
		if keyTooShort {
			return nil, nil
		}

		keysDontMatch := !bytes.Equal(n.Key, key[:len(n.Key)])
		if keysDontMatch {
			return nil, nil
		}

		key = key[len(n.Key):]
		if n.isCollapsed() {
			n, err = tr.resolveExtNode(n)
			if err != nil {
				return nil, err
			}
		}
		value, err = tr.tryGet(n.child, key)
		if err != nil {
			return nil, err
		}

		return value, err
	case *branchNode:
		childPos := key[firstByte]
		key, err = removeFirstByte(key)
		if err != nil {
			return nil, err
		}

		if n.isCollapsed(childPos) {
			n, err = tr.resolveBrNode(n)
			if err != nil {
				return nil, err
			}
		}

		value, err = tr.tryGet(n.children[childPos], key)
		return value, err
	case *leafNode:
		if n.isCollapsed(tr) {
			n, err = tr.resolveLeafNode(n)
			if err != nil {
				return nil, err
			}
		}
		if bytes.Equal(key, n.Key) {
			return n.Value, nil
		}
		return nil, nil
	case nil:
		return nil, nil
	default:
		return nil, ErrInvalidNode
	}
}

func (tr *patriciaMerkleTree) resolveExtNode(en *extensionNode) (*extensionNode, error) {
	child, err := tr.getNodeFromDB(en.EncodedChild)
	if err != nil {
		return nil, err
	}
	en.child = child
	return en, nil
}

func (tr *patriciaMerkleTree) resolveBrNode(bn *branchNode) (*branchNode, error) {
	for i := range bn.children {
		if bn.EncodedChildren[i] != nil {
			child, err := tr.getNodeFromDB(bn.EncodedChildren[i])
			if err != nil {
				return nil, err
			}
			bn.children[i] = child
		}
	}
	return bn, nil
}

func (tr *patriciaMerkleTree) resolveLeafNode(ln *leafNode) (*leafNode, error) {
	node, err := tr.getNodeFromDB(ln.Value)
	if err != nil {
		return nil, err
	}
	if node, ok := node.(*leafNode); ok {
		ln = node
	}
	return ln, nil
}

func (tr *patriciaMerkleTree) getNodeFromDB(n []byte) (node, error) {
	encChild, err := tr.dbw.Get(n)
	if err != nil {
		return nil, err
	}
	node, err := tr.decodeNode(encChild)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (tr *patriciaMerkleTree) isCollapsed(n node) bool {
	switch n := n.(type) {
	case *extensionNode:
		return n.isCollapsed()
	case *branchNode:
		for i := range n.children {
			if n.children[i] == nil && n.EncodedChildren[i] != nil {
				return true
			}
		}
		return false
	case *leafNode:
		_, err := tr.decodeNode(n.Value)
		if err != nil {
			return false
		}
		return true
	default:
		return false
	}
}

func (tr *patriciaMerkleTree) Update(key, value []byte) error {
	hexKey := keyBytesToHex(key)
	if len(value) != 0 {
		_, newRoot, err := tr.insert(tr.root, hexKey, value)
		if err != nil {
			return err
		}
		tr.root = newRoot
	} else {
		_, newRoot, err := tr.delete(tr.root, hexKey)
		if err != nil {
			return err
		}
		tr.root = newRoot
	}
	return nil
}

func (tr *patriciaMerkleTree) insert(n node, key []byte, value []byte) (bool, node, error) {
	var err error
	switch n := n.(type) {
	case *extensionNode:
		keyMatchLen := prefixLen(key, n.Key)

		if n.isCollapsed() {
			n, err = tr.resolveExtNode(n)
			if err != nil {
				return false, nil, err
			}
		}

		// If the whole key matches, keep this extension node as is
		// and only update the value.
		if keyMatchLen == len(n.Key) {
			dirty, newNode, err := tr.insert(n.child, key[keyMatchLen:], value)
			if !dirty || err != nil {
				return false, n, err
			}
			return true, &extensionNode{n.Key, nil, newNode, nil, true}, nil
		}
		// Otherwise branch out at the index where they differ.
		branch := &branchNode{}
		branch.dirty = true
		oldChildPos := n.Key[keyMatchLen]
		newChildPos := key[keyMatchLen]

		branch.children[oldChildPos] = &extensionNode{n.Key[keyMatchLen+1:], nil, n.child, nil, true}
		_, branch.children[newChildPos], err = tr.insert(nil, key[keyMatchLen+1:], value)
		if err != nil {
			return false, nil, err
		}
		// Replace this extension node with the branch if it occurs at index 0.
		if keyMatchLen == 0 {
			return true, branch, nil
		}
		// Otherwise, replace it with an extension node leading up to the branch.
		return true, &extensionNode{key[:keyMatchLen], nil, branch, nil, true}, nil
	case *branchNode:
		childPos := key[firstByte]
		key, err = removeFirstByte(key)

		if n.isCollapsed(childPos) {
			n, err = tr.resolveBrNode(n)
			if err != nil {
				return false, nil, err
			}
		}

		if err != nil {
			return false, nil, err
		}
		dirty, newNode, err := tr.insert(n.children[childPos], key, value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.children[childPos] = newNode
		return true, n, nil
	case *leafNode:
		if n.isCollapsed(tr) {
			n, err = tr.resolveLeafNode(n)
			if err != nil {
				return false, nil, err
			}
		}
		if bytes.Equal(key, n.Key) {
			n.Value = value
			return true, n, nil
		}

		keyMatchLen := prefixLen(key, n.Key)
		branch := &branchNode{}
		branch.dirty = true
		oldChildPos := n.Key[keyMatchLen]
		newChildPos := key[keyMatchLen]

		_, branch.children[oldChildPos], err = tr.insert(nil, n.Key[keyMatchLen+1:], n.Value)
		if err != nil {
			return false, nil, err
		}

		_, branch.children[newChildPos], err = tr.insert(nil, key[keyMatchLen+1:], value)
		if err != nil {
			return false, nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0.
		if keyMatchLen == 0 {
			return true, branch, nil
		}
		// Otherwise, replace it with a short node leading up to the branch.
		return true, &extensionNode{key[:keyMatchLen], nil, branch, nil, true}, nil
	case nil:
		return true, &leafNode{key, value, nil, true}, nil
	default:
		return false, nil, ErrInvalidNode
	}
}

// Delete removes any existing value for key from the trie.
func (tr *patriciaMerkleTree) Delete(key []byte) error {
	k := keyBytesToHex(key)
	_, newRoot, err := tr.delete(tr.root, k)
	if err != nil {
		return err
	}
	tr.root = newRoot
	return nil
}

func (tr *patriciaMerkleTree) delete(n node, key []byte) (bool, node, error) {
	var err error
	switch n := n.(type) {
	case *extensionNode:
		keyMatchLen := prefixLen(key, n.Key)
		if keyMatchLen < len(n.Key) {
			return false, n, nil // don't replace n on mismatch
		}
		if n.isCollapsed() {
			n, err = tr.resolveExtNode(n)
			if err != nil {
				return false, nil, err
			}
		}

		dirty, newNode, err := tr.delete(n.child, key[len(n.Key):])
		if !dirty || err != nil {
			return false, n, err
		}
		switch newNode := newNode.(type) {
		case *leafNode:
			return true, &leafNode{concat(n.Key, newNode.Key...), newNode.Value, nil, true}, nil
		case *extensionNode:
			return true, &extensionNode{concat(n.Key, newNode.Key...), nil, newNode.child, nil, true}, nil
		default:
			return true, &extensionNode{n.Key, nil, newNode, nil, true}, nil
		}
	case *branchNode:
		childPos := key[firstByte]
		key, err := removeFirstByte(key)
		if err != nil {
			return false, nil, err
		}
		if n.isCollapsed(childPos) {
			n, err = tr.resolveBrNode(n)
			if err != nil {
				return false, nil, err
			}
		}
		dirty, newNode, err := tr.delete(n.children[childPos], key)
		if !dirty || err != nil {
			return false, n, err
		}

		n = n.copy()
		n.hash = nil
		n.children[childPos] = newNode
		if newNode == nil {
			n.EncodedChildren[childPos] = nil
		}

		pos := getChildPosition(n)

		if pos >= 0 {
			if pos != 16 {
				dirty, newNode, err := reduceNode(n, pos)
				if !dirty || err != nil {
					return false, n, err
				}
				return true, newNode, nil
			}
			child := n.children[pos]
			if child, ok := child.(*leafNode); ok {
				return true, &leafNode{[]byte{byte(pos)}, child.Value, nil, true}, nil
			}

		}
		return true, n, nil
	case *leafNode:
		keyMatchLen := prefixLen(key, n.Key)
		if n.isCollapsed(tr) {
			n, err = tr.resolveLeafNode(n)
			if err != nil {
				return false, nil, err
			}
		}
		if keyMatchLen == len(key) {
			return true, nil, nil
		}
		return false, n, nil
	case nil:
		return false, nil, nil
	default:
		return false, nil, ErrInvalidNode
	}

}

func getChildPosition(n *branchNode) int {
	pos := -1
	for i, cld := range &n.children {
		if cld != nil {
			if pos == -1 {
				pos = i
			} else {
				return moreThanOneChildren
			}
		}
	}
	return pos
}

func reduceNode(n *branchNode, pos int) (bool, node, error) {
	switch child := n.children[pos].(type) {
	case *extensionNode:
		k := append([]byte{byte(pos)}, child.Key...)
		return true, &extensionNode{k, nil, child.child, nil, true}, nil
	case *branchNode:
		return true, &extensionNode{[]byte{byte(pos)}, nil, n.children[pos], nil, true}, nil
	case *leafNode:
		k := append([]byte{byte(pos)}, child.Key...)
		return true, &leafNode{k, child.Value, nil, true}, nil
	default:
		return false, nil, ErrInvalidNode
	}

}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}

// Root returns the hash of the root node
func (tr *patriciaMerkleTree) Root() ([]byte, error) {
	h, err := tr.setHash(tr.root)
	if err != nil {
		return nil, err
	}
	return h.getHash(), nil
}

// Prove returns the Merkle proof for the given key
func (tr *patriciaMerkleTree) Prove(key []byte) ([][]byte, error) {
	it := newNodeIterator(tr)

	ok, err := it.Next()
	if err != nil {
		return nil, err
	}
	for ok {
		if it.Leaf() {
			leafKey, err := it.LeafKey()
			if err != nil {
				return nil, err
			}
			if bytes.Equal(leafKey, key) {
				return it.LeafProof()
			}
		}
		ok, err = it.Next()
		if err != nil {
			return nil, err
		}
	}
	return nil, ErrProve
}

// VerifyProof checks Merkle proofs. The given proof must contain the value for
// key in a trie with the given root hash.
func (tr *patriciaMerkleTree) VerifyProof(proofs [][]byte, key []byte) (bool, error) {
	key = keyBytesToHex(key)
	for i := range proofs {
		encNode := proofs[i]
		if encNode == nil {
			return false, nil
		}

		n, err := tr.decodeNode(encNode)
		if err != nil {
			return false, err
		}

		switch n := n.(type) {
		case nil:
			return false, nil
		case *extensionNode:
			key = key[len(n.Key):]
		case *branchNode:
			key = key[1:]
		case *leafNode:
			if bytes.Equal(key, n.Key) {
				return true, nil
			}
			return false, nil
		}
	}
	return false, nil
}

func (tr *patriciaMerkleTree) Commit() error {
	if tr.isCollapsed(tr.root) {
		return nil
	}
	n, err := tr.setHash(tr.root)
	if err != nil {
		return err
	}
	err = tr.commit(n)
	if err != nil {
		return err
	}
	tr.root, err = tr.collapseNode(tr.root)
	if err != nil {
		return err
	}
	return nil
}

func (tr *patriciaMerkleTree) commit(n node) error {
	switch n := n.(type) {
	case *extensionNode:
		err := tr.commit(n.child)
		if err != nil {
			return err
		}
		if !n.dirty {
			return nil
		}
		key := n.getHash()
		collapsedNode, err := tr.collapseNode(n)
		if err != nil {
			return err
		}
		val, err := tr.encodeNode(collapsedNode)
		if err != nil {
			return err
		}
		err = tr.dbw.Put(key, val)
		if err != nil {
			return err
		}
	case *branchNode:
		for i := range n.children {
			if n.children[i] != nil {
				err := tr.commit(n.children[i])
				if err != nil {
					return err
				}
			}
		}
		if !n.dirty {
			return nil
		}
		key := n.getHash()
		collapsedNode, err := tr.collapseNode(n)
		if err != nil {
			return err
		}
		val, err := tr.encodeNode(collapsedNode)
		if err != nil {
			return err
		}
		err = tr.dbw.Put(key, val)
		if err != nil {
			return err
		}
	case *leafNode:
		if !n.dirty {
			return nil
		}
		key := n.getHash()
		val, err := tr.encodeNode(n)
		if err != nil {
			return err
		}
		err = tr.dbw.Put(key, val)
		if err != nil {
			return err
		}
	default:
		return ErrInvalidNode
	}
	return nil

}

func (tr *patriciaMerkleTree) collapseNode(n node) (node, error) {
	var err error
	if n.getHash() == nil {
		tr.root, err = tr.setHash(n)
		if err != nil {
			return n, err
		}
	}
	switch n := n.(type) {
	case *leafNode:
		return n, nil
	case *extensionNode:
		collapsed := n.copy()
		collapsed.EncodedChild = n.child.getHash()
		collapsed.child = nil
		return collapsed, nil
	case *branchNode:
		collapsed := n.copy()
		for i := range n.EncodedChildren {
			if n.children[i] != nil {
				collapsed.EncodedChildren[i] = n.children[i].getHash()
				collapsed.children[i] = nil
			}
		}
		return collapsed, nil
	default:
		return nil, ErrInvalidNode
	}
}

// setHash collapses a node down into a hash node, also returning a copy of the
// original node initialized with the computed hash to replace the original one.
func (tr *patriciaMerkleTree) setHash(n node) (node, error) {
	node, err := tr.hashChildren(n)
	if err != nil {
		return n, err
	}
	n = node

	hashed, err := tr.hash(n)
	if err != nil {
		return n, err
	}

	switch n := n.(type) {
	case *leafNode:
		n.hash = hashed
	case *extensionNode:
		n.hash = hashed
	case *branchNode:
		n.hash = hashed
	}

	return n, nil
}

func (tr *patriciaMerkleTree) hashChildren(n node) (node, error) {
	switch n := n.(type) {
	case *extensionNode:
		node, err := tr.setHash(n.child)
		if err != nil {
			return n, err
		}
		n.child = node
		return n, nil
	case *branchNode:
		for i := 0; i < nrOfChildren; i++ {
			if n.children[i] != nil {
				child, err := tr.setHash(n.children[i])
				if err != nil {
					return n, err
				}
				n.children[i] = child
			}
		}
		return n, nil
	}

	return n, nil
}

// hash collapses any node into a hash
func (tr *patriciaMerkleTree) hash(n node) ([]byte, error) {
	switch n := n.(type) {
	case *extensionNode:
		child, err := tr.encodeNode(n.child)
		if err != nil {
			return nil, err
		}
		n.EncodedChild = child
	case *branchNode:
		for i := range n.EncodedChildren {
			if n.children[i] != nil {
				child, err := tr.encodeNode(n.children[i])
				if err != nil {
					return nil, err
				}
				n.EncodedChildren[i] = child
			}
		}
	}
	tmp, err := tr.encodeNode(n)
	if err != nil {
		return nil, err
	}
	hash := n.getHash()
	if hash == nil {
		hash = tr.hasher.Compute(string(tmp))
	}
	return hash, nil
}

func (tr *patriciaMerkleTree) encodeNode(n node) ([]byte, error) {
	marshaledNode, err := tr.marshalizer.Marshal(n)
	if err != nil {
		return nil, err
	}
	switch n.(type) {
	case *extensionNode:
		marshaledNode = append(marshaledNode, extension)
	case *leafNode:
		marshaledNode = append(marshaledNode, leaf)
	case *branchNode:
		marshaledNode = append(marshaledNode, branch)
	}
	return marshaledNode, nil
}

func (tr *patriciaMerkleTree) decodeNode(encNode []byte) (node, error) {
	if encNode == nil || len(encNode) < 1 {
		return nil, ErrInvalidEncoding
	}
	nodeType := encNode[len(encNode)-1]
	encNode = encNode[:len(encNode)-1]
	var decNode node

	switch nodeType {
	case extension:
		decNode = &extensionNode{}
	case leaf:
		decNode = &leafNode{}
	case branch:
		decNode = &branchNode{}
	default:
		return nil, ErrInvalidNode
	}

	err := tr.marshalizer.Unmarshal(decNode, encNode)
	if err != nil {
		return nil, err
	}
	return decNode, nil

}

func removeFirstByte(val []byte) ([]byte, error) {
	if len(val) > 0 {
		return val[1:], nil
	}
	return nil, ErrValueTooShort
}
