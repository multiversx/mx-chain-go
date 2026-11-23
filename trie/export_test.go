package trie

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
)

func (ts *trieSyncer) trieNodeIntercepted(hash []byte, val interface{}) {
	ts.mutOperation.Lock()
	defer ts.mutOperation.Unlock()

	log.Trace("trie node intercepted", "hash", hash)

	n, ok := ts.nodesForTrie[string(hash)]
	if !ok || n.received {
		return
	}

	interceptedNode, err := trieNode(val, marshalizer, hasherMock)
	if err != nil {
		return
	}

	ts.nodesForTrie[string(hash)] = trieNodeInfo{
		trieNode: interceptedNode,
		received: true,
	}
}

// WaitForOperationToComplete -
func WaitForOperationToComplete(tsm common.StorageManager) {
	for tsm.IsPruningBlocked() {
		time.Sleep(10 * time.Millisecond)
	}
}

// CreateSmallTestTrieAndStorageManager -
func CreateSmallTestTrieAndStorageManager() (*patriciaMerkleTrie, *trieStorageManager) {
	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()

	return tr, trieStorage
}

// GetDirtyHashes -
func GetDirtyHashes(tr common.Trie) common.ModifiedHashes {
	hashes, _ := tr.GetAllHashes()
	dirtyHashes := make(common.ModifiedHashes)
	for _, hash := range hashes {
		dirtyHashes[string(hash)] = struct{}{}
	}

	return dirtyHashes
}

// WriteInChanNonBlocking -
func WriteInChanNonBlocking(errChan chan error, err error) {
	select {
	case errChan <- err:
	default:
	}
}

type StorageManagerExtensionStub struct {
	*storageManager.StorageManagerStub
}

// IsBaseTrieStorageManager -
func IsBaseTrieStorageManager(tsm common.StorageManager) bool {
	_, ok := tsm.(*trieStorageManager)
	return ok
}

// IsTrieStorageManagerInEpoch -
func IsTrieStorageManagerInEpoch(tsm common.StorageManager) bool {
	_, ok := tsm.(*trieStorageManagerInEpoch)
	return ok
}

// NewBaseIterator -
func NewBaseIterator(trie common.Trie) (*baseIterator, error) {
	return newBaseIterator(trie)
}

// GetDefaultTrieStorageManagerParameters -
func GetDefaultTrieStorageManagerParameters() NewTrieStorageManagerArgs {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:           1000,
		SnapshotsBufferLen:         10,
		SnapshotsGoroutinesPerCore: 1,
	}

	return NewTrieStorageManagerArgs{
		MainStorer:     testscommon.NewSnapshotPruningStorerMock(),
		Marshalizer:    &marshal.GogoProtoMarshalizer{},
		Hasher:         &testscommon.KeccakMock{},
		GeneralConfig:  generalCfg,
		IdleProvider:   &testscommon.ProcessStatusHandlerStub{},
		Identifier:     dataRetriever.UserAccountsUnit.String(),
		StatsCollector: statistics.NewStateStatistics(),
	}
}

// SnapshotsGoroutineNum -
func SnapshotsGoroutineNum(perCore uint32) int32 {
	return snapshotsGoroutineNum(perCore)
}

// GetNumCollapsedNodes returns the number of collapsed nodes in the trie
func (tr *patriciaMerkleTrie) GetNumCollapsedNodes() (int, error) {
	count := 0

	nextNodes := make([]node, 0)
	nextNodes = append(nextNodes, tr.root)
	for len(nextNodes) > 0 {
		currentNode := nextNodes[0]
		switch current := currentNode.(type) {
		case *branchNode:
			for i := range current.children {
				if current.children[i] != nil {
					nextNodes = append(nextNodes, current.children[i])
				} else if len(current.EncodedChildren[i]) != 0 {
					count++
				}
			}
		case *extensionNode:
			if current.child != nil {
				nextNodes = append(nextNodes, current.child)
			}
		case *leafNode:
			// do nothing
		default:
			return 0, ErrInvalidNode
		}
		nextNodes = nextNodes[1:]
	}
	return count, nil
}

type PatriciaMerkleTrieForTests = patriciaMerkleTrie

func (tr *patriciaMerkleTrie) RootNodeForTests() node {
	return tr.root
}

func (tr *patriciaMerkleTrie) RootIsCollapsedForTests() bool {
	return tr.root.isCollapsed()
}

func (tr *patriciaMerkleTrie) RootSizeForTests() int {
	return tr.root.sizeInBytes()
}

// ResidentSizeForTests measures the decoded nodes independently of the collapse manager.
func (tr *patriciaMerkleTrie) ResidentSizeForTests() int {
	return residentTrieNodeSize(tr.root)
}

func (tr *patriciaMerkleTrie) CollapseManagerForTests() common.TrieCollapseManager {
	return tr.collapseManager
}

func residentTrieNodeSize(current node) int {
	if current == nil {
		return 0
	}
	size := current.sizeInBytes()
	switch typed := current.(type) {
	case *branchNode:
		for _, child := range typed.children {
			size += residentTrieNodeSize(child)
		}
	case *extensionNode:
		size += residentTrieNodeSize(typed.child)
	}
	return size
}

type TrieReadFailureStore struct {
	common.BaseStorer
	FailKey []byte
}

func (s *TrieReadFailureStore) Get(key []byte) ([]byte, error) {
	if s.FailKey != nil && bytes.Equal(s.FailKey, key) {
		s.FailKey = nil
		return nil, errors.New("transient delete read failure")
	}
	return s.BaseStorer.Get(key)
}

func NewTrieForSizeAccounting(t *testing.T, budget uint64) *patriciaMerkleTrie {
	t.Helper()
	args := GetDefaultTrieStorageManagerParameters()
	storage, err := NewTrieStorageManager(args)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	flags.AddActiveFlags(common.AutoBalanceDataTriesFlag, common.FixDataTrieLeafValidationFlag)
	manager, err := collapseManager.NewCollapseManager(budget, 100)
	require.NoError(t, err)
	tr, err := NewTrie(storage, args.Marshalizer, args.Hasher, flags, manager)
	require.NoError(t, err)
	return tr
}
