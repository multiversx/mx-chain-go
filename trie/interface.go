package trie

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type node interface {
	getHash() []byte
	setHash() error
	setGivenHash([]byte)
	setHashConcurrent(wg *sync.WaitGroup, c chan error)
	setRootHash() error
	getCollapsed() (node, error) // a collapsed node is a node that instead of the children holds the children hashes
	isCollapsed() bool
	isPosCollapsed(pos int) bool
	isDirty() bool
	getEncodedNode() ([]byte, error)
	resolveCollapsed(pos byte, tmc MetricsCollector, db common.TrieStorageInteractor) error
	hashNode() ([]byte, error)
	hashChildren() error
	tryGet(key []byte, tmc MetricsCollector, db common.TrieStorageInteractor) ([]byte, error)
	getNext(key []byte, tmc MetricsCollector, db common.TrieStorageInteractor) (node, []byte, error)
	insert(newData core.TrieData, tmc MetricsCollector, db common.TrieStorageInteractor) (node, [][]byte, error)
	delete(key []byte, tmc MetricsCollector, db common.TrieStorageInteractor) (bool, node, [][]byte, error)
	reduceNode(pos int, tmc MetricsCollector) (node, bool, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int, tmc MetricsCollector, db common.TrieStorageInteractor)
	getDirtyHashes(common.ModifiedHashes) error
	getChildren(tmc MetricsCollector, db common.TrieStorageInteractor) ([]node, error)
	isValid() bool
	getNodeData(common.KeyBuilder) ([]common.TrieNodeData, error)
	setDirty(bool)
	loadChildren(func([]byte) (node, error)) ([][]byte, []node, error)
	getAllLeavesOnChannel(chan core.KeyValueHolder, common.KeyBuilder, common.TrieLeafParser, common.TrieStorageInteractor, marshal.Marshalizer, chan struct{}, context.Context, MetricsCollector) error
	getAllHashes(tmc MetricsCollector, db common.TrieStorageInteractor) ([][]byte, error)
	getNextHashAndKey([]byte) (bool, []byte, []byte)
	getValue() []byte
	getVersion() (core.TrieNodeVersion, error)
	collectLeavesForMigration(migrationArgs vmcommon.ArgsMigrateDataTrieLeaves, tmc MetricsCollector, db common.TrieStorageInteractor, keyBuilder common.KeyBuilder) (bool, error)
	shouldCollapseChild([]byte, MetricsCollector) bool

	commitDirty(originDb common.TrieStorageInteractor, targetDb common.BaseStorer) error
	commitSnapshot(originDb common.TrieStorageInteractor, leavesChan chan core.KeyValueHolder, missingNodesChan chan []byte, ctx context.Context, stats common.TrieStatisticsHandler, idleProvider IdleNodeProvider, tmc MetricsCollector) error

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
	sizeInBytes() int
	collectStats(handler common.TrieStatisticsHandler, tmc MetricsCollector, db common.TrieStorageInteractor) error

	IsInterfaceNil() bool
}

type dbWithGetFromEpoch interface {
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
}

type snapshotNode interface {
	commitSnapshot(originDb common.TrieStorageInteractor, leavesChan chan core.KeyValueHolder, missingNodesChan chan []byte, ctx context.Context, stats common.TrieStatisticsHandler, idleProvider IdleNodeProvider, tmc MetricsCollector) error
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

// MetricsCollector is used to collect metrics about the trie
type MetricsCollector interface {
	SetDepth(depth uint32)
	GetMaxDepth() uint32
	AddSizeLoadedInMem(size int)
	GetSizeLoadedInMem() int
}
