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

type patriciaMerkleTree struct {
	root node

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
	key = keyBytesToHex(key)
	value, newroot, didResolve, err := tr.tryGet(tr.root, key, 0)
	if err == nil && didResolve {
		tr.root = newroot
	}
	return value, err
}

// NodeIterator returns a new node iterator for the current trie
func (tr *patriciaMerkleTree) NodeIterator() trie3.NodeIterator {
	return newNodeIterator(tr)
}

func (tr *patriciaMerkleTree) Update(key, value []byte) error {
	k := keyBytesToHex(key)
	if len(value) != 0 {
		_, n, err := tr.insert(tr.root, nil, k, value)
		if err != nil {
			return err
		}
		tr.root = n
	} else {
		_, n, err := tr.delete(tr.root, nil, k)
		if err != nil {
			return err
		}
		tr.root = n
	}
	return nil
}

func (tr *patriciaMerkleTree) insert(n node, prefix, key []byte, value []byte) (bool, node, error) {
	switch n := n.(type) {
	case *extensionNode:
		matchlen := prefixLen(key, n.Key)
		// If the whole key matches, keep this short node as is
		// and only update the value.
		if matchlen == len(n.Key) {
			dirty, nn, err := tr.insert(n.nextNode, append(prefix, key[:matchlen]...), key[matchlen:], value)
			if !dirty || err != nil {
				return false, n, err
			}
			return true, &extensionNode{n.Key, nil, nn, nil, dirty}, nil
		}
		// Otherwise branch out at the index where they differ.
		branch := &branchNode{}
		var err error

		branch.childrenNodes[n.Key[matchlen]] = &extensionNode{n.Key[matchlen+1:], nil, n.nextNode, nil, false}
		if err != nil {
			return false, nil, err
		}

		_, branch.childrenNodes[key[matchlen]], err = tr.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
		if err != nil {
			return false, nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0.
		if matchlen == 0 {
			return true, branch, nil
		}
		// Otherwise, replace it with a short node leading up to the branch.
		return true, &extensionNode{key[:matchlen], nil, branch, nil, true}, nil
	case *branchNode:
		dirty, nn, err := tr.insert(n.childrenNodes[key[0]], append(prefix, key[0]), key[1:], value)
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.childrenNodes[key[0]] = nn
		return true, n, nil
	case *leafNode:
		if len(key) == 0 {
			n.Value = value
			return true, n, nil
		}
		matchlen := prefixLen(key, n.Key)
		branch := &branchNode{}
		var err error

		_, branch.childrenNodes[n.Key[matchlen]], err = tr.insert(nil, append(prefix, n.Key[:matchlen+1]...), n.Key[matchlen+1:], n.Value)
		if err != nil {
			return false, nil, err
		}

		_, branch.childrenNodes[key[matchlen]], err = tr.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
		if err != nil {
			return false, nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0.
		if matchlen == 0 {
			return true, branch, nil
		}
		// Otherwise, replace it with a short node leading up to the branch.
		return true, &extensionNode{key[:matchlen], nil, branch, nil, true}, nil

	case nil:
		return true, &leafNode{key, value, nil, true}, nil

	default:
		return false, nil, trie3.ErrInvalidNode
	}
}

// Delete removes any existing value for key from the trie.
func (tr *patriciaMerkleTree) Delete(key []byte) error {
	k := keyBytesToHex(key)
	_, n, err := tr.delete(tr.root, nil, k)
	if err != nil {
		return err
	}
	tr.root = n
	return nil
}

func (tr *patriciaMerkleTree) delete(n node, prefix, key []byte) (bool, node, error) {
	switch n := n.(type) {
	case *leafNode:
		matchlen := prefixLen(key, n.Key)
		if matchlen == len(key) {
			return true, nil, nil // remove n entirely for whole matches
		}
		return false, n, nil
	case *extensionNode:
		matchlen := prefixLen(key, n.Key)
		if matchlen < len(n.Key) {
			return false, n, nil // don't replace n on mismatch
		}
		// The key is longer than n.Key. Remove the remaining suffix
		// from the subtrie. Child can never be nil here since the
		// subtrie must contain at least two other values with keys
		// longer than n.Key.
		dirty, child, err := tr.delete(n.nextNode, append(prefix, key[:len(n.Key)]...), key[len(n.Key):])
		if !dirty || err != nil {
			return false, n, err
		}
		switch child := child.(type) {
		case *leafNode:
			// Deleting from the subtrie reduced it to another
			// short node. Merge the nodes to avoid creating a
			// shortNode{..., shortNode{...}}. Use concat (which
			// always creates a new slice) instead of append to
			// avoid modifying n.Key since it might be shared with
			// other nodes.
			return true, &leafNode{concat(n.Key, child.Key...), child.Value, nil, true}, nil
		case *extensionNode:
			return true, &extensionNode{concat(n.Key, child.Key...), nil, child.nextNode, nil, true}, nil
		default:
			return true, &extensionNode{n.Key, nil, child, nil, true}, nil
		}
	case *branchNode:
		dirty, nn, err := tr.delete(n.childrenNodes[key[0]], append(prefix, key[0]), key[1:])
		if !dirty || err != nil {
			return false, n, err
		}
		n = n.copy()
		n.hash = nil
		n.childrenNodes[key[0]] = nn
		if nn == nil {
			n.Children[key[0]] = nil
		}

		// Check how many non-nil entries are left after deleting and
		// reduce the full node to a short node if only one entry is
		// left. Since n must've contained at least two children
		// before deletion (otherwise it would not be a full node) n
		// can never be reduced to nil.
		//
		// When the loop is done, pos contains the index of the single
		// value that is left in n or -2 if n contains at least two
		// values.
		pos := -1
		for i, cld := range &n.childrenNodes {
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
				// If the remaining entry is a short node, it replaces
				// n and its key gets the missing nibble tacked to the
				// front. This avoids creating an invalid
				// shortNode{..., shortNode{...}}.
				cnode := n.childrenNodes[pos]

				if cnode, ok := cnode.(*leafNode); ok {
					k := append([]byte{byte(pos)}, cnode.Key...)
					return true, &leafNode{k, cnode.Value, nil, true}, nil
				}
				if cnode, ok := cnode.(*extensionNode); ok {
					k := append([]byte{byte(pos)}, cnode.Key...)
					return true, &extensionNode{k, nil, cnode.nextNode, nil, true}, nil
				}
			}
			// Otherwise, n is replaced by a one-nibble short node
			// containing the child.
			cnode := n.childrenNodes[pos]

			if cnode, ok := cnode.(*leafNode); ok {
				return true, &leafNode{[]byte{byte(pos)}, cnode.Value, nil, true}, nil
			}
			if _, ok := cnode.(*branchNode); ok {
				return true, &extensionNode{[]byte{byte(pos)}, nil, n.childrenNodes[pos], nil, true}, nil
			}

		}

		// n still contains at least two values and cannot be reduced.
		return true, n, nil
	case nil:
		return false, nil, nil
	default:
		return false, nil, trie3.ErrInvalidNode
	}

}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}

func (tr *patriciaMerkleTree) tryGet(origNode node, key []byte, pos int) (value []byte, newnode node, didResolve bool, err error) {
	switch n := (origNode).(type) {
	case nil:
		return nil, nil, false, nil
	case *leafNode:
		return n.Value, n, false, nil
	case *extensionNode:

		isSmallLen := len(key)-pos < len(n.Key)
		keysDontMatch := !bytes.Equal(n.Key, key[pos:pos+len(n.Key)])

		if isSmallLen || keysDontMatch {
			// key not found in trie
			return nil, n, false, nil
		}
		value, newnode, didResolve, err = tr.tryGet(n.nextNode, key, pos+len(n.Key))
		if err == nil && didResolve {
			n = n.copy()
			n.nextNode = newnode
		}
		return value, n, didResolve, err
	case *branchNode:
		value, newnode, didResolve, err = tr.tryGet(n.childrenNodes[key[pos]], key, pos+1)
		if err == nil && didResolve {
			n = n.copy()
			n.childrenNodes[key[pos]] = newnode
		}
		return value, n, didResolve, err
	default:
		return nil, nil, false, trie2.ErrInvalidNode
	}
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
		collapsed.Next = n.nextNode.getHash()
		collapsed.nextNode = nil
		return collapsed, nil
	case *branchNode:
		collapsed := n.copy()
		for i := range n.Children {
			if n.childrenNodes[i] != nil {
				collapsed.Children[i] = n.childrenNodes[i].getHash()
				collapsed.childrenNodes[i] = nil
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
		node, err := tr.setHash(n.nextNode)
		if err != nil {
			return n, err
		}
		n.nextNode = node
		return n, nil
	case *branchNode:
		for i := 0; i < nrOfChildren; i++ {
			if n.childrenNodes[i] != nil {
				child, err := tr.setHash(n.childrenNodes[i])
				if err != nil {
					return n, err
				}
				n.childrenNodes[i] = child
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
		child, err := tr.encodeNode(n.nextNode)
		if err != nil {
			return nil, err
		}
		n.Next = child
	case *branchNode:
		for i := range n.Children {
			if n.childrenNodes[i] != nil {
				child, err := tr.encodeNode(n.childrenNodes[i])
				if err != nil {
					return nil, err
				}
				n.Children[i] = child
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
