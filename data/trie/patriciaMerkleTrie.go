package trie

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

const (
	extension = iota
	leaf
	branch
)

var emptyTrieHash = make([]byte, 32)

type patriciaMerkleTrie struct {
	root         node
	db           data.DBWriteCacher
	marshalizer  marshal.Marshalizer
	hasher       hashing.Hasher
	mutOperation sync.RWMutex
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(db data.DBWriteCacher, msh marshal.Marshalizer, hsh hashing.Hasher) (*patriciaMerkleTrie, error) {
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

// Get starts at the root and searches for the given key.
// If the key is present in the tree, it returns the corresponding value
func (tr *patriciaMerkleTrie) Get(key []byte) ([]byte, error) {
	tr.mutOperation.RLock()
	defer tr.mutOperation.RUnlock()

	if tr.root == nil {
		return nil, nil
	}
	hexKey := keyBytesToHex(key)

	return tr.root.tryGet(hexKey, tr.db, tr.marshalizer)
}

// Update updates the value at the given key.
// If the key is not in the trie, it will be added.
// If the value is empty, the key will be removed from the trie
func (tr *patriciaMerkleTrie) Update(key, value []byte) error {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

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
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

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
	tr.mutOperation.RLock()
	defer tr.mutOperation.RUnlock()

	if tr.root == nil {
		return emptyTrieHash, nil
	}

	hash := tr.root.getHash()
	if hash != nil {
		return hash, nil
	}
	err := tr.root.setRootHash(tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, err
	}
	return tr.root.getHash(), nil
}

// Prove returns the Merkle proof for the given key
func (tr *patriciaMerkleTrie) Prove(key []byte) ([][]byte, error) {
	tr.mutOperation.RLock()
	defer tr.mutOperation.RUnlock()

	if tr.root == nil {
		return nil, ErrNilNode
	}

	var proof [][]byte
	hexKey := keyBytesToHex(key)
	node := tr.root

	err := node.setRootHash(tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, err
	}

	for {
		encNode, err := node.getEncodedNode(tr.marshalizer)
		if err != nil {
			return nil, err
		}
		proof = append(proof, encNode)

		node, hexKey, err = node.getNext(hexKey, tr.db, tr.marshalizer)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return proof, nil
		}
	}
}

// VerifyProof checks Merkle proofs.
func (tr *patriciaMerkleTrie) VerifyProof(proofs [][]byte, key []byte) (bool, error) {
	tr.mutOperation.RLock()
	defer tr.mutOperation.RUnlock()

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
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return nil
	}
	if tr.root.isCollapsed() {
		return nil
	}
	err := tr.root.setRootHash(tr.marshalizer, tr.hasher)
	if err != nil {
		return err
	}
	err = tr.root.commit(0, tr.db, tr.marshalizer, tr.hasher)
	if err != nil {
		return err
	}
	return nil
}

// Recreate returns a new trie that has the given root hash and database
func (tr *patriciaMerkleTrie) Recreate(root []byte) (data.Trie, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	newTr, err := NewTrie(tr.db, tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, err
	}

	if emptyTrie(root) {
		return newTr, nil
	}

	encRoot, err := tr.db.Get(root)
	if err != nil {
		return nil, err
	}

	newRoot, err := decodeNode(encRoot, tr.marshalizer)
	if err != nil {
		return nil, err
	}

	newTr.root = newRoot
	return newTr, nil
}

// DeepClone returns a new trie with all nodes deeply copied
func (tr *patriciaMerkleTrie) DeepClone() (data.Trie, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	clonedTrie, err := NewTrie(tr.db, tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, err
	}

	if tr.root == nil {
		return clonedTrie, nil
	}

	clonedTrie.root = tr.root.deepClone()

	return clonedTrie, nil
}

// String outputs a graphical view of the trie. Mainly used in tests/debugging
func (tr *patriciaMerkleTrie) String() string {
	writer := bytes.NewBuffer(make([]byte, 0))

	if tr.root == nil {
		_, _ = fmt.Fprintln(writer, "*** EMPTY TRIE ***")
	} else {
		tr.root.print(writer, 0)
	}

	return writer.String()
}

func emptyTrie(root []byte) bool {
	if bytes.Equal(root, make([]byte, 0)) {
		return true
	}
	if bytes.Equal(root, emptyTrieHash) {
		return true
	}

	return false
}
