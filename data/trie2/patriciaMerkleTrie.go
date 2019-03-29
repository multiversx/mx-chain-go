package trie2

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

type patriciaMerkleTrie struct {
	root        node
	db          DBWriteCacher
	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(db DBWriteCacher, msh marshal.Marshalizer, hsh hashing.Hasher) (*patriciaMerkleTrie, error) {
	if db == nil {
		return nil, ErrNilDatabase
	}
	if msh == nil {
		return nil, ErrNilMarshalizer
	}
	if hsh == nil {
		return nil, ErrNilHasher
	}
	return &patriciaMerkleTrie{db: db, marshalizer: msh, hasher: hsh}, nil
}

// NewNodeIterator returns a new node iterator for the current trie
func (tr *patriciaMerkleTrie) NewNodeIterator() NodeIterator {
	return newNodeIterator(tr)
}

// Get starts at the root and searches for the given key.
// If the key is present in the tree, it returns the corresponding value
func (tr *patriciaMerkleTrie) Get(key []byte) ([]byte, error) {
	hexKey := keyBytesToHex(key)
	value, err := tr.root.tryGet(hexKey, tr.db, tr.marshalizer)
	return value, err
}

// Update updates the value at the given key.
// If the key is not in the trie, it will be added.
// If the value is empty, the key will be removed from the trie
func (tr *patriciaMerkleTrie) Update(key, value []byte) error {
	hexKey := keyBytesToHex(key)
	node := newLeafNode(hexKey, value)
	if len(value) != 0 {
		if tr.root == nil {
			tr.root = newLeafNode(hexKey, value)
			return nil
		}
		_, newRoot, err := tr.root.insert(node, tr.db, tr.marshalizer)
		if err != nil {
			return err
		}
		tr.root = newRoot
	} else {
		if tr.root == nil {
			return nil
		}
		_, newRoot, err := tr.root.delete(hexKey, tr.db, tr.marshalizer)
		if err != nil {
			return err
		}
		tr.root = newRoot
	}
	return nil
}

// Delete removes the node that has the given key from the tree
func (tr *patriciaMerkleTrie) Delete(key []byte) error {
	hexKey := keyBytesToHex(key)
	if tr.root == nil {
		return nil
	}
	_, newRoot, err := tr.root.delete(hexKey, tr.db, tr.marshalizer)
	if err != nil {
		return err
	}
	tr.root = newRoot
	return nil
}

// Root returns the hash of the root node
func (tr *patriciaMerkleTrie) Root() ([]byte, error) {
	hash := tr.root.getHash()
	if !tr.root.isDirty() && hash != nil {
		return hash, nil
	}
	err := tr.root.setHash(tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, err
	}
	return tr.root.getHash(), nil
}

// Prove returns the Merkle proof for the given key
func (tr *patriciaMerkleTrie) Prove(key []byte) ([][]byte, error) {
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

// VerifyProof checks Merkle proofs.
func (tr *patriciaMerkleTrie) VerifyProof(proofs [][]byte, key []byte) (bool, error) {
	key = keyBytesToHex(key)
	for i := range proofs {
		encNode := proofs[i]
		if encNode == nil {
			return false, nil
		}

		n, err := decodeNode(encNode, tr.marshalizer)
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

// Commit adds all the dirty nodes to the database
func (tr *patriciaMerkleTrie) Commit() error {
	if tr.root.isCollapsed() {
		return nil
	}
	err := tr.root.setHash(tr.marshalizer, tr.hasher)
	if err != nil {
		return err
	}
	err = tr.root.commit(tr.db, tr.marshalizer, tr.hasher)
	if err != nil {
		return err
	}
	newRoot, err := tr.root.getCollapsed(tr.marshalizer, tr.hasher)
	if err != nil {
		return err
	}
	tr.root = newRoot
	return nil
}
