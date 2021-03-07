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
	marginExisting            map[string]node
	marginMissing             map[string]struct{}
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

	dlts := &doubleListTrieSyncer{
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

	return dlts, nil
}

// StartSyncing completes the trie, asking for missing trie nodes on the network. All concurrent calls will be serialized
// so this function is treated as a large critical section. This was done so the inner processing can be done without using
// other mutexes.
func (dlts *doubleListTrieSyncer) StartSyncing(rootHash []byte, ctx context.Context) error {
	if len(rootHash) == 0 || bytes.Equal(rootHash, EmptyTrieHash) {
		return nil
	}
	if ctx == nil {
		return ErrNilContext
	}

	dlts.mutOperation.Lock()
	defer func() {
		dlts.mutOperation.Unlock()
	}()

	dlts.lastSyncedTrieNode = time.Now()
	dlts.marginExisting = make(map[string]node)
	dlts.marginMissing = make(map[string]struct{})

	dlts.rootFound = false
	dlts.rootHash = rootHash

	dlts.marginMissing[string(rootHash)] = struct{}{}

	for {
		isSynced, err := dlts.checkIsSynced()
		if err != nil {
			return err
		}
		if isSynced {
			return nil
		}

		select {
		case <-time.After(dlts.waitTimeBetweenChecks):
			continue
		case <-ctx.Done():
			return ErrContextClosing
		}
	}
}

func (dlts *doubleListTrieSyncer) checkIsSynced() (bool, error) {
	err := dlts.processMargin()
	if err != nil {
		return false, err
	}

	if len(dlts.marginMissing) > 0 {
		marginSlice := make([][]byte, 0, len(dlts.marginMissing))
		for hash := range dlts.marginMissing {
			marginSlice = append(marginSlice, []byte(hash))
		}

		dlts.request(marginSlice)

		return false, nil
	}

	return len(dlts.marginMissing)+len(dlts.marginExisting) == 0, nil
}

func (dlts *doubleListTrieSyncer) request(hashes [][]byte) {
	dlts.requestHandler.RequestTrieNodes(dlts.shardId, hashes, dlts.topic)
	dlts.trieSyncStatistics.SetNumMissing(dlts.rootHash, len(hashes))
}

func (dlts *doubleListTrieSyncer) processMargin() error {
	dlts.processMissingMargin()

	return dlts.processNodesMargin()
}

func (dlts *doubleListTrieSyncer) processMissingMargin() {
	for hash := range dlts.marginMissing {
		n, err := dlts.getNode([]byte(hash))
		if err != nil {
			continue
		}

		delete(dlts.marginMissing, hash)

		dlts.marginExisting[string(n.getHash())] = n
	}
}

func (dlts *doubleListTrieSyncer) processNodesMargin() error {
	for hash, element := range dlts.marginExisting {
		err := encodeNodeAndCommitToDB(element, dlts.trie.trieStorage.Database())
		if err != nil {
			return err
		}

		dlts.trieSyncStatistics.AddNumReceived(1)
		dlts.resetWatchdog()

		if !dlts.rootFound && bytes.Equal([]byte(hash), dlts.rootHash) {
			var collapsedRoot node
			collapsedRoot, err = element.getCollapsed()
			if err != nil {
				return nil
			}

			dlts.trie.root = collapsedRoot
		}

		var children []node
		var missingChildrenHashes [][]byte
		missingChildrenHashes, children, err = element.loadChildren(dlts.getNode)
		if err != nil {
			return err
		}

		if len(missingChildrenHashes) > 0 && len(dlts.marginMissing) > dlts.maxHardCapForMissingNodes {
			break
		}

		delete(dlts.marginExisting, hash)

		for _, child := range children {
			dlts.marginExisting[string(child.getHash())] = child
		}

		for _, missingHash := range missingChildrenHashes {
			dlts.marginMissing[string(missingHash)] = struct{}{}
		}
	}

	return nil
}

func (dlts *doubleListTrieSyncer) resetWatchdog() {
	dlts.lastSyncedTrieNode = time.Now()
}

// Trie returns the synced trie
func (dlts *doubleListTrieSyncer) Trie() data.Trie {
	return dlts.trie
}

func (dlts *doubleListTrieSyncer) getNode(hash []byte) (node, error) {
	n, ok := dlts.interceptedNodes.Get(hash)
	if ok {
		return trieNode(n)
	}

	existingNode, err := getNodeFromDBAndDecode(hash, dlts.trie.trieStorage.Database(), dlts.trie.marshalizer, dlts.trie.hasher)
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
func (dlts *doubleListTrieSyncer) IsInterfaceNil() bool {
	return dlts == nil
}
