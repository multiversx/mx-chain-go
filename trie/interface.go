package trie

import (
	"context"
	"io"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// nodeData is used when computing merkle proofs. It is used as a DTO to avoid multiple storage accesses / serializations
type nodeData struct {
	currentNode node
	encodedNode []byte
	hexKey      []byte
}

// nodeWithHash is used as a DTO to avoid multiple hashing / serialization operations.
// It is used to store the hash of a node along with the node itself
type nodeWithHash struct {
	node node
	hash []byte
}

// keyData is used when traversing the trie to reach a certain leaf.
// At first, keyReminder contains the full key needed to reach the leaf node, and the path key is empty.
// For each trie node that is traversed, the keyData changes. The path for the already traversed nodes will be
// subtracted from the keyRemainder, and it will be added to the pathKey. So at all points during the traversal,
// pathKey + keyReminder = original key.
type keyData struct {
	keyRemainder []byte // remaining part of a key
	pathKey      []byte // path traversed in the trie
}

type node interface {
	snapshotNode
	setDirty(bool)
	isDirty() bool
	getEncodedNode(trieCtx common.TrieContext) ([]byte, error)
	tryGet(keyData *keyData, depth uint32, trieCtx common.TrieContext) ([]byte, uint32, error)
	getNext(key []byte, trieCtx common.TrieContext) (*nodeData, error)
	insert(pathKey common.KeyBuilder, newData []core.TrieData, goRoutinesManager common.TrieGoroutinesManager, modifiedHashes common.AtomicBytesSlice, trieCtx common.TrieContext) node
	delete(pathKey common.KeyBuilder, data []core.TrieData, goRoutinesManager common.TrieGoroutinesManager, modifiedHashes common.AtomicBytesSlice, trieCtx common.TrieContext) (bool, node)
	reduceNode(pos int, mutexKey string, trieCtx common.TrieContext) (node, bool, error)
	print(writer io.Writer, index int, trieCtx common.TrieContext)
	getChildren(trieCtx common.TrieContext) ([]nodeWithHash, error)
	getNodeData(common.KeyBuilder) ([]common.TrieNodeData, error)
	loadChildren(func([]byte) (node, error)) ([][]byte, []nodeWithHash, error)
	getAllLeavesOnChannel(chan core.KeyValueHolder, common.KeyBuilder, common.TrieLeafParser, chan struct{}, context.Context, common.TrieContext) error
	getNextHashAndKey([]byte) (bool, []byte, []byte)
	getValue() []byte
	getVersion() (core.TrieNodeVersion, error)
	collectLeavesForMigration(migrationArgs vmcommon.ArgsMigrateDataTrieLeaves, keyBuilder common.KeyBuilder, trieCtx common.TrieContext) (bool, error)

	commitDirty(
		level byte,
		pathKey common.KeyBuilder,
		maxTrieLevelInMemory uint,
		goRoutinesManager common.TrieGoroutinesManager,
		hashesCollector common.TrieHashesCollector,
		trieCtx common.TrieContext,
	)

	sizeInBytes() int
	collectStats(handler common.TrieStatisticsHandler, depthLevel int, nodeSize uint64, trieCtx common.TrieContext) error

	IsInterfaceNil() bool
}

type dbWithGetFromEpoch interface {
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
}

type snapshotNode interface {
	commitSnapshot(
		trieCtx common.TrieContext,
		leavesChan chan core.KeyValueHolder,
		missingNodesChan chan []byte,
		ctx context.Context,
		stats common.TrieStatisticsHandler,
		idleProvider IdleNodeProvider,
		nodeBytes []byte,
		depthLevel int,
	) error
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string)
	RequestInterval() time.Duration
	IsInterfaceNil() bool
}

// TimeoutHandler is able to tell if a timeout has occurred
type TimeoutHandler interface {
	ResetWatchdog()
	IsTimeout() bool
	IsInterfaceNil() bool
}

// epochStorer is used for storers that have information stored by epochs
type epochStorer interface {
	SetEpochForPutOperation(epoch uint32)
}

type snapshotPruningStorer interface {
	common.BaseStorer
	GetFromOldEpochsWithoutAddingToCache(key []byte, maxEpochToSearchFrom uint32) ([]byte, core.OptionalUint32, error)
	GetFromLastEpoch(key []byte) ([]byte, error)
	PutInEpoch(key []byte, data []byte, epoch uint32) error
	PutInEpochWithoutCache(key []byte, data []byte, epoch uint32) error
	GetLatestStorageEpoch() (uint32, error)
	GetFromCurrentEpoch(key []byte) ([]byte, error)
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
	RemoveFromCurrentEpoch(key []byte) error
	RemoveFromAllActiveEpochs(key []byte) error
}

// EpochNotifier can notify upon an epoch change and provide the current epoch
type EpochNotifier interface {
	RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler)
	IsInterfaceNil() bool
}

// IdleNodeProvider can determine if the node is idle or not
type IdleNodeProvider interface {
	IsIdle() bool
	IsInterfaceNil() bool
}

// RootManager is used to manage the root node and hashes related to it
type RootManager interface {
	GetRootNode() node
	GetRootHash() []byte
	SetDataForRootChange(rootData RootData)
	GetRootData() RootData
	ResetCollectedHashes()
	GetOldHashes() [][]byte
	GetOldRootHash() []byte
}
