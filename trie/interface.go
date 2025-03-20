package trie

import (
	"context"
	"io"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type nodeData struct {
	currentNode node
	encodedNode []byte
	hexKey      []byte
}

type baseTrieNode interface {
	getHash() []byte
	setGivenHash([]byte)
	setDirty(bool)
	isDirty() bool
	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
}

type node interface {
	baseTrieNode
	setHash(goRoutinesManager common.TrieGoroutinesManager)
	getEncodedNode() ([]byte, error)
	tryGet(key []byte, depth uint32, db common.TrieStorageInteractor) ([]byte, uint32, error)
	getNext(key []byte, db common.TrieStorageInteractor) (*nodeData, error)
	insert(newData []core.TrieData, goRoutinesManager common.TrieGoroutinesManager, modifiedHashes common.AtomicBytesSlice, db common.TrieStorageInteractor) node
	delete(data []core.TrieData, goRoutinesManager common.TrieGoroutinesManager, modifiedHashes common.AtomicBytesSlice, db common.TrieStorageInteractor) (bool, node)
	reduceNode(pos int, db common.TrieStorageInteractor) (node, bool, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int, db common.TrieStorageInteractor)
	getChildren(db common.TrieStorageInteractor) ([]node, error)

	loadChildren(func([]byte) (node, error)) ([][]byte, []node, error)
	getAllLeavesOnChannel(chan core.KeyValueHolder, common.KeyBuilder, common.TrieLeafParser, common.TrieStorageInteractor, marshal.Marshalizer, chan struct{}, context.Context) error
	getNextHashAndKey([]byte) (bool, []byte, []byte)
	getValue() []byte
	getVersion() (core.TrieNodeVersion, error)
	collectLeavesForMigration(migrationArgs vmcommon.ArgsMigrateDataTrieLeaves, db common.TrieStorageInteractor, keyBuilder common.KeyBuilder) (bool, error)

	commitDirty(level byte, maxTrieLevelInMemory uint, goRoutinesManager common.TrieGoroutinesManager, hashesCollector common.TrieHashesCollector, originDb common.TrieStorageInteractor, targetDb common.BaseStorer)
	commitSnapshot(originDb common.TrieStorageInteractor, leavesChan chan core.KeyValueHolder, missingNodesChan chan []byte, ctx context.Context, stats common.TrieStatisticsHandler, idleProvider IdleNodeProvider, nodeBytes []byte, depthLevel int) error

	sizeInBytes() int
	collectStats(handler common.TrieStatisticsHandler, depthLevel int, nodeSize uint64, db common.TrieStorageInteractor) error

	IsInterfaceNil() bool
}

type dbWithGetFromEpoch interface {
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
}

type snapshotNode interface {
	commitSnapshot(originDb common.TrieStorageInteractor, leavesChan chan core.KeyValueHolder, missingNodesChan chan []byte, ctx context.Context, stats common.TrieStatisticsHandler, idleProvider IdleNodeProvider, nodeBytes []byte, depthLevel int) error
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
	SetNewRootNode(newRoot node)
	SetDataForRootChange(newRoot node, oldRootHash []byte, oldHashes [][]byte)
	ResetCollectedHashes()
	GetOldHashes() [][]byte
	GetOldRootHash() []byte
}
