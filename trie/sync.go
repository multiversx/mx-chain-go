package trie

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

type trieNodeInfo struct {
	trieNode node
	received bool
}

// TODO consider removing this implementation
type trieSyncer struct {
	baseSyncTrie
	rootFound                 bool
	shardId                   uint32
	topic                     string
	rootHash                  []byte
	nodesForTrie              map[string]trieNodeInfo
	waitTimeBetweenRequests   time.Duration
	marshalizer               marshal.Marshalizer
	hasher                    hashing.Hasher
	db                        common.TrieStorageInteractor
	requestHandler            RequestHandler
	interceptedNodesCacher    storage.Cacher
	mutOperation              sync.RWMutex
	handlerID                 string
	trieSyncStatistics        data.SyncStatisticsHandler
	timeoutHandler            TimeoutHandler
	maxHardCapForMissingNodes int
	leavesChan                chan core.KeyValueHolder
}

const maxNewMissingAddedPerTurn = 10

// ArgTrieSyncer is the argument for the trie syncer
type ArgTrieSyncer struct {
	Marshalizer               marshal.Marshalizer
	Hasher                    hashing.Hasher
	DB                        common.StorageManager
	RequestHandler            RequestHandler
	InterceptedNodes          storage.Cacher
	ShardId                   uint32
	Topic                     string
	TrieSyncStatistics        common.SizeSyncStatisticsHandler
	MaxHardCapForMissingNodes int
	CheckNodesOnDisk          bool
	TimeoutHandler            TimeoutHandler
	LeavesChan                chan core.KeyValueHolder
}

// NewTrieSyncer creates a new instance of trieSyncer
func NewTrieSyncer(arg ArgTrieSyncer) (*trieSyncer, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	stsm, err := NewSyncTrieStorageManager(arg.DB)
	if err != nil {
		return nil, err
	}

	ts := &trieSyncer{
		requestHandler:            arg.RequestHandler,
		interceptedNodesCacher:    arg.InterceptedNodes,
		db:                        stsm,
		marshalizer:               arg.Marshalizer,
		hasher:                    arg.Hasher,
		nodesForTrie:              make(map[string]trieNodeInfo),
		topic:                     arg.Topic,
		shardId:                   arg.ShardId,
		waitTimeBetweenRequests:   time.Second,
		handlerID:                 core.UniqueIdentifier(),
		trieSyncStatistics:        arg.TrieSyncStatistics,
		timeoutHandler:            arg.TimeoutHandler,
		maxHardCapForMissingNodes: arg.MaxHardCapForMissingNodes,
		leavesChan:                arg.LeavesChan,
	}

	return ts, nil
}

func checkArguments(arg ArgTrieSyncer) error {
	if check.IfNil(arg.RequestHandler) {
		return ErrNilRequestHandler
	}
	if check.IfNil(arg.InterceptedNodes) {
		return data.ErrNilCacher
	}
	if check.IfNil(arg.DB) {
		return fmt.Errorf("%w in NewTrieSyncer", ErrNilDatabase)
	}
	if check.IfNil(arg.Marshalizer) {
		return fmt.Errorf("%w in NewTrieSyncer", ErrNilMarshalizer)
	}
	if check.IfNil(arg.Hasher) {
		return fmt.Errorf("%w in NewTrieSyncer", ErrNilHasher)
	}
	if len(arg.Topic) == 0 {
		return ErrInvalidTrieTopic
	}
	if check.IfNil(arg.TrieSyncStatistics) {
		return ErrNilTrieSyncStatistics
	}
	if check.IfNil(arg.TimeoutHandler) {
		return ErrNilTimeoutHandler
	}
	if arg.MaxHardCapForMissingNodes < 1 {
		return fmt.Errorf("%w provided: %v", ErrInvalidMaxHardCapForMissingNodes, arg.MaxHardCapForMissingNodes)
	}

	return nil
}

// StartSyncing completes the trie, asking for missing trie nodes on the network
func (ts *trieSyncer) StartSyncing(rootHash []byte, ctx context.Context) error {
	if common.IsEmptyTrie(rootHash) {
		return nil
	}
	if ctx == nil {
		return ErrNilContext
	}

	ts.mutOperation.Lock()
	ts.nodesForTrie = make(map[string]trieNodeInfo)
	ts.nodesForTrie[string(rootHash)] = trieNodeInfo{received: false}
	ts.mutOperation.Unlock()

	ts.rootFound = false
	ts.rootHash = rootHash

	timeStart := time.Now()
	defer func() {
		ts.setSyncDuration(time.Since(timeStart))
	}()

	for {
		shouldRetryAfterRequest, err := ts.checkIfSynced()
		if err != nil {
			return err
		}

		numUnResolved := ts.requestNodes()
		if !shouldRetryAfterRequest && numUnResolved == 0 {
			return nil
		}

		select {
		case <-time.After(ts.waitTimeBetweenRequests):
			continue
		case <-ctx.Done():
			return core.ErrContextClosing
		}
	}
}

func (ts *trieSyncer) checkIfSynced() (bool, error) {
	var currentNode node
	var err error
	var nextNodes []node
	missingNodes := make(map[string]struct{})
	currentMissingNodes := make([][]byte, 0)
	checkedNodes := make(map[string]struct{})

	newElement := true
	shouldRetryAfterRequest := false

	ts.mutOperation.Lock()
	defer ts.mutOperation.Unlock()

	for newElement {
		newElement = false

		for nodeHash, nodeInfo := range ts.nodesForTrie {
			currentMissingNodes = currentMissingNodes[:0]
			if _, ok := checkedNodes[nodeHash]; ok {
				continue
			}

			currentNode = nodeInfo.trieNode
			if !nodeInfo.received {
				currentNode, err = ts.getNode([]byte(nodeHash))
				if err != nil {
					continue
				}
			}

			checkedNodes[nodeHash] = struct{}{}

			currentMissingNodes, nextNodes, err = currentNode.loadChildren(ts.getNode)
			if err != nil {
				return false, err
			}
			log.Trace("loaded children for node", "hash", currentNode.getHash())

			if len(currentMissingNodes) > 0 {
				hardcapReached := false
				for _, hash := range currentMissingNodes {
					missingNodes[string(hash)] = struct{}{}
					if len(missingNodes) > ts.maxHardCapForMissingNodes {
						hardcapReached = true
						break
					}
				}

				if hardcapReached {
					newElement = false
					break
				}

				nextNodes = append(nextNodes, currentNode)
				_ = ts.addNew(nextNodes)
				shouldRetryAfterRequest = true

				if len(missingNodes) > maxNewMissingAddedPerTurn {
					newElement = false
				}

				continue
			}

			nextNodes, err = currentNode.getChildren(ts.db)
			if err != nil {
				return false, err
			}

			tmpNewElement := ts.addNew(nextNodes)
			newElement = newElement || tmpNewElement

			delete(ts.nodesForTrie, nodeHash)

			var numBytes int
			numBytes, err = encodeNodeAndCommitToDB(currentNode, ts.db)
			if err != nil {
				return false, err
			}

			writeLeafNodeToChan(currentNode, ts.leavesChan)

			ts.timeoutHandler.ResetWatchdog()

			ts.updateStats(uint64(numBytes), currentNode)

			if !ts.rootFound && bytes.Equal([]byte(nodeHash), ts.rootHash) {
				ts.rootFound = true
			}
		}
	}

	if ts.timeoutHandler.IsTimeout() {
		return false, ErrTimeIsOut
	}

	for hash := range missingNodes {
		ts.nodesForTrie[hash] = trieNodeInfo{received: false}
	}

	return shouldRetryAfterRequest, nil
}

// adds new elements to needed hash map, lock ts.nodeHashesMutex before calling
func (ts *trieSyncer) addNew(nextNodes []node) bool {
	newElement := false
	for _, nextNode := range nextNodes {
		nextHash := string(nextNode.getHash())

		nodeInfo, ok := ts.nodesForTrie[nextHash]
		if !ok || !nodeInfo.received {
			newElement = true
			ts.trieSyncStatistics.AddNumProcessed(1)
			ts.nodesForTrie[nextHash] = trieNodeInfo{
				trieNode: nextNode,
				received: true,
			}
		}
	}

	return newElement
}

func (ts *trieSyncer) getNode(hash []byte) (node, error) {
	nodeInfo, ok := ts.nodesForTrie[string(hash)]
	if ok && nodeInfo.received {
		return nodeInfo.trieNode, nil
	}

	return getNodeFromCacheOrStorage(
		hash,
		ts.interceptedNodesCacher,
		ts.db,
		ts.marshalizer,
		ts.hasher,
	)
}

func getNodeFromCacheOrStorage(
	hash []byte,
	interceptedNodesCacher storage.Cacher,
	db common.TrieStorageInteractor,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (node, error) {
	n, err := getNodeFromCache(hash, interceptedNodesCacher, marshalizer, hasher)
	if err == nil {
		return n, nil
	}

	existingNode, err := getNodeFromDBAndDecode(hash, db, marshalizer, hasher)
	if err != nil {
		return nil, ErrNodeNotFound
	}
	err = existingNode.setHash()
	if err != nil {
		return nil, ErrNodeNotFound
	}

	return existingNode, nil
}

func getNodeFromCache(
	hash []byte,
	interceptedNodesCacher storage.Cacher,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (node, error) {
	n, ok := interceptedNodesCacher.Get(hash)
	if ok {
		interceptedNodesCacher.Remove(hash)
		return trieNode(n, marshalizer, hasher)
	}

	return nil, ErrNodeNotFound
}

func trieNode(
	data interface{},
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (node, error) {
	n, ok := data.(*InterceptedTrieNode)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	decodedNode, err := decodeNode(n.GetSerialized(), marshalizer, hasher)
	if err != nil {
		return nil, err
	}

	err = decodedNode.setHash()
	if err != nil {
		return nil, err
	}
	decodedNode.setDirty(true)

	return decodedNode, nil
}

func writeLeafNodeToChan(element node, ch chan core.KeyValueHolder) {
	if ch == nil {
		return
	}

	leafNodeElement, isLeaf := element.(*leafNode)
	if !isLeaf {
		return
	}

	trieLeaf := keyValStorage.NewKeyValStorage(leafNodeElement.Key, leafNodeElement.Value, core.TrieNodeVersion(leafNodeElement.Version))
	ch <- trieLeaf
}

func (ts *trieSyncer) requestNodes() uint32 {
	ts.mutOperation.RLock()
	numUnResolvedNodes := uint32(len(ts.nodesForTrie))
	hashes := make([][]byte, 0)
	for hash, nodeInfo := range ts.nodesForTrie {
		if !nodeInfo.received {
			hashes = append(hashes, []byte(hash))
		}
	}
	ts.trieSyncStatistics.SetNumMissing(ts.rootHash, len(hashes))
	if len(hashes) > 0 {
		ts.requestHandler.RequestTrieNodes(ts.shardId, hashes, ts.topic)
	}
	ts.mutOperation.RUnlock()

	return numUnResolvedNodes
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *trieSyncer) IsInterfaceNil() bool {
	return ts == nil
}
