package trie

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
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

	dbEvictionWaitingList data.DBRemoveCacher
	oldHashes             [][]byte
	oldRoot               []byte
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(
	db data.DBWriteCacher,
	msh marshal.Marshalizer,
	hsh hashing.Hasher,
	evictionDb storage.Persister,
	evictionCacheSize int,
) (*patriciaMerkleTrie, error) {
	if db == nil || db.IsInterfaceNil() {
		return nil, ErrNilDatabase
	}
	if msh == nil || msh.IsInterfaceNil() {
		return nil, ErrNilMarshalizer
	}
	if hsh == nil || hsh.IsInterfaceNil() {
		return nil, ErrNilHasher
	}
	if evictionDb == nil || evictionDb.IsInterfaceNil() {
		return nil, ErrNilDatabase
	}
	if evictionCacheSize < 1 {
		return nil, data.ErrInvalidCacheSize
	}

	evictionWaitList, err := evictionWaitingList.NewEvictionWaitingList(evictionCacheSize, evictionDb, msh)
	if err != nil {
		return nil, err
	}

	return &patriciaMerkleTrie{
		db:                    db,
		dbEvictionWaitingList: evictionWaitList,
		oldHashes:             make([][]byte, 0),
		oldRoot:               make([]byte, 0),
		marshalizer:           msh,
		hasher:                hsh,
	}, nil
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

		if !tr.root.isDirty() {
			tr.oldRoot = tr.root.getHash()
		}

		_, newRoot, oldHashes, err := tr.root.insert(node, tr.db, tr.marshalizer)
		if err != nil {
			return err
		}
		tr.root = newRoot
		tr.oldHashes = append(tr.oldHashes, oldHashes...)
	} else {
		if tr.root == nil {
			return nil
		}

		if !tr.root.isDirty() {
			tr.oldRoot = tr.root.getHash()
		}

		_, newRoot, oldHashes, err := tr.root.delete(hexKey, tr.db, tr.marshalizer)
		if err != nil {
			return err
		}
		tr.root = newRoot
		tr.oldHashes = append(tr.oldHashes, oldHashes...)
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

	if !tr.root.isDirty() {
		tr.oldRoot = tr.root.getHash()
	}

	_, newRoot, oldHashes, err := tr.root.delete(hexKey, tr.db, tr.marshalizer)
	if err != nil {
		return err
	}
	tr.root = newRoot
	tr.oldHashes = append(tr.oldHashes, oldHashes...)

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

	wantHash, err := tr.Root()
	if err != nil {
		return false, err
	}

	key = keyBytesToHex(key)
	for i := range proofs {
		encNode := proofs[i]
		if encNode == nil {
			return false, nil
		}

		hash := tr.hasher.Compute(string(encNode))
		if !bytes.Equal(wantHash, hash) {
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
			wantHash = n.EncodedChild
		case *branchNode:
			wantHash = n.EncodedChildren[key[0]]
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

	if len(tr.oldHashes) > 0 && len(tr.oldRoot) > 0 {
		err := tr.dbEvictionWaitingList.Put(tr.oldRoot, tr.oldHashes)
		if err != nil {
			return err
		}
		tr.oldRoot = make([]byte, 0)
		tr.oldHashes = make([][]byte, 0)
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

	newTr, err := NewTrie(tr.db, tr.marshalizer, tr.hasher, memorydb.New(), tr.dbEvictionWaitingList.GetSize())
	if err != nil {
		return nil, err
	}
	newTr.dbEvictionWaitingList = tr.dbEvictionWaitingList

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

	newRoot.setGivenHash(root)
	newTr.root = newRoot
	return newTr, nil
}

// DeepClone returns a new trie with all nodes deeply copied
func (tr *patriciaMerkleTrie) DeepClone() (data.Trie, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	clonedTrie, err := NewTrie(tr.db, tr.marshalizer, tr.hasher, memorydb.New(), tr.dbEvictionWaitingList.GetSize())
	if err != nil {
		return nil, err
	}
	clonedTrie.dbEvictionWaitingList = tr.dbEvictionWaitingList

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

// IsInterfaceNil returns true if there is no value under the interface
func (tr *patriciaMerkleTrie) IsInterfaceNil() bool {
	if tr == nil {
		return true
	}
	return false
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

// Rollback invalidates the hashes that correspond to the given root hash from the eviction waiting list
func (tr *patriciaMerkleTrie) Rollback(rootHash []byte) error {
	_, err := tr.dbEvictionWaitingList.Evict(rootHash)
	return err
}

// Prune removes from the database all the old hashes that correspond to the given root hash
func (tr *patriciaMerkleTrie) Prune(rootHash []byte) error {
	hashes, err := tr.dbEvictionWaitingList.Evict(rootHash)
	if err != nil {
		return err
	}

	for i := range hashes {
		err := tr.db.Remove(hashes[i])
		if err != nil {
			return err
		}
	}

	return nil
}
