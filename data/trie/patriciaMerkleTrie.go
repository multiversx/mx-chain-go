package trie

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
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

	trieStorage  data.StorageManager
	marshalizer  marshal.Marshalizer
	hasher       hashing.Hasher
	mutOperation sync.RWMutex

	oldHashes [][]byte
	oldRoot   []byte
	newHashes data.ModifiedHashes

	maxTrieLevelInMemory uint
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(
	trieStorage data.StorageManager,
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
	if maxTrieLevelInMemory <= 0 {
		return nil, ErrInvalidLevelValue
	}
	log.Trace("created new trie", "max trie level in memory", maxTrieLevelInMemory)

	return &patriciaMerkleTrie{
		trieStorage:          trieStorage,
		marshalizer:          msh,
		hasher:               hsh,
		oldHashes:            make([][]byte, 0),
		oldRoot:              make([]byte, 0),
		newHashes:            make(data.ModifiedHashes),
		maxTrieLevelInMemory: maxTrieLevelInMemory,
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

		_, newRoot, oldHashes, err = tr.root.insert(newLn, tr.trieStorage.Database())
		if err != nil {
			return err
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

		_, newRoot, oldHashes, err = tr.root.delete(hexKey, tr.trieStorage.Database())
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

	if tr.trieStorage.IsPruningEnabled() {
		err = tr.markForEviction()
		if err != nil {
			return err
		}
	}

	tr.newHashes = make(data.ModifiedHashes)
	tr.oldRoot = make([]byte, 0)
	tr.oldHashes = make([][]byte, 0)

	if log.GetLevel() == logger.LogTrace {
		log.Trace("started committing trie", "trie", tr.root.getHash())
	}

	err = tr.root.commit(false, 0, tr.maxTrieLevelInMemory, tr.trieStorage.Database(), tr.trieStorage.Database())
	if err != nil {
		return err
	}

	return nil
}

func (tr *patriciaMerkleTrie) markForEviction() error {
	newRoot := tr.root.getHash()

	if bytes.Equal(newRoot, tr.oldRoot) {
		log.Trace("old root and new root are identical", "rootHash", newRoot)
		return nil
	}

	oldHashes := make(data.ModifiedHashes)
	for i := range tr.oldHashes {
		oldHashes[hex.EncodeToString(tr.oldHashes[i])] = struct{}{}
	}

	log.Trace("trie hashes sizes", "newHashes", len(tr.newHashes), "oldHashes", len(oldHashes))
	removeDuplicatedKeys(oldHashes, tr.newHashes)

	if len(tr.newHashes) > 0 && len(newRoot) > 0 {
		newRoot = append(newRoot, byte(data.NewRoot))
		err := tr.trieStorage.MarkForEviction(newRoot, tr.newHashes)
		if err != nil {
			return err
		}

		logMapWithTrace("MarkForEviction newHashes", "hash", tr.newHashes)
	}

	if len(oldHashes) > 0 && len(tr.oldRoot) > 0 {
		tr.oldRoot = append(tr.oldRoot, byte(data.OldRoot))
		err := tr.trieStorage.MarkForEviction(tr.oldRoot, oldHashes)
		if err != nil {
			return err
		}

		logMapWithTrace("MarkForEviction oldHashes", "hash", oldHashes)
	}
	return nil
}

func removeDuplicatedKeys(oldHashes map[string]struct{}, newHashes map[string]struct{}) {
	for key := range oldHashes {
		_, ok := newHashes[key]
		if ok {
			delete(oldHashes, key)
			delete(newHashes, key)
			log.Trace("found in newHashes and oldHashes", "hash", key)
		}
	}
}

// Recreate returns a new trie that has the given root hash and database
func (tr *patriciaMerkleTrie) Recreate(root []byte) (data.Trie, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if emptyTrie(root) {
		return NewTrie(
			tr.trieStorage,
			tr.marshalizer,
			tr.hasher,
			tr.maxTrieLevelInMemory,
		)
	}

	newTr := tr.recreateFromMainDb(root)
	if !check.IfNil(newTr) {
		return newTr, nil
	}

	return tr.recreateFromSnapshotDb(root)
}

func (tr *patriciaMerkleTrie) recreateFromMainDb(rootHash []byte) data.Trie {
	_, err := tr.Database().Get(rootHash)
	if err != nil {
		return nil
	}

	newTr, _, err := tr.recreateFromDb(rootHash, tr.Database(), tr.trieStorage)
	if err != nil {
		log.Warn("trie recreate error:", "error", err, "root", hex.EncodeToString(rootHash))
		return nil
	}

	return newTr
}

func (tr *patriciaMerkleTrie) recreateFromSnapshotDb(rootHash []byte) (data.Trie, error) {
	db := tr.trieStorage.GetSnapshotThatContainsHash(rootHash)
	if db == nil {
		return nil, ErrHashNotFound
	}
	defer db.DecreaseNumReferences()

	newTr, newRoot, err := tr.recreateFromDb(rootHash, db, tr.trieStorage)
	if err != nil {
		err = fmt.Errorf("trie recreate error: %w, for root %v", err, hex.EncodeToString(rootHash))
		return nil, err
	}

	err = newRoot.commit(true, 0, tr.maxTrieLevelInMemory, db, tr.Database())
	if err != nil {
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
		tr.root.print(writer, 0, tr.Database())
	}

	return writer.String()
}

// ClosePersister will close trie persister
func (tr *patriciaMerkleTrie) ClosePersister() error {
	return tr.trieStorage.Database().Close()
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

// Prune removes from the database all the old hashes that correspond to the given root hash
func (tr *patriciaMerkleTrie) Prune(rootHash []byte, identifier data.TriePruningIdentifier) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	tr.trieStorage.Prune(rootHash, identifier)
}

// CancelPrune invalidates the hashes that correspond to the given root hash from the eviction waiting list
func (tr *patriciaMerkleTrie) CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier) {
	tr.mutOperation.Lock()

	tr.trieStorage.CancelPrune(rootHash, identifier)
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

	logArrayWithTrace("old trie hash", "hash", oldHashes)

	tr.mutOperation.Unlock()

	return oldHashes
}

// GetDirtyHashes returns all the dirty hashes from the trie
func (tr *patriciaMerkleTrie) GetDirtyHashes() (data.ModifiedHashes, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	if tr.root == nil {
		return nil, nil
	}

	err := tr.root.setRootHash()
	if err != nil {
		return nil, err
	}

	dirtyHashes := make(data.ModifiedHashes)
	err = tr.root.getDirtyHashes(dirtyHashes)
	if err != nil {
		return nil, err
	}

	logMapWithTrace("new trie hash", "hash", dirtyHashes)

	return dirtyHashes, nil
}

// SetNewHashes adds the given hashes to tr.newHashes
func (tr *patriciaMerkleTrie) SetNewHashes(newHashes data.ModifiedHashes) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	tr.newHashes = newHashes
}

// SetCheckpoint adds the current state of the trie to the snapshot database
func (tr *patriciaMerkleTrie) SetCheckpoint(rootHash []byte) {
	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not snapshot empty trie")
		return
	}

	tr.trieStorage.SetCheckpoint(rootHash)
}

// TakeSnapshot creates a new database in which the current state of the trie is saved.
// If the maximum number of snapshots has been reached, the oldest snapshot is removed.
func (tr *patriciaMerkleTrie) TakeSnapshot(rootHash []byte) {
	if bytes.Equal(rootHash, EmptyTrieHash) {
		log.Trace("should not snapshot empty trie")
		return
	}

	tr.trieStorage.TakeSnapshot(rootHash)
}

// Database returns the trie database
func (tr *patriciaMerkleTrie) Database() data.DBWriteCacher {
	return tr.trieStorage.Database()
}

func (tr *patriciaMerkleTrie) recreateFromDb(rootHash []byte, db data.DBWriteCacher, tsm data.StorageManager) (data.Trie, snapshotNode, error) {
	newTr, err := NewTrie(
		tsm,
		tr.marshalizer,
		tr.hasher,
		tr.maxTrieLevelInMemory,
	)
	if err != nil {
		return nil, nil, err
	}

	newRoot, err := getNodeFromDBAndDecode(rootHash, db, tr.marshalizer, tr.hasher)
	if err != nil {
		return nil, nil, err
	}

	newRoot.setGivenHash(rootHash)
	newTr.root = newRoot

	return newTr, newRoot, nil
}

// EnterSnapshotMode sets the snapshot mode on
func (tr *patriciaMerkleTrie) EnterSnapshotMode() {
	tr.trieStorage.EnterSnapshotMode()
}

// ExitSnapshotMode sets the snapshot mode off
func (tr *patriciaMerkleTrie) ExitSnapshotMode() {
	tr.trieStorage.ExitSnapshotMode()
}

func getDbThatContainsHash(trieStorage data.StorageManager, rootHash []byte) data.SnapshotDbHandler {
	db := trieStorage.GetSnapshotThatContainsHash(rootHash)
	if db != nil {
		return db
	}

	_, err := trieStorage.Database().Get(rootHash)
	if err != nil {
		return nil
	}

	return &snapshotDb{
		DBWriteCacher: trieStorage.Database(),
	}
}

// GetSerializedNodes returns a batch of serialized nodes from the trie, starting from the given hash
func (tr *patriciaMerkleTrie) GetSerializedNodes(rootHash []byte, maxBuffToSend uint64) ([][]byte, uint64, error) {
	tr.mutOperation.Lock()
	defer tr.mutOperation.Unlock()

	log.Trace("GetSerializedNodes", "rootHash", rootHash)
	size := uint64(0)

	db := getDbThatContainsHash(tr.trieStorage, rootHash)
	if db == nil {
		return nil, 0, ErrHashNotFound
	}
	defer db.DecreaseNumReferences()

	newTsm, err := NewTrieStorageManagerWithoutPruning(db)
	if err != nil {
		return nil, 0, err
	}

	newTr, _, err := tr.recreateFromDb(rootHash, db, newTsm)
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

// GetAllLeaves iterates the trie and returns a map that contains all leafNodes information
// TODO remove this function after GetAllLeavesOnChannel is fully functional
func (tr *patriciaMerkleTrie) GetAllLeaves() (map[string][]byte, error) {
	tr.mutOperation.RLock()
	defer tr.mutOperation.RUnlock()

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

// GetAllLeavesOnChannel adds all the trie leaves to the given channel
func (tr *patriciaMerkleTrie) GetAllLeavesOnChannel() chan core.KeyValueHolder {
	//TODO pass a context for cancellation purposes when needed
	leavesChannel := make(chan core.KeyValueHolder)

	tr.mutOperation.RLock()
	if tr.root == nil {
		tr.mutOperation.RUnlock()
		close(leavesChannel)

		return leavesChannel
	}
	tr.mutOperation.RUnlock()

	go func() {
		tr.mutOperation.RLock()
		err := tr.root.getAllLeavesOnChannel(leavesChannel, []byte{}, tr.Database(), tr.marshalizer)
		if err != nil {
			log.Error("could not get all trie leaves: ", "error", err)
		}
		tr.mutOperation.RUnlock()

		close(leavesChannel)
	}()

	return leavesChannel
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

	hashes, err = tr.root.getAllHashes(tr.Database())
	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// IsPruningEnabled returns true if state pruning is enabled
func (tr *patriciaMerkleTrie) IsPruningEnabled() bool {
	return tr.trieStorage.IsPruningEnabled()
}

func logArrayWithTrace(message string, paramName string, hashes [][]byte) {
	if log.GetLevel() == logger.LogTrace {
		for _, hash := range hashes {
			log.Trace(message, paramName, hash)
		}
	}
}

func logMapWithTrace(message string, paramName string, hashes data.ModifiedHashes) {
	if log.GetLevel() == logger.LogTrace {
		for key := range hashes {
			log.Trace(message, paramName, key)
		}
	}
}

// GetSnapshotDbBatchDelay returns the batch write delay in seconds
func (tr *patriciaMerkleTrie) GetSnapshotDbBatchDelay() int {
	return tr.trieStorage.GetSnapshotDbBatchDelay()
}
