package trie

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/errors"
)

var log = logger.GetOrCreate("trie")

var _ dataRetriever.TrieDataGetter = (*patriciaMerkleTrie)(nil)

const (
	extension = iota
	leaf
	branch
)

// EmptyTrieHash returns the value with empty trie hash
var EmptyTrieHash = make([]byte, 32)

type patriciaMerkleTrie struct {
	root node

	trieStorage  common.StorageManager
	marshalizer  marshal.Marshalizer
	hasher       hashing.Hasher
	mutOperation sync.RWMutex

	oldHashes            [][]byte
	oldRoot              []byte
	maxTrieLevelInMemory uint
	chanClose            chan struct{}
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(
	trieStorage common.StorageManager,
	msh marshal.Marshalizer,
	hsh hashing.Hasher,
	maxTrieLevelInMemory uint,
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
	if maxTrieLevelInMemory == 0 {
		return nil, ErrInvalidLevelValue
	}
	log.Trace("created new trie", "max trie level in memory", maxTrieLevelInMemory)

	return &patriciaMerkleTrie{
		trieStorage:          trieStorage,
		marshalizer:          msh,
		hasher:               hsh,
		oldHashes:            make([][]byte, 0),
		oldRoot:              make([]byte, 0),
		maxTrieLevelInMemory: maxTrieLevelInMemory,
		chanClose:            make(chan struct{}),
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

	val, err := tr.root.tryGet(hexKey, tr.trieStorage)
	if err != nil {
		err = fmt.Errorf("trie get error: %w, for key %v", err, hex.EncodeToString(key))
		return nil, err
	}

	return val, nil
}

// Update updates the value at the given key.
// If the key is not in the trie, it will be added.
// If the value is empty, the key will be removed from the trie
func (tr *patriciaMerkleTrie) Update(key, value []byte) error {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	log.Trace("update trie", "key", hex.EncodeToString(key), "val", hex.EncodeToString(value))

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

		newRoot, oldHashes, err = tr.root.insert(newLn, tr.trieStorage)
		if err != nil {
			return err
		}

		if check.IfNil(newRoot) {
			return nil
		}

		tr.root = newRoot
		tr.oldHashes = append(tr.oldHashes, oldHashes...)

		logArrayWithTrace("oldHashes after insert", "hash", oldHashes)
	} else {
		if tr.root == nil {
			return nil
		}

		if !tr.root.isDirty() {
			tr.oldRoot = tr.root.getHash()
		}

		_, newRoot, oldHashes, err = tr.root.delete(hexKey, tr.trieStorage)
		if err != nil {
			return err
		}
		tr.root = newRoot
		tr.oldHashes = append(tr.oldHashes, oldHashes...)

		logArrayWithTrace("oldHashes after delete", "hash", oldHashes)
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

	_, newRoot, oldHashes, err := tr.root.delete(hexKey, tr.trieStorage)
	if err != nil {
		return err
	}
	tr.root = newRoot
	tr.oldHashes = append(tr.oldHashes, oldHashes...)

	return nil
}

// RootHash returns the hash of the root node
func (tr *patriciaMerkleTrie) RootHash() ([]byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	return tr.getRootHash()
}

func (tr *patriciaMerkleTrie) getRootHash() ([]byte, error) {
	if tr.root == nil {
		return EmptyTrieHash, nil
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

// Commit adds all the dirty nodes to the database
func (tr *patriciaMerkleTrie) Commit() error {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return nil
	}
	if !tr.root.isDirty() {
		return nil
	}
	err := tr.root.setRootHash()
	if err != nil {
		return err
	}

	tr.oldRoot = make([]byte, 0)
	tr.oldHashes = make([][]byte, 0)

	if log.GetLevel() == logger.LogTrace {
		log.Trace("started committing trie", "trie", tr.root.getHash())
	}

	err = tr.root.commitDirty(0, tr.maxTrieLevelInMemory, tr.trieStorage, tr.trieStorage)
	if err != nil {
		return err
	}

	return nil
}

// Recreate returns a new trie that has the given root hash and database
func (tr *patriciaMerkleTrie) Recreate(root []byte) (common.Trie, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	return tr.recreate(root, tr.trieStorage)
}

// RecreateFromEpoch returns a new trie, given the options
func (tr *patriciaMerkleTrie) RecreateFromEpoch(options common.RootHashHolder) (common.Trie, error) {
	if check.IfNil(options) {
		return nil, ErrNilRootHashHolder
	}

	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if !options.GetEpoch().HasValue {
		return tr.recreate(options.GetRootHash(), tr.trieStorage)
	}

	tsmie, err := newTrieStorageManagerInEpoch(tr.trieStorage, options.GetEpoch().Value)
	if err != nil {
		return nil, err
	}

	return tr.recreate(options.GetRootHash(), tsmie)
}

func (tr *patriciaMerkleTrie) recreate(root []byte, tsm common.StorageManager) (*patriciaMerkleTrie, error) {
	if emptyTrie(root) {
		return NewTrie(
			tr.trieStorage,
			tr.marshalizer,
			tr.hasher,
			tr.maxTrieLevelInMemory,
		)
	}

	_, err := tsm.Get(root)
	if err != nil {
		return nil, err
	}

	newTr, _, err := tr.recreateFromDb(root, tsm)
	if err != nil {
		if errors.IsClosingError(err) {
			log.Debug("could not recreate", "rootHash", root, "error", err)
			return nil, err
		}

		log.Warn("trie recreate error:", "error", err, "root", hex.EncodeToString(root))
		return nil, err
	}

	return newTr, nil
}

// String outputs a graphical view of the trie. Mainly used in tests/debugging
func (tr *patriciaMerkleTrie) String() string {
	writer := bytes.NewBuffer(make([]byte, 0))

	if tr.root == nil {
		_, _ = fmt.Fprintln(writer, "*** EMPTY TRIE ***")
	} else {
		tr.root.print(writer, 0, tr.trieStorage)
	}

	return writer.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (tr *patriciaMerkleTrie) IsInterfaceNil() bool {
	return tr == nil
}

func emptyTrie(root []byte) bool {
	if len(root) == 0 {
		return true
	}
	if bytes.Equal(root, EmptyTrieHash) {
		return true
	}
	return false
}

// GetObsoleteHashes resets the oldHashes and oldRoot variables and returns the old hashes
func (tr *patriciaMerkleTrie) GetObsoleteHashes() [][]byte {
	tr.mutOperation.Lock()
	oldHashes := tr.oldHashes
	logArrayWithTrace("old trie hash", "hash", oldHashes)

	tr.mutOperation.Unlock()

	return oldHashes
}

// GetDirtyHashes returns all the dirty hashes from the trie
func (tr *patriciaMerkleTrie) GetDirtyHashes() (common.ModifiedHashes, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return nil, nil
	}

	err := tr.root.setRootHash()
	if err != nil {
		return nil, err
	}

	dirtyHashes := make(common.ModifiedHashes)
	err = tr.root.getDirtyHashes(dirtyHashes)
	if err != nil {
		return nil, err
	}

	logMapWithTrace("new trie hash", "hash", dirtyHashes)

	return dirtyHashes, nil
}

func (tr *patriciaMerkleTrie) recreateFromDb(rootHash []byte, tsm common.StorageManager) (*patriciaMerkleTrie, snapshotNode, error) {
	newTr, err := NewTrie(
		tsm,
		tr.marshalizer,
		tr.hasher,
		tr.maxTrieLevelInMemory,
	)
	if err != nil {
		return nil, nil, err
	}

	newRoot, err := getNodeFromDBAndDecode(rootHash, tsm, tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, nil, err
	}

	newRoot.setGivenHash(rootHash)
	newTr.root = newRoot

	return newTr, newRoot, nil
}

// GetSerializedNode returns the serialized node (if existing) provided the node's hash
func (tr *patriciaMerkleTrie) GetSerializedNode(hash []byte) ([]byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	log.Trace("GetSerializedNode", "hash", hash)

	return tr.trieStorage.Get(hash)
}

// GetSerializedNodes returns a batch of serialized nodes from the trie, starting from the given hash
func (tr *patriciaMerkleTrie) GetSerializedNodes(rootHash []byte, maxBuffToSend uint64) ([][]byte, uint64, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	log.Trace("GetSerializedNodes", "rootHash", rootHash)
	size := uint64(0)

	newTr, _, err := tr.recreateFromDb(rootHash, tr.trieStorage)
	if err != nil {
		return nil, 0, err
	}

	it, err := NewIterator(newTr)
	if err != nil {
		return nil, 0, err
	}

	encNode, err := it.MarshalizedNode()
	if err != nil {
		return nil, 0, err
	}

	nodes := make([][]byte, 0)
	nodes = append(nodes, encNode)
	size += uint64(len(encNode))

	for it.HasNext() {
		err = it.Next()
		if err != nil {
			return nil, 0, err
		}

		encNode, err = it.MarshalizedNode()
		if err != nil {
			return nil, 0, err
		}

		if size+uint64(len(encNode)) > maxBuffToSend {
			return nodes, 0, nil
		}
		nodes = append(nodes, encNode)
		size += uint64(len(encNode))
	}

	remainingSpace := maxBuffToSend - size

	return nodes, remainingSpace, nil
}

// GetAllLeavesOnChannel adds all the trie leaves to the given channel
func (tr *patriciaMerkleTrie) GetAllLeavesOnChannel(
	leavesChannel chan core.KeyValueHolder,
	ctx context.Context,
	rootHash []byte,
	keyBuilder common.KeyBuilder,
) error {
	tr.mutOperation.RLock()
	newTrie, err := tr.recreate(rootHash, tr.trieStorage)
	if err != nil {
		tr.mutOperation.RUnlock()
		close(leavesChannel)
		return err
	}

	if check.IfNil(newTrie) || newTrie.root == nil {
		tr.mutOperation.RUnlock()
		close(leavesChannel)
		return nil
	}

	tr.trieStorage.EnterPruningBufferingMode()
	tr.mutOperation.RUnlock()

	go func() {
		err = newTrie.root.getAllLeavesOnChannel(
			leavesChannel,
			keyBuilder,
			tr.trieStorage,
			tr.marshalizer,
			tr.chanClose,
			ctx,
		)
		if err != nil {
			log.Error("could not get all trie leaves: ", "error", err)
		}

		tr.mutOperation.Lock()
		tr.trieStorage.ExitPruningBufferingMode()
		tr.mutOperation.Unlock()

		close(leavesChannel)
	}()

	return nil
}

// GetAllHashes returns all the hashes from the trie
func (tr *patriciaMerkleTrie) GetAllHashes() ([][]byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	hashes := make([][]byte, 0)
	if tr.root == nil {
		return hashes, nil
	}

	err := tr.root.setRootHash()
	if err != nil {
		return nil, err
	}

	hashes, err = tr.root.getAllHashes(tr.trieStorage)
	if err != nil {
		return nil, err
	}

	return hashes, nil
}

func logArrayWithTrace(message string, paramName string, hashes [][]byte) {
	if log.GetLevel() == logger.LogTrace {
		for _, hash := range hashes {
			log.Trace(message, paramName, hash)
		}
	}
}

func logMapWithTrace(message string, paramName string, hashes common.ModifiedHashes) {
	if log.GetLevel() == logger.LogTrace {
		for key := range hashes {
			log.Trace(message, paramName, []byte(key))
		}
	}
}

// GetProof computes a Merkle proof for the node that is present at the given key
func (tr *patriciaMerkleTrie) GetProof(key []byte) ([][]byte, []byte, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return nil, nil, ErrNilNode
	}

	var proof [][]byte
	hexKey := keyBytesToHex(key)
	currentNode := tr.root

	err := currentNode.setRootHash()
	if err != nil {
		return nil, nil, err
	}

	for {
		encodedNode, errGet := currentNode.getEncodedNode()
		if errGet != nil {
			return nil, nil, errGet
		}
		proof = append(proof, encodedNode)
		value := currentNode.getValue()

		currentNode, hexKey, errGet = currentNode.getNext(hexKey, tr.trieStorage)
		if errGet != nil {
			return nil, nil, errGet
		}

		if currentNode == nil {
			return proof, value, nil
		}
	}
}

// VerifyProof verifies the given Merkle proof
func (tr *patriciaMerkleTrie) VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	wantHash := rootHash
	key = keyBytesToHex(key)
	for _, encodedNode := range proof {
		if encodedNode == nil {
			return false, nil
		}

		hash := tr.hasher.Compute(string(encodedNode))
		if !bytes.Equal(wantHash, hash) {
			return false, nil
		}

		n, errDecode := decodeNode(encodedNode, tr.marshalizer, tr.hasher)
		if errDecode != nil {
			return false, errDecode
		}

		var proofVerified bool
		proofVerified, wantHash, key = n.getNextHashAndKey(key)
		if proofVerified {
			return true, nil
		}
	}

	return false, nil
}

// GetNumNodes will return the trie nodes statistics DTO
func (tr *patriciaMerkleTrie) GetNumNodes() common.NumNodesDTO {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	n := tr.root
	if check.IfNil(n) {
		return common.NumNodesDTO{}
	}

	return n.getNumNodes()
}

// GetStorageManager returns the storage manager for the trie
func (tr *patriciaMerkleTrie) GetStorageManager() common.StorageManager {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	return tr.trieStorage
}

// GetOldRoot returns the rootHash of the trie before the latest changes
func (tr *patriciaMerkleTrie) GetOldRoot() []byte {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	return tr.oldRoot
}

// Close stops all the active goroutines started by the trie
func (tr *patriciaMerkleTrie) Close() error {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if !isChannelClosed(tr.chanClose) {
		close(tr.chanClose)
	}

	return nil
}

func isChannelClosed(ch chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}
