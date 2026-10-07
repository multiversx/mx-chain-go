package trie

import (
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

// TODO print the size of these maps/array by including the values in the trieSyncStatistics
type depthFirstTrieSyncer struct {
	baseSyncTrie
	shardId                   uint32
	topic                     string
	rootHash                  []byte
	waitTimeBetweenChecks     time.Duration
	marshaller                marshal.Marshalizer
	hasher                    hashing.Hasher
	db                        common.TrieStorageInteractor
	requestHandler            RequestHandler
	interceptedNodesCacher    storage.Cacher
	mutOperation              sync.RWMutex
	trieSyncStatistics        common.SizeSyncStatisticsHandler
	timeoutHandler            TimeoutHandler
	maxHardCapForMissingNodes int
	checkNodesOnDisk          bool
	nodes                     *trieNodesHandler
	requestedHashes           map[string]*request
	leavesChan                chan core.KeyValueHolder
}

// NewDepthFirstTrieSyncer creates a new instance of trieSyncer that uses the depth-first algorithm
func NewDepthFirstTrieSyncer(arg ArgTrieSyncer) (*depthFirstTrieSyncer, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	stsm, err := NewSyncTrieStorageManager(arg.DB)
	if err != nil {
		return nil, err
	}

	d := &depthFirstTrieSyncer{
		requestHandler:            arg.RequestHandler,
		interceptedNodesCacher:    arg.InterceptedNodes,
		db:                        stsm,
		marshaller:                arg.Marshalizer,
		hasher:                    arg.Hasher,
		topic:                     arg.Topic,
		shardId:                   arg.ShardId,
		waitTimeBetweenChecks:     time.Millisecond * 100,
		trieSyncStatistics:        arg.TrieSyncStatistics,
		timeoutHandler:            arg.TimeoutHandler,
		maxHardCapForMissingNodes: arg.MaxHardCapForMissingNodes,
		checkNodesOnDisk:          arg.CheckNodesOnDisk,
		leavesChan:                arg.LeavesChan,
	}

	return d, nil
}

// StartSyncing completes the trie, asking for missing trie nodes on the network. All concurrent calls will be serialized
// so this function is treated as a large critical section. This was done so the inner processing can be done without using
// other mutexes.
func (d *depthFirstTrieSyncer) StartSyncing(rootHash []byte, ctx context.Context) error {
	if common.IsEmptyTrie(rootHash) {
		return nil
	}
	if ctx == nil {
		return ErrNilContext
	}

	d.mutOperation.Lock()
	defer d.mutOperation.Unlock()

	d.nodes = newTrieNodesHandler()

	d.rootHash = rootHash

	d.nodes.addInitialRootHash(string(rootHash))
	d.requestedHashes = make(map[string]*request)

	timeStart := time.Now()
	defer func() {
		d.setSyncDuration(time.Since(timeStart))
	}()

	for {
		isSynced, err := d.checkIsSyncedWhileProcessingMissingAndExisting()
		if err != nil {
			return err
		}
		if isSynced {
			d.trieSyncStatistics.SetNumMissing(d.rootHash, 0)
			return nil
		}

		select {
		case <-time.After(d.waitTimeBetweenChecks):
			continue
		case <-ctx.Done():
			return core.ErrContextClosing
		}
	}
}

func (d *depthFirstTrieSyncer) checkIsSyncedWhileProcessingMissingAndExisting() (bool, error) {
	if d.timeoutHandler.IsTimeout() {
		return false, ErrTrieSyncTimeout
	}

	start := time.Now()
	defer func() {
		d.trieSyncStatistics.AddProcessingTime(time.Since(start))
		d.trieSyncStatistics.IncrementIteration()
	}()

	err := d.processMissingAndExisting()
	if err != nil {
		return false, err
	}

	d.requestMissingNodes()

	return d.nodes.jobDone(), nil
}

func (d *depthFirstTrieSyncer) requestMissingNodes() {
	if d.nodes.noMissingHashes() {
		return
	}

	marginSlice := make([][]byte, 0, maxNumRequestedNodesPerBatch)
	for _, hash := range d.nodes.hashesOrder {
		if !d.nodes.hashIsMissing(hash) {
			continue
		}

		n, errGet := d.getNodeFromCache([]byte(hash))
		if errGet == nil {
			d.nodes.processMissingHashWasFound(n, hash)
			delete(d.requestedHashes, hash)

			continue
		}

		// TODO remove this and instead use the mechanism provided by the request handler
		r, ok := d.requestedHashes[hash]
		if !ok {
			marginSlice = append(marginSlice, []byte(hash))
			d.requestedHashes[hash] = &request{
				timestamp: time.Now().UnixNano(),
			}
		} else {
			delta := time.Now().UnixNano() - r.timestamp
			if delta > deltaReRequest {
				marginSlice = append(marginSlice, []byte(hash))
				r.timestamp = time.Now().UnixNano()
			}
		}
	}

	d.request(marginSlice)
}

func (d *depthFirstTrieSyncer) request(hashes [][]byte) {
	d.requestHandler.RequestTrieNodes(d.shardId, hashes, d.topic)
	d.trieSyncStatistics.SetNumMissing(d.rootHash, len(d.nodes.missingHashes))
}

func (d *depthFirstTrieSyncer) processMissingAndExisting() error {
	d.processMissingHashes()

	for {
		processed, err := d.processFirstExistingNode()
		if err != nil {
			return err
		}
		if !processed {
			break
		}
		if len(d.nodes.missingHashes) > d.maxHardCapForMissingNodes {
			break
		}
	}

	return nil
}

func (d *depthFirstTrieSyncer) processMissingHashes() {
	for _, hash := range d.nodes.hashesOrder {
		if !d.nodes.hashIsMissing(hash) {
			continue
		}

		n, err := d.getNodeFromCache([]byte(hash))
		if err != nil {
			continue
		}

		d.nodes.processMissingHashWasFound(n, hash)
		delete(d.requestedHashes, hash)
	}
}

func (d *depthFirstTrieSyncer) processFirstExistingNode() (bool, error) {
	for index, hash := range d.nodes.hashesOrder {
		element, isExisting := d.nodes.getExistingNode(hash)
		if !isExisting {
			continue
		}

		err := d.storeTrieNode(element)
		if err != nil {
			return false, err
		}

		missingChildrenHashes, children, err := element.loadChildren(d.getNode)
		if err != nil {
			return false, err
		}

		childrenNotLeaves, err := d.storeLeaves(children)
		if err != nil {
			return false, err
		}

		d.nodes.replaceParentWithChildren(index, hash, childrenNotLeaves, missingChildrenHashes)

		return true, nil
	}

	return false, nil
}

func (d *depthFirstTrieSyncer) storeTrieNode(element node) error {
	numBytes, err := encodeNodeAndCommitToDB(element, d.db)
	if err != nil {
		return err
	}
	d.timeoutHandler.ResetWatchdog()

	d.trieSyncStatistics.AddNumProcessed(1)
	if numBytes > core.MaxBufferSizeToSendTrieNodes {
		d.trieSyncStatistics.AddNumLarge(1)
	}
	d.trieSyncStatistics.AddNumBytesReceived(uint64(numBytes))
	d.updateStats(uint64(numBytes), element)

	writeLeafNodeToChan(element, d.leavesChan)

	return nil
}

func (d *depthFirstTrieSyncer) storeLeaves(children []node) ([]node, error) {
	childrenNotLeaves := make([]node, 0, len(children))
	for _, element := range children {
		_, isLeaf := element.(*leafNode)
		if !isLeaf {
			childrenNotLeaves = append(childrenNotLeaves, element)
			continue
		}

		err := d.storeTrieNode(element)
		if err != nil {
			return nil, err
		}
	}

	return childrenNotLeaves, nil
}

func (d *depthFirstTrieSyncer) getNode(hash []byte) (node, error) {
	if d.checkNodesOnDisk {
		return getNodeFromCacheOrStorage(
			hash,
			d.interceptedNodesCacher,
			d.db,
			d.marshaller,
			d.hasher,
		)
	}
	return d.getNodeFromCache(hash)
}

func (d *depthFirstTrieSyncer) getNodeFromCache(hash []byte) (node, error) {
	return getNodeFromCache(
		hash,
		d.interceptedNodesCacher,
		d.marshaller,
		d.hasher,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *depthFirstTrieSyncer) IsInterfaceNil() bool {
	return d == nil
}
