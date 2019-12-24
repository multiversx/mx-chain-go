package trie

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("trie")

const (
	extension = iota
	leaf
	branch
)

const maxSnapshots = 2

var emptyTrieHash = make([]byte, 32)

type patriciaMerkleTrie struct {
	root node

	trieStorage  data.StorageManager
	marshalizer  marshal.Marshalizer
	hasher       hashing.Hasher
	mutOperation sync.RWMutex

	oldHashes [][]byte
	oldRoot   []byte
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(
	trieStorage data.StorageManager,
	msh marshal.Marshalizer,
	hsh hashing.Hasher,
) (*patriciaMerkleTrie, error) {
	if check.IfNil(trieStorage) {
		return nil, ErrNilTrieStorage
	}
	if check.IfNil(msh) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hsh) {
		return nil, ErrNilHasher
	}

	return &patriciaMerkleTrie{
		trieStorage: trieStorage,
		marshalizer: msh,
		hasher:      hsh,
		oldHashes:   make([][]byte, 0),
		oldRoot:     make([]byte, 0),
	}, nil
}

// Get starts at the root and searches for the given key.
// If the key is present in the tree, it returns the corresponding value
func (tr *patriciaMerkleTrie) Get(key []byte) ([]byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return nil, nil
	}
	hexKey := keyBytesToHex(key)

	val, err := tr.root.tryGet(hexKey, tr.trieStorage.Database())
	if err != nil {
		log.Debug("trie get", "error", key)
	}

	return val, err
}

// Update updates the value at the given key.
// If the key is not in the trie, it will be added.
// If the value is empty, the key will be removed from the trie
func (tr *patriciaMerkleTrie) Update(key, value []byte) error {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	hexKey := keyBytesToHex(key)
	newLn, err := newLeafNode(hexKey, value, tr.marshalizer, tr.hasher)
	if err != nil {
		return err
	}

	var newRoot node
	var oldHashes [][]byte
	if len(value) != 0 {
		if tr.root == nil {
			newRoot, err = newLeafNode(hexKey, value, tr.marshalizer, tr.hasher)
			if err != nil {
				return err
			}

			tr.root = newRoot
			return nil
		}

		if !tr.root.isDirty() {
			tr.oldRoot = tr.root.getHash()
		}

		_, newRoot, oldHashes, err = tr.root.insert(newLn, tr.trieStorage.Database())
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

		_, newRoot, oldHashes, err = tr.root.delete(hexKey, tr.trieStorage.Database())
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

	_, newRoot, oldHashes, err := tr.root.delete(hexKey, tr.trieStorage.Database())
	if err != nil {
		return err
	}
	tr.root = newRoot
	tr.oldHashes = append(tr.oldHashes, oldHashes...)

	return nil
}

// Root returns the hash of the root node
func (tr *patriciaMerkleTrie) Root() ([]byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return emptyTrieHash, nil
	}

	hash := tr.root.getHash()
	if hash != nil {
		return hash, nil
	}
	err := tr.root.setRootHash()
	if err != nil {
		return nil, err
	}
	return tr.root.getHash(), nil
}

// Prove returns the Merkle proof for the given key
func (tr *patriciaMerkleTrie) Prove(key []byte) ([][]byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return nil, ErrNilNode
	}

	var proof [][]byte
	hexKey := keyBytesToHex(key)
	n := tr.root

	err := n.setRootHash()
	if err != nil {
		return nil, err
	}

	var encNode []byte
	for {
		encNode, err = n.getEncodedNode()
		if err != nil {
			return nil, err
		}
		proof = append(proof, encNode)

		n, hexKey, err = n.getNext(hexKey, tr.trieStorage.Database())
		if err != nil {
			return nil, err
		}
		if n == nil {
			return proof, nil
		}
	}
}

// VerifyProof checks Merkle proofs.
func (tr *patriciaMerkleTrie) VerifyProof(proofs [][]byte, key []byte) (bool, error) {
	wantHash, err := tr.Root()
	if err != nil {
		return false, err
	}

	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

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

		var n node
		n, err = decodeNode(encNode, tr.marshalizer, tr.hasher)
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
	err := tr.root.setRootHash()
	if err != nil {
		return err
	}

	newRoot := tr.root.getHash()
	newHashes, err := tr.root.getDirtyHashes()
	if err != nil {
		return err
	}

	if len(newHashes) > 0 && len(newRoot) > 0 {
		newRoot = append(newRoot, byte(data.NewRoot))
		err = tr.trieStorage.MarkForEviction(newRoot, newHashes)
		if err != nil {
			return err
		}
	}

	if len(tr.oldHashes) > 0 && len(tr.oldRoot) > 0 {
		tr.oldRoot = append(tr.oldRoot, byte(data.OldRoot))
		err = tr.trieStorage.MarkForEviction(tr.oldRoot, tr.oldHashes)
		if err != nil {
			return err
		}
		tr.oldRoot = make([]byte, 0)
		tr.oldHashes = make([][]byte, 0)
	}

	err = tr.root.commit(false, 0, tr.trieStorage.Database(), tr.trieStorage.Database())
	if err != nil {
		return err
	}
	return nil
}

// Recreate returns a new trie that has the given root hash and database
func (tr *patriciaMerkleTrie) Recreate(root []byte) (data.Trie, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if emptyTrie(root) {
		clonedTrieStorage := tr.trieStorage.Clone()
		return NewTrie(
			clonedTrieStorage,
			tr.marshalizer,
			tr.hasher,
		)
	}

	newTr, err := tr.recreateFromDb(root)
	if err != nil {
		log.Debug("trie recreate", "error", root)
	}

	return newTr, err
}

// DeepClone returns a new trie with all nodes deeply copied
func (tr *patriciaMerkleTrie) DeepClone() (data.Trie, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	clonedTrieStorage := tr.trieStorage.Clone()
	clonedTrie, err := NewTrie(
		clonedTrieStorage,
		tr.marshalizer,
		tr.hasher,
	)
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

// IsInterfaceNil returns true if there is no value under the interface
func (tr *patriciaMerkleTrie) IsInterfaceNil() bool {
	return tr == nil
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

// Prune removes from the database all the old hashes that correspond to the given root hash
func (tr *patriciaMerkleTrie) Prune(rootHash []byte, identifier data.TriePruningIdentifier) error {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	rootHash = append(rootHash, byte(identifier))
	log.Trace("trie prune", "root", rootHash)
	return tr.trieStorage.Prune(rootHash)
}

// CancelPrune invalidates the hashes that correspond to the given root hash from the eviction waiting list
func (tr *patriciaMerkleTrie) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	tr.mutOperation.Lock()
	rootHash = append(rootHash, byte(identifier))
	log.Trace("trie cancel prune", "root", rootHash)
	tr.trieStorage.CancelPrune(rootHash)
	tr.mutOperation.Unlock()
}

// AppendToOldHashes appends the given hashes to the trie's oldHashes variable
func (tr *patriciaMerkleTrie) AppendToOldHashes(hashes [][]byte) {
	tr.mutOperation.Lock()
	tr.oldHashes = append(tr.oldHashes, hashes...)
	tr.mutOperation.Unlock()
}

// ResetOldHashes resets the oldHashes and oldRoot variables and returns the old hashes
func (tr *patriciaMerkleTrie) ResetOldHashes() [][]byte {
	tr.mutOperation.Lock()
	oldHashes := tr.oldHashes
	tr.oldHashes = make([][]byte, 0)
	tr.oldRoot = make([]byte, 0)
	tr.mutOperation.Unlock()

	return oldHashes
}

// SetCheckpoint adds the current state of the trie to the snapshot database
func (tr *patriciaMerkleTrie) SetCheckpoint(rootHash []byte) {
	tr.trieStorage.SetCheckpoint(rootHash, tr.marshalizer, tr.hasher)
}

// TakeSnapshot creates a new database in which the current state of the trie is saved.
// If the maximum number of snapshots has been reached, the oldest snapshot is removed.
func (tr *patriciaMerkleTrie) TakeSnapshot(rootHash []byte) {
	tr.trieStorage.TakeSnapshot(rootHash, tr.marshalizer, tr.hasher)
}

// Database returns the trie database
func (tr *patriciaMerkleTrie) Database() data.DBWriteCacher {
	return tr.trieStorage.Database()
}

func (tr *patriciaMerkleTrie) recreateFromDb(rootHash []byte) (data.Trie, error) {
	db := tr.trieStorage.GetDbThatContainsHash(rootHash)
	if db == nil {
		return nil, ErrHashNotFound
	}

	clonedTrieStorage := tr.trieStorage.Clone()
	newTr, err := NewTrie(
		clonedTrieStorage,
		tr.marshalizer,
		tr.hasher,
	)
	if err != nil {
		return nil, err
	}
	newTr.trieStorage.SetDatabase(db)

	newRoot, err := getNodeFromDBAndDecode(rootHash, db, tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, err
	}

	newRoot.setGivenHash(rootHash)
	newTr.root = newRoot
	return newTr, nil
}

// GetSerializedNodes returns a batch of serialized nodes from the trie, starting from the given hash
func (tr *patriciaMerkleTrie) GetSerializedNodes(rootHash []byte, maxBuffToSend uint64) ([][]byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	size := uint64(0)

	newTr, err := tr.recreateFromDb(rootHash)
	if err != nil {
		return nil, err
	}

	it, err := NewIterator(newTr)
	if err != nil {
		return nil, err
	}

	encNode, err := it.MarshalizedNode()
	if err != nil {
		return nil, err
	}

	nodes := make([][]byte, 0)
	nodes = append(nodes, encNode)
	size += uint64(len(encNode))

	for it.HasNext() {
		err = it.Next()
		if err != nil {
			return nil, err
		}

		encNode, err = it.MarshalizedNode()
		if err != nil {
			return nil, err
		}

		if size+uint64(len(encNode)) > maxBuffToSend {
			return nodes, nil
		}
		nodes = append(nodes, encNode)
		size += uint64(len(encNode))
	}

	return nodes, nil
}

// GetAllLeaves iterates the trie and returns a map that contains all leafNodes information
func (tr *patriciaMerkleTrie) GetAllLeaves() (map[string][]byte, error) {
	if tr.root == nil {
		return map[string][]byte{}, nil
	}

	leaves := make(map[string][]byte)
	err := tr.root.getAllLeaves(leaves, []byte{}, tr.Database(), tr.marshalizer)
	if err != nil {
		return nil, err
	}

	return leaves, nil
}
