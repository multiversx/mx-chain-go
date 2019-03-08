package patriciaMerkleTrie

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie3"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

const (
	extension = iota
	leaf
	branch
)
const firstByte = 0

type patriciaMerkleTree struct {
	root        node
	dbw         trie3.DBWriteCacher
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

func NewTrie(hsh hashing.Hasher, msh marshal.Marshalizer, dbw trie3.DBWriteCacher) (*patriciaMerkleTree, error) {

	if hsh == nil {
		return nil, trie3.ErrNilHasher
	}

	if msh == nil {
		return nil, trie3.ErrNilMarshalizer
	}

	return &patriciaMerkleTree{dbw: dbw, hasher: hsh, marshalizer: msh}, nil
}

func (tr *patriciaMerkleTree) Get(key []byte) ([]byte, error) {
	hexKey := keyBytesToHex(key)
	value, err := tryGet(tr.root, hexKey)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func tryGet(n node, key []byte) (value []byte, err error) {
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
		value, err = tryGet(n.child, key)
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

		value, err = tryGet(n.children[childPos], key)
		return value, err
	case *leafNode:
		if bytes.Equal(key, n.Key) {
			return n.Value, nil
		}
		return nil, nil
	case nil:
		return nil, nil
	default:
		return nil, trie2.ErrInvalidNode
	}
}

func (tr *patriciaMerkleTree) Update(key, value []byte) error {
	hexKey := keyBytesToHex(key)
	if len(value) != 0 {
		newRoot, err := insert(tr.root, hexKey, value)
		if err != nil {
			return err
		}
		tr.root = newRoot
	} else {
		newRoot, err := tr.delete(tr.root, hexKey)
		if err != nil {
			return err
		}
		tr.root = newRoot
	}
	return nil
}

func insert(n node, key []byte, value []byte) (node, error) {
	var err error
	switch n := n.(type) {
	case *extensionNode:
		keyMatchLen := prefixLen(key, n.Key)
		// If the whole key matches, keep this extension node as is
		// and only update the value.
		if keyMatchLen == len(n.Key) {
			newNode, err := insert(n.child, key[keyMatchLen:], value)
			if err != nil {
				return n, err
			}
			return &extensionNode{n.Key, nil, newNode, nil, true}, nil
		}
		// Otherwise branch out at the index where they differ.
		branch := &branchNode{}
		oldChildPos := n.Key[keyMatchLen]
		newChildPos := key[keyMatchLen]

		branch.children[oldChildPos] = &extensionNode{n.Key[keyMatchLen+1:], nil, n.child, nil, true}
		branch.children[newChildPos], err = insert(nil, key[keyMatchLen+1:], value)
		if err != nil {
			return nil, err
		}
		// Replace this extension node with the branch if it occurs at index 0.
		if keyMatchLen == 0 {
			return branch, nil
		}
		// Otherwise, replace it with an extension node leading up to the branch.
		return &extensionNode{key[:keyMatchLen], nil, branch, nil, true}, nil
	case *branchNode:
		childPos := key[firstByte]
		key, err = removeFirstByte(key)
		if err != nil {
			return nil, err
		}
		newNode, err := insert(n.children[childPos], key, value)
		if err != nil {
			return n, err
		}
		n = n.copy()
		n.children[childPos] = newNode
		return n, nil
	case *leafNode:
		if bytes.Equal(key, n.Key) {
			n.Value = value
			return n, nil
		}

		keyMatchLen := prefixLen(key, n.Key)
		branch := &branchNode{}
		oldChildPos := n.Key[keyMatchLen]
		newChildPos := key[keyMatchLen]

		branch.children[oldChildPos], err = insert(nil, n.Key[keyMatchLen+1:], n.Value)
		if err != nil {
			return nil, err
		}

		branch.children[newChildPos], err = insert(nil, key[keyMatchLen+1:], value)
		if err != nil {
			return nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0.
		if keyMatchLen == 0 {
			return branch, nil
		}
		// Otherwise, replace it with a short node leading up to the branch.
		return &extensionNode{key[:keyMatchLen], nil, branch, nil, true}, nil
	case nil:
		return &leafNode{key, value, nil, true}, nil
	default:
		return nil, trie3.ErrInvalidNode
	}
}

// Delete removes any existing value for key from the trie.
func (tr *patriciaMerkleTree) Delete(key []byte) error {
	k := keyBytesToHex(key)
	newRoot, err := tr.delete(tr.root, k)
	if err != nil {
		return err
	}
	tr.root = newRoot
	return nil
}

func (tr *patriciaMerkleTree) delete(n node, key []byte) (node, error) {
	switch n := n.(type) {
	case *extensionNode:
		keyMatchLen := prefixLen(key, n.Key)
		if keyMatchLen < len(n.Key) {
			return n, nil // don't replace n on mismatch
		}

		newNode, err := tr.delete(n.child, key[len(n.Key):])
		if err != nil {
			return n, err
		}
		switch newNode := newNode.(type) {
		case *leafNode:
			return &leafNode{concat(n.Key, newNode.Key...), newNode.Value, nil, true}, nil
		case *extensionNode:
			return &extensionNode{concat(n.Key, newNode.Key...), nil, newNode.child, nil, true}, nil
		default:
			return &extensionNode{n.Key, nil, newNode, nil, true}, nil
		}
	case *branchNode:
		childPos := key[firstByte]
		key, err := removeFirstByte(key)
		if err != nil {
			return nil, err
		}
		newNode, err := tr.delete(n.children[childPos], key)
		if err != nil {
			return n, err
		}

		n = n.copy()
		n.hash = nil
		n.children[childPos] = newNode
		if newNode == nil {
			n.EncodedChildren[childPos] = nil
		}

		// When the loop is done, pos contains the index of the single
		// value that is left in n or -2 if n contains at least two
		// values.
		pos := -1
		for i, cld := range &n.children {
			if cld != nil {
				if pos == -1 {
					pos = i
				} else {
					pos = -2
					break
				}
			}
		}
		if pos >= 0 {
			if pos != 16 {
				switch cnode := n.children[pos].(type) {
				case *extensionNode:
					k := append([]byte{byte(pos)}, cnode.Key...)
					return &extensionNode{k, nil, cnode.child, nil, true}, nil
				case *branchNode:
					return &extensionNode{[]byte{byte(pos)}, nil, n.children[pos], nil, true}, nil
				case *leafNode:
					k := append([]byte{byte(pos)}, cnode.Key...)
					return &leafNode{k, cnode.Value, nil, true}, nil
				}
			}
			cnode := n.children[pos]

			if cnode, ok := cnode.(*leafNode); ok {
				return &leafNode{[]byte{byte(pos)}, cnode.Value, nil, true}, nil
			}

		}

		// n still contains at least two values and cannot be reduced.
		return n, nil
	case *leafNode:
		keyMatchLen := prefixLen(key, n.Key)
		if keyMatchLen == len(key) {
			return nil, nil // remove n entirely for whole matches
		}
		return n, nil
	case nil:
		return nil, nil
	default:
		return nil, trie3.ErrInvalidNode
	}

}

// NodeIterator returns a new node iterator for the current trie
func (tr *patriciaMerkleTree) NodeIterator() trie3.NodeIterator {
	return newNodeIterator(tr)
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

	return nil, trie2.ErrProve
}

// VerifyProof checks Merkle proofs. The given proof must contain the value for
// key in a trie with the given root hash.
func (tr *patriciaMerkleTree) VerifyProof(proofs [][]byte, key []byte) (bool, error) {
	key = keyBytesToHex(key)
	for i := range proofs {
		buf := proofs[i]
		if buf == nil {
			return false, nil
		}

		n, err := tr.decodeNode(buf)
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
		return nil, trie3.ErrInvalidNode
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
	tmp, err := tr.marshalizer.Marshal(n)
	if err != nil {
		return nil, err
	}

	switch n.(type) {
	case *extensionNode:
		tmp = append(tmp, extension)
	case *leafNode:
		tmp = append(tmp, leaf)
	case *branchNode:
		tmp = append(tmp, branch)
	}
	return tmp, nil
}

func (tr *patriciaMerkleTree) decodeNode(encNode []byte) (node, error) {
	if encNode == nil || len(encNode) < 1 {
		return nil, trie3.ErrInvalidEncoding
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
		return nil, trie3.ErrInvalidNode
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
	return nil, trie3.ErrValueTooShort
}
