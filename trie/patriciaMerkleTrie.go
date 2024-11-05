package trie

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/trieBatchManager"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("trie")

var _ dataRetriever.TrieDataGetter = (*patriciaMerkleTrie)(nil)

const (
	extension = iota
	leaf
	branch
)

const rootDepthLevel = 0

type patriciaMerkleTrie struct {
	RootManager

	trieStorage             common.StorageManager
	marshalizer             marshal.Marshalizer
	hasher                  hashing.Hasher
	enableEpochsHandler     common.EnableEpochsHandler
	trieNodeVersionVerifier core.TrieNodeVersionVerifier
	batchManager            common.TrieBatchManager
	goroutinesThrottler     core.Throttler
	trieOperationInProgress *atomic.Flag
	updateTrieMutex         sync.RWMutex

	maxTrieLevelInMemory uint
	chanClose            chan struct{}
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(
	trieStorage common.StorageManager,
	msh marshal.Marshalizer,
	hsh hashing.Hasher,
	enableEpochsHandler common.EnableEpochsHandler,
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
	if check.IfNil(enableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}
	if maxTrieLevelInMemory == 0 {
		return nil, ErrInvalidLevelValue
	}
	log.Trace("created new trie", "max trie level in memory", maxTrieLevelInMemory)

	tnvv, err := core.NewTrieNodeVersionVerifier(enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	// TODO give this as an argument
	trieThrottler, err := throttler.NewNumGoRoutinesThrottler(20)
	if err != nil {
		return nil, err
	}

	return &patriciaMerkleTrie{
		RootManager:             NewRootManager(),
		trieStorage:             trieStorage,
		marshalizer:             msh,
		hasher:                  hsh,
		maxTrieLevelInMemory:    maxTrieLevelInMemory,
		chanClose:               make(chan struct{}),
		enableEpochsHandler:     enableEpochsHandler,
		trieNodeVersionVerifier: tnvv,
		batchManager:            trieBatchManager.NewTrieBatchManager(),
		goroutinesThrottler:     trieThrottler,
		trieOperationInProgress: &atomic.Flag{},
	}, nil
}

// Get starts at the root and searches for the given key.
// If the key is present in the tree, it returns the corresponding value
func (tr *patriciaMerkleTrie) Get(key []byte) ([]byte, uint32, error) {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	hexKey := keyBytesToHex(key)
	//TODO in order to reduce the memory usage, store keys as bytes and convert them to hex only when needed
	val, found := tr.batchManager.Get(hexKey)
	if found {
		return val, 0, nil
	}

	rootNode := tr.GetRootNode()
	if check.IfNil(rootNode) {
		return nil, 0, nil
	}

	val, depth, err := rootNode.tryGet(hexKey, rootDepthLevel, tr.trieStorage)
	if err != nil {
		err = fmt.Errorf("trie get error: %w, for key %v", err, hex.EncodeToString(key))
		return nil, depth, err
	}

	return val, depth, nil
}

// Update updates the value at the given key.
// If the key is not in the trie, it will be added.
// If the value is empty, the key will be removed from the trie
func (tr *patriciaMerkleTrie) Update(key, value []byte) error {
	log.Trace("update trie",
		"key", hex.EncodeToString(key),
		"val", hex.EncodeToString(value),
	)

	return tr.updateBatch(key, value, core.NotSpecified)
}

// UpdateWithVersion does the same thing as Update, but the new leaf that is created will be of the specified version
func (tr *patriciaMerkleTrie) UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) error {
	log.Trace("update trie with version",
		"key", hex.EncodeToString(key),
		"val", hex.EncodeToString(value),
		"version", version,
	)

	return tr.updateBatch(key, value, version)
}

func (tr *patriciaMerkleTrie) updateBatch(key []byte, value []byte, version core.TrieNodeVersion) error {
	hexKey := keyBytesToHex(key)
	if len(value) != 0 {
		newData := core.TrieData{
			Key:     hexKey,
			Value:   value,
			Version: version,
		}
		tr.batchManager.Add(newData)
		return nil
	}

	tr.batchManager.MarkForRemoval(hexKey)
	return nil
}

// Delete removes the node that has the given key from the tree
func (tr *patriciaMerkleTrie) Delete(key []byte) {
	hexKey := keyBytesToHex(key)
	tr.batchManager.MarkForRemoval(hexKey)
}

func (tr *patriciaMerkleTrie) updateTrie() error {
	tr.updateTrieMutex.Lock()
	defer tr.updateTrieMutex.Unlock()

	batch, err := tr.batchManager.MarkTrieUpdateInProgress()
	if err != nil {
		return err
	}
	defer tr.batchManager.MarkTrieUpdateCompleted()

	err = tr.insertBatch(batch.GetSortedDataForInsertion())
	if err != nil {
		return err
	}

	return tr.deleteBatch(batch.GetSortedDataForRemoval())
}

func (tr *patriciaMerkleTrie) insertBatch(sortedDataForInsertion []core.TrieData) error {
	if len(sortedDataForInsertion) == 0 {
		return nil
	}

	rootNode := tr.GetRootNode()
	if rootNode == nil {
		newRoot, err := newLeafNode(sortedDataForInsertion[0], tr.marshalizer, tr.hasher)
		if err != nil {
			return err
		}

		sortedDataForInsertion = sortedDataForInsertion[1:]
		if len(sortedDataForInsertion) == 0 {
			tr.SetNewRootNode(newRoot)
			return nil
		}

		rootNode = newRoot
	}

	var oldRootHash []byte
	if !rootNode.isDirty() {
		oldRootHash = rootNode.getHash()
	}

	manager, err := NewGoroutinesManager(tr.goroutinesThrottler, errChan.NewErrChanWrapper(), tr.chanClose)
	if err != nil {
		return err
	}

	newRoot, oldHashes := rootNode.insert(sortedDataForInsertion, manager, tr.trieStorage)
	err = manager.GetError()
	if err != nil {
		return err
	}

	if check.IfNil(newRoot) {
		return nil
	}

	tr.SetDataForRootChange(newRoot, oldRootHash, oldHashes)

	logArrayWithTrace("oldHashes after insert", "hash", oldHashes)
	return nil
}

func (tr *patriciaMerkleTrie) deleteBatch(data []core.TrieData) error {
	if len(data) == 0 {
		return nil
	}

	rootNode := tr.GetRootNode()
	if rootNode == nil {
		return nil
	}

	var oldRootHash []byte
	if !rootNode.isDirty() {
		oldRootHash = rootNode.getHash()
	}

	manager, err := NewGoroutinesManager(tr.goroutinesThrottler, errChan.NewErrChanWrapper(), tr.chanClose)
	if err != nil {
		return err
	}

	_, newRoot, oldHashes := rootNode.delete(data, manager, tr.trieStorage)
	err = manager.GetError()
	if err != nil {
		return err
	}

	tr.SetDataForRootChange(newRoot, oldRootHash, oldHashes)
	logArrayWithTrace("oldHashes after delete", "hash", oldHashes)

	return nil
}

// RootHash returns the hash of the root node
func (tr *patriciaMerkleTrie) RootHash() ([]byte, error) {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	err := tr.updateTrie()
	if err != nil {
		return nil, err
	}

	return tr.getRootHash()
}

func (tr *patriciaMerkleTrie) getRootHash() ([]byte, error) {
	rootNode := tr.GetRootNode()

	if rootNode == nil {
		return common.EmptyTrieHash, nil
	}

	hash := rootNode.getHash()
	if hash != nil {
		return hash, nil
	}
	err := rootNode.setRootHash()
	if err != nil {
		return nil, err
	}
	return rootNode.getHash(), nil
}

// Commit adds all the dirty nodes to the database
func (tr *patriciaMerkleTrie) Commit() error {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	err := tr.updateTrie()
	if err != nil {
		return err
	}

	rootNode := tr.GetRootNode()
	if rootNode == nil {
		log.Trace("trying to commit empty trie")
		return nil
	}
	if !rootNode.isDirty() {
		log.Trace("trying to commit clean trie", "root", rootNode.getHash())

		tr.ResetCollectedHashes()

		return nil
	}
	err = rootNode.setRootHash()
	if err != nil {
		return err
	}

	tr.ResetCollectedHashes()
	if log.GetLevel() == logger.LogTrace {
		log.Trace("started committing trie", "trie", rootNode.getHash())
	}

	return rootNode.commitDirty(0, tr.maxTrieLevelInMemory, tr.trieStorage, tr.trieStorage)
}

// Recreate returns a new trie that has the given root hash and database
func (tr *patriciaMerkleTrie) Recreate(root []byte) (common.Trie, error) {
	return tr.recreate(root, tr.trieStorage)
}

// RecreateFromEpoch returns a new trie, given the options
func (tr *patriciaMerkleTrie) RecreateFromEpoch(options common.RootHashHolder) (common.Trie, error) {
	if check.IfNil(options) {
		return nil, ErrNilRootHashHolder
	}

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
	if common.IsEmptyTrie(root) {
		return NewTrie(
			tr.trieStorage,
			tr.marshalizer,
			tr.hasher,
			tr.enableEpochsHandler,
			tr.maxTrieLevelInMemory,
		)
	}

	newTr, _, err := tr.recreateFromDb(root, tsm)
	if err != nil {
		if core.IsClosingError(err) {
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
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	err := tr.updateTrie()
	if err != nil {
		log.Warn("print trie - could not save batched changes", "error", err)
		return ""
	}

	writer := bytes.NewBuffer(make([]byte, 0))

	rootNode := tr.GetRootNode()
	if rootNode == nil {
		_, _ = fmt.Fprintln(writer, "*** EMPTY TRIE ***")
	} else {
		rootNode.print(writer, 0, tr.trieStorage)
	}

	return writer.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (tr *patriciaMerkleTrie) IsInterfaceNil() bool {
	return tr == nil
}

// GetObsoleteHashes resets the oldHashes and oldRoot variables and returns the old hashes
func (tr *patriciaMerkleTrie) GetObsoleteHashes() [][]byte {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	err := tr.updateTrie()
	if err != nil {
		log.Warn("get obsolete hashes - could not save batched changes", "error", err)
	}

	oldHashes := tr.GetOldHashes()
	logArrayWithTrace("old trie hash", "hash", oldHashes)

	return oldHashes
}

// GetDirtyHashes returns all the dirty hashes from the trie
func (tr *patriciaMerkleTrie) GetDirtyHashes() (common.ModifiedHashes, error) {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	err := tr.updateTrie()
	if err != nil {
		return nil, err
	}

	rootNode := tr.GetRootNode()
	if rootNode == nil {
		return nil, nil
	}

	err = rootNode.setRootHash()
	if err != nil {
		return nil, err
	}

	dirtyHashes := make(common.ModifiedHashes)
	err = rootNode.getDirtyHashes(dirtyHashes)
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
		tr.enableEpochsHandler,
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
	newTr.SetNewRootNode(newRoot)

	return newTr, newRoot, nil
}

func (tr *patriciaMerkleTrie) waitForCurrentOperation() {

	for {
		switch {
		case !tr.trieOperationInProgress.SetReturningPrevious():
			return
		case isChannelClosed(tr.chanClose):
			return
		default:
		}
	}
}

// GetSerializedNode returns the serialized node (if existing) provided the node's hash
func (tr *patriciaMerkleTrie) GetSerializedNode(hash []byte) ([]byte, error) {
	// TODO: investigate if we can move the critical section behavior in the trie node resolver as this call will compete with a normal trie.Get operation
	//  which might occur during processing.
	//  warning: A critical section here or on the trie node resolver must be kept as to not overwhelm the node with requests that affects the block processing flow
	tr.waitForCurrentOperation()
	defer tr.trieOperationInProgress.Reset()

	log.Trace("GetSerializedNode", "hash", hash)

	return tr.trieStorage.Get(hash)
}

// GetSerializedNodes returns a batch of serialized nodes from the trie, starting from the given hash
func (tr *patriciaMerkleTrie) GetSerializedNodes(rootHash []byte, maxBuffToSend uint64) ([][]byte, uint64, error) {
	// TODO: investigate if we can move the critical section behavior in the trie node resolver as this call will compete with a normal trie.Get operation
	//  which might occur during processing.
	//  warning: A critical section here or on the trie node resolver must be kept as to not overwhelm the node with requests that affects the block processing flow
	tr.waitForCurrentOperation()
	defer tr.trieOperationInProgress.Reset()

	log.Trace("GetSerializedNodes", "rootHash", rootHash)
	size := uint64(0)

	newTr, _, err := tr.recreateFromDb(rootHash, tr.trieStorage)
	if err != nil {
		return nil, 0, err
	}

	it, err := NewDFSIterator(newTr)
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
	leavesChannels *common.TrieIteratorChannels,
	ctx context.Context,
	rootHash []byte,
	keyBuilder common.KeyBuilder,
	trieLeafParser common.TrieLeafParser,
) error {
	if leavesChannels == nil {
		return ErrNilTrieIteratorChannels
	}
	if leavesChannels.LeavesChan == nil {
		return ErrNilTrieIteratorLeavesChannel
	}
	if leavesChannels.ErrChan == nil {
		return ErrNilTrieIteratorErrChannel
	}
	if check.IfNil(keyBuilder) {
		return ErrNilKeyBuilder
	}
	if check.IfNil(trieLeafParser) {
		return ErrNilTrieLeafParser
	}

	newTrie, err := tr.recreate(rootHash, tr.trieStorage)
	if err != nil {
		close(leavesChannels.LeavesChan)
		leavesChannels.ErrChan.Close()
		return err
	}

	rootNode := newTrie.GetRootNode()

	if check.IfNil(newTrie) || rootNode == nil {
		close(leavesChannels.LeavesChan)
		leavesChannels.ErrChan.Close()
		return nil
	}

	tr.trieStorage.EnterPruningBufferingMode()

	go func() {
		err = rootNode.getAllLeavesOnChannel(
			leavesChannels.LeavesChan,
			keyBuilder,
			trieLeafParser,
			tr.trieStorage,
			tr.marshalizer,
			tr.chanClose,
			ctx,
		)
		if err != nil {
			leavesChannels.ErrChan.WriteInChanNonBlocking(err)
			log.Error("could not get all trie leaves: ", "error", err)
		}

		tr.trieStorage.ExitPruningBufferingMode()

		close(leavesChannels.LeavesChan)
		leavesChannels.ErrChan.Close()
	}()

	return nil
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
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	rootNode := tr.GetRootNode()
	if rootNode == nil {
		return nil, nil, ErrNilNode
	}

	var proof [][]byte
	hexKey := keyBytesToHex(key)
	currentNode := rootNode

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
	ok, err := tr.verifyProof(rootHash, tr.hasher.Compute(string(key)), proof)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	return tr.verifyProof(rootHash, key, proof)
}

func (tr *patriciaMerkleTrie) verifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error) {
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

// GetStorageManager returns the storage manager for the trie
func (tr *patriciaMerkleTrie) GetStorageManager() common.StorageManager {
	return tr.trieStorage
}

// GetOldRoot returns the rootHash of the trie before the latest changes
func (tr *patriciaMerkleTrie) GetOldRoot() []byte {
	return tr.GetOldRootHash()
}

// GetTrieStats will collect and return the statistics for the given rootHash
func (tr *patriciaMerkleTrie) GetTrieStats(address string, rootHash []byte) (common.TrieStatisticsHandler, error) {
	newTrie, err := tr.recreate(rootHash, tr.trieStorage)
	if err != nil {
		return nil, err
	}

	ts := statistics.NewTrieStatistics()
	err = newTrie.GetRootNode().collectStats(ts, rootDepthLevel, newTrie.trieStorage)
	if err != nil {
		return nil, err
	}
	ts.AddAccountInfo(address, rootHash)

	return ts, nil
}

// CollectLeavesForMigration will collect trie leaves that need to be migrated. The leaves are collected in the trieMigrator.
// The traversing of the trie is done in a DFS manner, and it will stop when the gas runs out (this will be signaled by the trieMigrator).
func (tr *patriciaMerkleTrie) CollectLeavesForMigration(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	rootNode := tr.GetRootNode()
	if check.IfNil(rootNode) {
		return nil
	}
	if check.IfNil(args.TrieMigrator) {
		return errors.ErrNilTrieMigrator
	}

	err := tr.checkIfMigrationPossible(args)
	if err != nil {
		return err
	}

	_, err = rootNode.collectLeavesForMigration(args, tr.trieStorage, keyBuilder.NewKeyBuilder())
	if err != nil {
		return err
	}

	return nil
}

func (tr *patriciaMerkleTrie) checkIfMigrationPossible(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	if !tr.trieNodeVersionVerifier.IsValidVersion(args.NewVersion) {
		return fmt.Errorf("%w: newVersion %v", errors.ErrInvalidTrieNodeVersion, args.NewVersion)
	}

	if !tr.trieNodeVersionVerifier.IsValidVersion(args.OldVersion) {
		return fmt.Errorf("%w: oldVersion %v", errors.ErrInvalidTrieNodeVersion, args.OldVersion)
	}

	if args.NewVersion == core.NotSpecified && args.OldVersion == core.AutoBalanceEnabled {
		return fmt.Errorf("%w: cannot migrate from %v to %v", errors.ErrInvalidTrieNodeVersion, core.AutoBalanceEnabled, core.NotSpecified)
	}

	return nil
}

// IsMigratedToLatestVersion returns true if the trie is migrated to the latest version
func (tr *patriciaMerkleTrie) IsMigratedToLatestVersion() (bool, error) {
	rootNode := tr.GetRootNode()
	if check.IfNil(rootNode) {
		return true, nil
	}

	version, err := rootNode.getVersion()
	if err != nil {
		return false, err
	}

	versionForNewlyAddedData := core.GetVersionForNewData(tr.enableEpochsHandler)
	return version == versionForNewlyAddedData, nil
}

// Close stops all the active goroutines started by the trie
func (tr *patriciaMerkleTrie) Close() error {
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
