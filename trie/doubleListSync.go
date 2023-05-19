package trie

import (
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/storage"
)

// this value of 5 seconds yield better results in terms of network bandwidth & sync times
// whenever we need to change this, we should run comparative benchmarks
const deltaReRequest = 5 * int64(time.Second)
const maxNumRequestedNodesPerBatch = 1000

type request struct {
	timestamp int64
}

// TODO consider removing this implementation
type doubleListTrieSyncer struct {
	baseSyncTrie
	rootFound                 bool
	shardId                   uint32
	topic                     string
	rootHash                  []byte
	waitTimeBetweenChecks     time.Duration
	marshalizer               marshal.Marshalizer
	hasher                    hashing.Hasher
	db                        common.DBWriteCacher
	requestHandler            RequestHandler
	interceptedNodesCacher    storage.Cacher
	mutOperation              sync.RWMutex
	handlerID                 string
	trieSyncStatistics        common.SizeSyncStatisticsHandler
	timeoutHandler            TimeoutHandler
	maxHardCapForMissingNodes int
	checkNodesOnDisk          bool
	existingNodes             map[string]node
	missingHashes             map[string]struct{}
	requestedHashes           map[string]*request
	leavesChan                chan core.KeyValueHolder
}

// NewDoubleListTrieSyncer creates a new instance of trieSyncer that uses 2 list for keeping the "margin" nodes.
// One is used for keeping track of the loaded nodes (their children will need to be checked) and the other one that holds
// missing nodes
func NewDoubleListTrieSyncer(arg ArgTrieSyncer) (*doubleListTrieSyncer, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	stsm, err := NewSyncTrieStorageManager(arg.DB)
	if err != nil {
		return nil, err
	}

	d := &doubleListTrieSyncer{
		requestHandler:            arg.RequestHandler,
		interceptedNodesCacher:    arg.InterceptedNodes,
		db:                        stsm,
		marshalizer:               arg.Marshalizer,
		hasher:                    arg.Hasher,
		topic:                     arg.Topic,
		shardId:                   arg.ShardId,
		waitTimeBetweenChecks:     time.Millisecond * 100,
		handlerID:                 core.UniqueIdentifier(),
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
func (d *doubleListTrieSyncer) StartSyncing(rootHash []byte, ctx context.Context) error {
	if common.IsEmptyTrie(rootHash) {
		return nil
	}
	if ctx == nil {
		return ErrNilContext
	}

	d.mutOperation.Lock()
	defer d.mutOperation.Unlock()

	d.existingNodes = make(map[string]node)
	d.missingHashes = make(map[string]struct{})
	d.requestedHashes = make(map[string]*request)

	d.rootFound = false
	d.rootHash = rootHash

	d.missingHashes[string(rootHash)] = struct{}{}

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
			return errors.ErrContextClosing
		}
	}
}

func (d *doubleListTrieSyncer) checkIsSyncedWhileProcessingMissingAndExisting() (bool, error) {
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

	if len(d.missingHashes) > 0 {
		marginSlice := make([][]byte, 0, maxNumRequestedNodesPerBatch)
		for hash := range d.missingHashes {
			n, errGet := d.getNodeFromCache([]byte(hash))
			if errGet == nil {
				d.existingNodes[hash] = n
				delete(d.missingHashes, hash)
				delete(d.requestedHashes, hash)

				continue
			}

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
		return false, nil
	}

	return len(d.missingHashes)+len(d.existingNodes) == 0, nil
}

func (d *doubleListTrieSyncer) request(hashes [][]byte) {
	d.requestHandler.RequestTrieNodes(d.shardId, hashes, d.topic)
	d.trieSyncStatistics.SetNumMissing(d.rootHash, len(d.missingHashes))
}

func (d *doubleListTrieSyncer) processMissingAndExisting() error {
	d.processMissingHashes()

	return d.processExistingNodes()
}

func (d *doubleListTrieSyncer) processMissingHashes() {
	for hash := range d.missingHashes {
		n, err := d.getNodeFromCache([]byte(hash))
		if err != nil {
			continue
		}

		delete(d.missingHashes, hash)
		delete(d.requestedHashes, hash)

		d.existingNodes[string(n.getHash())] = n
	}
}

func (d *doubleListTrieSyncer) processExistingNodes() error {
	for hash, element := range d.existingNodes {
		numBytes, err := encodeNodeAndCommitToDB(element, d.db)
		if err != nil {
			return err
		}

		writeLeafNodeToChan(element, d.leavesChan)

		d.timeoutHandler.ResetWatchdog()

		var children []node
		var missingChildrenHashes [][]byte
		missingChildrenHashes, children, err = element.loadChildren(d.getNode)
		if err != nil {
			return err
		}

		d.trieSyncStatistics.AddNumProcessed(1)
		if numBytes > core.MaxBufferSizeToSendTrieNodes {
			d.trieSyncStatistics.AddNumLarge(1)
		}
		d.trieSyncStatistics.AddNumBytesReceived(uint64(numBytes))
		d.updateStats(uint64(numBytes), element)

		delete(d.existingNodes, hash)

		for _, child := range children {
			d.existingNodes[string(child.getHash())] = child
		}

		for _, missingHash := range missingChildrenHashes {
			d.missingHashes[string(missingHash)] = struct{}{}
		}

		if len(missingChildrenHashes) > 0 && len(d.missingHashes) > d.maxHardCapForMissingNodes {
			break
		}
	}

	return nil
}

func (d *doubleListTrieSyncer) getNode(hash []byte) (node, error) {
	if d.checkNodesOnDisk {
		return getNodeFromCacheOrStorage(
			hash,
			d.interceptedNodesCacher,
			d.db,
			d.marshalizer,
			d.hasher,
		)
	}
	return d.getNodeFromCache(hash)
}

func (d *doubleListTrieSyncer) getNodeFromCache(hash []byte) (node, error) {
	return getNodeFromCache(
		hash,
		d.interceptedNodesCacher,
		d.marshalizer,
		d.hasher,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *doubleListTrieSyncer) IsInterfaceNil() bool {
	return d == nil
}
