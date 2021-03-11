package trie

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ data.TrieSyncer = (*doubleListTrieSyncer)(nil)

type doubleListTrieSyncer struct {
	rootFound                 bool
	shardId                   uint32
	topic                     string
	rootHash                  []byte
	waitTimeBetweenChecks     time.Duration
	trie                      *patriciaMerkleTrie
	requestHandler            RequestHandler
	interceptedNodes          storage.Cacher
	mutOperation              sync.RWMutex
	handlerID                 string
	trieSyncStatistics        data.SyncStatisticsHandler
	lastSyncedTrieNode        time.Time
	timeoutBetweenCommits     time.Duration
	maxHardCapForMissingNodes int
	existingNodes             map[string]node
	missingHashes             map[string]struct{}
}

// NewDoubleListTrieSyncer creates a new instance of trieSyncer that uses 2 list for keeping the "margin" nodes.
// One is used for keeping track of the loaded nodes (their children will need to be checked) and the other one that holds
// missing nodes
func NewDoubleListTrieSyncer(arg ArgTrieSyncer) (*doubleListTrieSyncer, error) {
	err := checkArguments(arg)
	if err != nil {
		return nil, err
	}

	pmt, ok := arg.Trie.(*patriciaMerkleTrie)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	d := &doubleListTrieSyncer{
		requestHandler:            arg.RequestHandler,
		interceptedNodes:          arg.InterceptedNodes,
		trie:                      pmt,
		topic:                     arg.Topic,
		shardId:                   arg.ShardId,
		waitTimeBetweenChecks:     time.Millisecond * 100,
		handlerID:                 core.UniqueIdentifier(),
		trieSyncStatistics:        arg.TrieSyncStatistics,
		timeoutBetweenCommits:     arg.TimeoutBetweenTrieNodesCommits,
		maxHardCapForMissingNodes: arg.MaxHardCapForMissingNodes,
	}

	return d, nil
}

// StartSyncing completes the trie, asking for missing trie nodes on the network. All concurrent calls will be serialized
// so this function is treated as a large critical section. This was done so the inner processing can be done without using
// other mutexes.
func (d *doubleListTrieSyncer) StartSyncing(rootHash []byte, ctx context.Context) error {
	if len(rootHash) == 0 || bytes.Equal(rootHash, EmptyTrieHash) {
		return nil
	}
	if ctx == nil {
		return ErrNilContext
	}

	d.mutOperation.Lock()
	defer func() {
		d.mutOperation.Unlock()
	}()

	d.lastSyncedTrieNode = time.Now()
	d.existingNodes = make(map[string]node)
	d.missingHashes = make(map[string]struct{})

	d.rootFound = false
	d.rootHash = rootHash

	d.missingHashes[string(rootHash)] = struct{}{}

	for {
		isSynced, err := d.checkIsSyncedWhileProcessingMissingAndExisting()
		if err != nil {
			return err
		}
		if isSynced {
			return nil
		}

		select {
		case <-time.After(d.waitTimeBetweenChecks):
			continue
		case <-ctx.Done():
			return ErrContextClosing
		}
	}
}

func (d *doubleListTrieSyncer) checkIsSyncedWhileProcessingMissingAndExisting() (bool, error) {
	err := d.processMissingAndExisting()
	if err != nil {
		return false, err
	}

	if len(d.missingHashes) > 0 {
		marginSlice := make([][]byte, 0, len(d.missingHashes))
		for hash := range d.missingHashes {
			marginSlice = append(marginSlice, []byte(hash))
		}

		d.request(marginSlice)

		return false, nil
	}

	return len(d.missingHashes)+len(d.existingNodes) == 0, nil
}

func (d *doubleListTrieSyncer) request(hashes [][]byte) {
	d.requestHandler.RequestTrieNodes(d.shardId, hashes, d.topic)
	d.trieSyncStatistics.SetNumMissing(d.rootHash, len(hashes))
}

func (d *doubleListTrieSyncer) processMissingAndExisting() error {
	d.processMissingHashes()

	return d.processExistingNodes()
}

func (d *doubleListTrieSyncer) processMissingHashes() {
	for hash := range d.missingHashes {
		n, err := d.getNode([]byte(hash))
		if err != nil {
			continue
		}

		delete(d.missingHashes, hash)

		d.existingNodes[string(n.getHash())] = n
	}
}

func (d *doubleListTrieSyncer) processExistingNodes() error {
	for hash, element := range d.existingNodes {
		err := encodeNodeAndCommitToDB(element, d.trie.trieStorage.Database())
		if err != nil {
			return err
		}

		d.trieSyncStatistics.AddNumReceived(1)
		d.resetWatchdog()

		if !d.rootFound && bytes.Equal([]byte(hash), d.rootHash) {
			var collapsedRoot node
			collapsedRoot, err = element.getCollapsed()
			if err != nil {
				return nil
			}

			d.trie.root = collapsedRoot
		}

		var children []node
		var missingChildrenHashes [][]byte
		missingChildrenHashes, children, err = element.loadChildren(d.getNode)
		if err != nil {
			return err
		}

		if len(missingChildrenHashes) > 0 && len(d.missingHashes) > d.maxHardCapForMissingNodes {
			break
		}

		delete(d.existingNodes, hash)

		for _, child := range children {
			d.existingNodes[string(child.getHash())] = child
		}

		for _, missingHash := range missingChildrenHashes {
			d.missingHashes[string(missingHash)] = struct{}{}
		}
	}

	return nil
}

func (d *doubleListTrieSyncer) resetWatchdog() {
	d.lastSyncedTrieNode = time.Now()
}

// Trie returns the synced trie
func (d *doubleListTrieSyncer) Trie() data.Trie {
	return d.trie
}

func (d *doubleListTrieSyncer) getNode(hash []byte) (node, error) {
	n, ok := d.interceptedNodes.Get(hash)
	if ok {
		return trieNode(n)
	}

	existingNode, err := getNodeFromDBAndDecode(hash, d.trie.trieStorage.Database(), d.trie.marshalizer, d.trie.hasher)
	if err != nil {
		return nil, ErrNodeNotFound
	}
	err = existingNode.setHash()
	if err != nil {
		return nil, ErrNodeNotFound
	}

	return existingNode, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *doubleListTrieSyncer) IsInterfaceNil() bool {
	return d == nil
}
