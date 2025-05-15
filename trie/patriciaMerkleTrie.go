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
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/trieBatchManager"
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
	common.TrieContext

	enableEpochsHandler     common.EnableEpochsHandler
	trieNodeVersionVerifier core.TrieNodeVersionVerifier
	batchManager            common.TrieBatchManager
	goRoutinesManager       common.TrieGoroutinesManager
	trieOperationInProgress *atomic.Flag
	updateTrieMutex         sync.RWMutex
	throttler               core.Throttler

	maxTrieLevelInMemory uint
	chanClose            chan struct{}
	identifier           string
}

type trieContext struct {
	common.StorageManager
	marshal.Marshalizer
	hashing.Hasher
}

// GetStorage returns the storage manager used in the trie context
func (tc *trieContext) GetStorage() common.StorageManager {
	return tc.StorageManager
}

// GetMarshaller returns the marshaller used in the trie context
func (tc *trieContext) GetMarshaller() marshal.Marshalizer {
	return tc.Marshalizer
}

// GetHasher returns the hasher used in the trie context
func (tc *trieContext) GetHasher() hashing.Hasher {
	return tc.Hasher
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *trieContext) IsInterfaceNil() bool {
	return tc == nil
}

// TrieArgs is the arguments for creating a new trie
type TrieArgs struct {
	TrieStorage          common.StorageManager
	Marshalizer          marshal.Marshalizer
	Hasher               hashing.Hasher
	EnableEpochsHandler  common.EnableEpochsHandler
	MaxTrieLevelInMemory uint
	Throttler            core.Throttler
	Identifier           string
}

// NewTrie creates a new Patricia Merkle Trie
func NewTrie(
	args TrieArgs,
) (*patriciaMerkleTrie, error) {
	if check.IfNil(args.TrieStorage) {
		return nil, ErrNilTrieStorage
	}
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, errors.ErrNilEnableEpochsHandler
	}
	if args.MaxTrieLevelInMemory == 0 {
		return nil, ErrInvalidLevelValue
	}
	if check.IfNil(args.Throttler) {
		return nil, ErrNilThrottler
	}
	log.Trace("created new trie", "max trie level in memory", args.MaxTrieLevelInMemory)

	tnvv, err := core.NewTrieNodeVersionVerifier(args.EnableEpochsHandler)
	if err != nil {
		return nil, err
	}

	chanClose := make(chan struct{})
	goRoutinesManager, err := NewGoroutinesManager(args.Throttler, errChan.NewErrChanWrapper(), chanClose, args.Identifier)
	if err != nil {
		return nil, err
	}

	return &patriciaMerkleTrie{
		TrieContext: &trieContext{
			StorageManager: args.TrieStorage,
			Marshalizer:    args.Marshalizer,
			Hasher:         args.Hasher,
		},
		RootManager:             NewRootManager(),
		maxTrieLevelInMemory:    args.MaxTrieLevelInMemory,
		chanClose:               chanClose,
		enableEpochsHandler:     args.EnableEpochsHandler,
		trieNodeVersionVerifier: tnvv,
		batchManager:            trieBatchManager.NewTrieBatchManager(args.Identifier),
		goRoutinesManager:       goRoutinesManager,
		trieOperationInProgress: &atomic.Flag{},
		updateTrieMutex:         sync.RWMutex{},
		throttler:               args.Throttler,
		identifier:              args.Identifier,
	}, nil
}

// Get starts at the root and searches for the given key.
// If the key is present in the tree, it returns the corresponding value
func (tr *patriciaMerkleTrie) Get(key []byte) ([]byte, uint32, error) {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	val, found := tr.batchManager.Get(key)
	if found {
		return val, 0, nil
	}

	rootNode := tr.GetRootNode()
	if check.IfNil(rootNode) {
		return nil, 0, nil
	}

	hexKey := keyBuilder.KeyBytesToHex(key)
	val, depth, err := rootNode.tryGet(hexKey, rootDepthLevel, tr.TrieContext)
	if err != nil {
		err = fmt.Errorf("trie get error: %w, for key %v", err, hex.EncodeToString(key))
		return nil, depth, err
	}

	return val, depth, nil
}

// Update updates the value at the given key.
// If the key is not in the trie, it will be added.
// If the value is empty, the key will be removed from the trie
func (tr *patriciaMerkleTrie) Update(key, value []byte) {
	log.Trace("update trie", "key", key, "val", value)

	tr.updateBatch(key, value, core.NotSpecified)
}

// UpdateWithVersion does the same thing as Update, but the new leaf that is created will be of the specified version
func (tr *patriciaMerkleTrie) UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) {
	log.Trace("update trie with version", "key", key, "val", value, "version", version)

	tr.updateBatch(key, value, version)
}

func (tr *patriciaMerkleTrie) updateBatch(key []byte, value []byte, version core.TrieNodeVersion) {
	if len(value) != 0 {
		newData := core.TrieData{
			Key:     key,
			Value:   value,
			Version: version,
		}
		tr.batchManager.Add(newData)
		return
	}

	tr.batchManager.MarkForRemoval(key)
}

// Delete removes the node that has the given key from the tree
func (tr *patriciaMerkleTrie) Delete(key []byte) {
	tr.batchManager.MarkForRemoval(key)
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
	oldRootHash := tr.getOldRootHash(rootNode)
	var changedHashes [][]byte

	if check.IfNil(rootNode) {
		newRoot := newLeafNode(sortedDataForInsertion[0])
		sortedDataForInsertion = sortedDataForInsertion[1:]
		rootNode = newRoot
	}

	if len(sortedDataForInsertion) == 0 {
		rootHash, err := encodeNodeAndGetHash(rootNode, tr.TrieContext)
		if err != nil {
			return err
		}
		tr.SetDataForRootChange(RootData{
			newRoot:     rootNode,
			newRootHash: rootHash,
			oldRootHash: oldRootHash,
			oldHashes:   changedHashes,
		})
		return nil
	}

	err := tr.goRoutinesManager.SetNewErrorChannel(errChan.NewErrChanWrapper())
	if err != nil {
		return err
	}

	initialSliceCapacity := len(sortedDataForInsertion) * 2 // there are also intermediate nodes that are changed, so we need to collect more hashes
	oldHashes := common.NewModifiedHashesSlice(initialSliceCapacity)
	newRoot := rootNode.insert(sortedDataForInsertion, tr.goRoutinesManager, oldHashes, tr.TrieContext)
	err = tr.goRoutinesManager.GetError()
	if err != nil {
		return err
	}

	if check.IfNil(newRoot) {
		return nil
	}

	if len(oldRootHash) != 0 {
		oldHashes.Append([][]byte{oldRootHash})
	}

	rootHash, err := encodeNodeAndGetHash(newRoot, tr.TrieContext)
	if err != nil {
		return err
	}

	hashes := oldHashes.Get()
	tr.SetDataForRootChange(
		RootData{
			newRoot:     newRoot,
			newRootHash: rootHash,
			oldRootHash: oldRootHash,
			oldHashes:   hashes,
		})

	logArrayWithTrace("oldHashes after insert", "hash", hashes)
	return nil
}

func (tr *patriciaMerkleTrie) getOldRootHash(rootNode node) []byte {
	if check.IfNil(rootNode) {
		return []byte{}
	}

	if rootNode.isDirty() {
		return tr.GetOldRootHash()
	}

	return tr.GetRootHash()
}

func (tr *patriciaMerkleTrie) deleteBatch(data []core.TrieData) error {
	if len(data) == 0 {
		return nil
	}

	rootNode := tr.GetRootNode()
	if check.IfNil(rootNode) {
		return nil
	}

	oldRootHash := tr.getOldRootHash(rootNode)
	err := tr.goRoutinesManager.SetNewErrorChannel(errChan.NewErrChanWrapper())
	if err != nil {
		return err
	}

	initialSliceCapacity := len(data) * 2 // there are also intermediate nodes that are changed, so we need to collect more hashes
	modifiedHashes := common.NewModifiedHashesSlice(initialSliceCapacity)
	hasBeenModified, newRoot := rootNode.delete(data, tr.goRoutinesManager, modifiedHashes, tr.TrieContext)
	err = tr.goRoutinesManager.GetError()
	if err != nil {
		return err
	}
	if !hasBeenModified {
		return nil
	}

	rootHash := common.EmptyTrieHash
	if !check.IfNil(newRoot) {
		rootHash, err = encodeNodeAndGetHash(newRoot, tr.TrieContext)
		if err != nil {
			return err
		}
	}

	if len(oldRootHash) != 0 {
		modifiedHashes.Append([][]byte{oldRootHash})
	}

	oldHashes := modifiedHashes.Get()
	tr.SetDataForRootChange(
		RootData{
			newRoot:     newRoot,
			newRootHash: rootHash,
			oldRootHash: oldRootHash,
			oldHashes:   oldHashes,
		})
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

	tr.updateTrieMutex.Lock()
	defer tr.updateTrieMutex.Unlock()

	rootHash := tr.GetRootHash()
	if len(rootHash) == 0 {
		return common.EmptyTrieHash, nil
	}

	return rootHash, nil
}

// Commit adds all the dirty nodes to the database
func (tr *patriciaMerkleTrie) Commit(hashesCollector common.TrieHashesCollector) error {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	err := tr.updateTrie()
	if err != nil {
		return err
	}

	tr.updateTrieMutex.Lock()
	defer tr.updateTrieMutex.Unlock()

	rootNode := tr.GetRootNode()
	if check.IfNil(rootNode) {
		log.Trace("trying to commit empty trie")
		return nil
	}
	if !rootNode.isDirty() {
		log.Trace("trying to commit clean trie", "root", tr.GetRootNode())

		tr.ResetCollectedHashes()

		return nil
	}

	oldRootHash := tr.GetOldRootHash()
	if log.GetLevel() == logger.LogTrace {
		log.Trace("started committing trie", "trie", tr.GetRootNode())
	}

	err = tr.goRoutinesManager.SetNewErrorChannel(errChan.NewErrChanWrapper())
	if err != nil {
		return err
	}

	rootNode.commitDirty(0, tr.maxTrieLevelInMemory, tr.goRoutinesManager, hashesCollector, tr.TrieContext)
	err = tr.goRoutinesManager.GetError()
	if err != nil {
		return err
	}

	oldHashes := tr.GetOldHashes()
	hashesCollector.AddObsoleteHashes(oldRootHash, oldHashes)

	logArrayWithTrace("old trie hash", "hash", oldHashes)
	logMapWithTrace("new trie hash", "hash", hashesCollector.GetDirtyHashes())

	tr.ResetCollectedHashes()
	return nil
}

// Recreate returns a new trie, given the options
func (tr *patriciaMerkleTrie) Recreate(options common.RootHashHolder, identifier string) (common.Trie, error) {
	if check.IfNil(options) {
		return nil, ErrNilRootHashHolder
	}

	if !options.GetEpoch().HasValue {
		return tr.recreate(options.GetRootHash(), identifier, tr.TrieContext.GetStorage())
	}

	tsmie, err := newTrieStorageManagerInEpoch(tr.TrieContext.GetStorage(), options.GetEpoch().Value)
	if err != nil {
		return nil, err
	}

	return tr.recreate(options.GetRootHash(), identifier, tsmie)
}

func (tr *patriciaMerkleTrie) recreate(root []byte, identifier string, tsm common.StorageManager) (*patriciaMerkleTrie, error) {
	if common.IsEmptyTrie(root) {
		return NewTrie(
			TrieArgs{
				TrieStorage:          tsm,
				Marshalizer:          tr.TrieContext.GetMarshaller(),
				Hasher:               tr.TrieContext.GetHasher(),
				EnableEpochsHandler:  tr.enableEpochsHandler,
				MaxTrieLevelInMemory: tr.maxTrieLevelInMemory,
				Throttler:            tr.throttler,
				Identifier:           identifier,
			},
		)
	}

	newTr, _, err := tr.recreateFromDb(root, identifier, tsm)
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

// ToString outputs a graphical view of the trie. Mainly used in tests/kdebugging
func (tr *patriciaMerkleTrie) ToString() string {
	tr.trieOperationInProgress.SetValue(true)
	defer tr.trieOperationInProgress.Reset()

	tr.updateTrieMutex.Lock()
	defer tr.updateTrieMutex.Unlock()

	writer := bytes.NewBuffer(make([]byte, 0))

	rootNode := tr.GetRootNode()
	if rootNode == nil {
		_, _ = fmt.Fprintln(writer, "*** EMPTY TRIE ***")
	} else {
		rootNode.print(writer, 0, tr.TrieContext)
	}

	return writer.String()
}

// IsInterfaceNil returns true if there is no value under the interface
func (tr *patriciaMerkleTrie) IsInterfaceNil() bool {
	return tr == nil
}

func (tr *patriciaMerkleTrie) recreateFromDb(rootHash []byte, identifier string, tsm common.StorageManager) (*patriciaMerkleTrie, snapshotNode, error) {
	newTr, err := NewTrie(
		TrieArgs{
			tsm,
			tr.TrieContext.GetMarshaller(),
			tr.TrieContext.GetHasher(),
			tr.enableEpochsHandler,
			tr.maxTrieLevelInMemory,
			tr.throttler,
			identifier,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	newRoot, _, err := getNodeFromDBAndDecode(rootHash, tr.TrieContext)
	if err != nil {
		return nil, nil, err
	}

	newTr.SetDataForRootChange(RootData{
		newRoot:     newRoot,
		newRootHash: rootHash,
		oldRootHash: tr.GetOldRootHash(),
		oldHashes:   tr.GetOldHashes(),
	})

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

	return tr.TrieContext.Get(hash)
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

	it, err := NewDFSIterator(tr, rootHash)
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

	newTrie, err := tr.recreate(rootHash, "", tr.TrieContext)
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

	tr.TrieContext.EnterPruningBufferingMode()

	go func() {
		err = rootNode.getAllLeavesOnChannel(
			leavesChannels.LeavesChan,
			keyBuilder,
			trieLeafParser,
			tr.chanClose,
			ctx,
			tr.TrieContext,
		)
		if err != nil {
			leavesChannels.ErrChan.WriteInChanNonBlocking(err)
			log.Error("could not get all trie leaves: ", "error", err)
		}

		tr.TrieContext.ExitPruningBufferingMode()

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
func (tr *patriciaMerkleTrie) GetProof(key []byte, rootHash []byte) ([][]byte, []byte, error) {
	if common.IsEmptyTrie(rootHash) {
		return nil, nil, ErrNilNode
	}

	rootNode, encodedNode, err := getNodeFromDBAndDecode(rootHash, tr.TrieContext)
	if err != nil {
		return nil, nil, fmt.Errorf("trie get proof error: %w", err)
	}

	var proof [][]byte
	var errGet error

	data := &nodeData{
		currentNode: rootNode,
		encodedNode: encodedNode,
		hexKey:      keyBuilder.KeyBytesToHex(key),
	}

	for {
		proof = append(proof, data.encodedNode)
		value := data.currentNode.getValue()

		data, errGet = data.currentNode.getNext(data.hexKey, tr.TrieContext)
		if errGet != nil {
			return nil, nil, errGet
		}
		if data == nil {
			return proof, value, nil
		}
	}
}

// VerifyProof verifies the given Merkle proof
func (tr *patriciaMerkleTrie) VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error) {
	ok, err := tr.verifyProof(rootHash, tr.TrieContext.Compute(string(key)), proof)
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
	key = keyBuilder.KeyBytesToHex(key)
	for _, encodedNode := range proof {
		if encodedNode == nil {
			return false, nil
		}

		hash := tr.TrieContext.Compute(string(encodedNode))
		if !bytes.Equal(wantHash, hash) {
			return false, nil
		}

		n, errDecode := decodeNode(encodedNode, tr.TrieContext)
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
	return tr.TrieContext.GetStorage()
}

// GetTrieStats will collect and return the statistics for the given rootHash
func (tr *patriciaMerkleTrie) GetTrieStats(address string, rootHash []byte) (common.TrieStatisticsHandler, error) {
	if common.IsEmptyTrie(rootHash) {
		return statistics.NewTrieStatistics(), nil
	}

	rootNode, rootBytes, err := getNodeFromDBAndDecode(rootHash, tr.TrieContext)
	if err != nil {
		return nil, err
	}

	ts := statistics.NewTrieStatistics()
	err = rootNode.collectStats(ts, rootDepthLevel, uint64(len(rootBytes)), tr.TrieContext)
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

	tr.updateTrieMutex.Lock()
	defer tr.updateTrieMutex.Unlock()

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

	_, err = rootNode.collectLeavesForMigration(args, keyBuilder.NewKeyBuilder(), tr.TrieContext)
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

// GetNodeDataFromHash returns the node data for the given hash
func GetNodeDataFromHash(hash []byte, keyBuilder common.KeyBuilder, db common.TrieStorageInteractor, msh marshal.Marshalizer, hsh hashing.Hasher) ([]common.TrieNodeData, error) {
	storageManager, ok := db.(common.StorageManager)
	if !ok {
		return nil, fmt.Errorf("invalid db type: %T", db)
	}

	n, _, err := getNodeFromDBAndDecode(
		hash,
		&trieContext{
			storageManager,
			msh,
			hsh,
		},
	)
	if err != nil {
		return nil, err
	}

	return n.getNodeData(keyBuilder)
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
