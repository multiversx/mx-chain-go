package trie

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type deepFirstTrieSyncer struct {
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
	nodes                     *trieNodesHandler
	requestedHashes           map[string]*request
}

// NewDeepFirstTrieSyncer creates a new instance of trieSyncer that uses the deep first algorithm
func NewDeepFirstTrieSyncer(arg ArgTrieSyncer) (*deepFirstTrieSyncer, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	stsm, err := NewSyncTrieStorageManager(arg.DB)
	if err != nil {
		return nil, err
	}

	d := &deepFirstTrieSyncer{
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
	}

	return d, nil
}

// StartSyncing completes the trie, asking for missing trie nodes on the network. All concurrent calls will be serialized
// so this function is treated as a large critical section. This was done so the inner processing can be done without using
// other mutexes.
func (d *deepFirstTrieSyncer) StartSyncing(rootHash []byte, ctx context.Context) error {
	if len(rootHash) == 0 || bytes.Equal(rootHash, EmptyTrieHash) {
		return nil
	}
	if ctx == nil {
		return ErrNilContext
	}

	d.mutOperation.Lock()
	defer d.mutOperation.Unlock()

	d.nodes = newTrieNodesHandler()

	d.rootFound = false
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
			return errors.ErrContextClosing
		}
	}
}

func (d *deepFirstTrieSyncer) checkIsSyncedWhileProcessingMissingAndExisting() (bool, error) {
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

	if len(d.nodes.missingHashes) > 0 {
		marginSlice := make([][]byte, 0, maxNumRequestedNodesPerBatch)
		for _, hash := range d.nodes.hashesOrder {
			_, isMissing := d.nodes.missingHashes[hash]
			if !isMissing {
				continue
			}

			n, errGet := d.getNodeFromCache([]byte(hash))
			if errGet == nil {
				d.nodes.processMissingHashWasFound(n, hash)
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

	return d.nodes.jobDone(), nil
}

func (d *deepFirstTrieSyncer) request(hashes [][]byte) {
	d.requestHandler.RequestTrieNodes(d.shardId, hashes, d.topic)
	d.trieSyncStatistics.SetNumMissing(d.rootHash, len(d.nodes.missingHashes))
}

func (d *deepFirstTrieSyncer) processMissingAndExisting() error {
	d.processMissingHashes()

	for {
		processed, err := d.processFirstExistingNode()
		if err != nil {
			return err
		}

		if len(d.nodes.missingHashes) > d.maxHardCapForMissingNodes {
			break
		}
		if !processed {
			break
		}
	}

	return nil
}

func (d *deepFirstTrieSyncer) processMissingHashes() {
	for _, hash := range d.nodes.hashesOrder {
		_, isMissing := d.nodes.missingHashes[hash]
		if !isMissing {
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

func (d *deepFirstTrieSyncer) processFirstExistingNode() (bool, error) {
	for index, hash := range d.nodes.hashesOrder {
		element, isExisting := d.nodes.existingNodes[hash]
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

func (d *deepFirstTrieSyncer) storeTrieNode(element node) error {
	numBytes, err := encodeNodeAndCommitToDB(element, d.db)
	if err != nil {
		return err
	}
	d.timeoutHandler.ResetWatchdog()

	d.trieSyncStatistics.AddNumReceived(1)
	if numBytes > core.MaxBufferSizeToSendTrieNodes {
		d.trieSyncStatistics.AddNumLarge(1)
	}
	d.trieSyncStatistics.AddNumBytesReceived(uint64(numBytes))
	d.updateStats(uint64(numBytes), element)

	return nil
}

func (d *deepFirstTrieSyncer) storeLeaves(children []node) ([]node, error) {
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

func (d *deepFirstTrieSyncer) getNode(hash []byte) (node, error) {
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

func (d *deepFirstTrieSyncer) getNodeFromCache(hash []byte) (node, error) {
	return getNodeFromCache(
		hash,
		d.interceptedNodesCacher,
		d.marshalizer,
		d.hasher,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *deepFirstTrieSyncer) IsInterfaceNil() bool {
	return d == nil
}
