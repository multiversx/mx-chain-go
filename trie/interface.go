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
	resolveCollapsed(pos byte, db common.TrieStorageInteractor) error
	hashNode() ([]byte, error)
	hashChildren() error
	tryGet(key []byte, depth uint32, db common.TrieStorageInteractor) ([]byte, uint32, error)
	getNext(key []byte, db common.TrieStorageInteractor) (node, []byte, error)
	insert(newData core.TrieData, db common.TrieStorageInteractor) (node, [][]byte, error)
	delete(key []byte, db common.TrieStorageInteractor) (bool, node, [][]byte, error)
	reduceNode(pos int) (node, bool, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int, db common.TrieStorageInteractor)
	getDirtyHashes(common.ModifiedHashes) error
	getChildren(db common.TrieStorageInteractor) ([]node, error)
	isValid() bool
	getNodeData(common.KeyBuilder) ([]common.TrieNodeData, error)
	setDirty(bool)
	loadChildren(func([]byte) (node, error)) ([][]byte, []node, error)
	getAllLeavesOnChannel(chan core.KeyValueHolder, common.KeyBuilder, common.TrieLeafParser, common.TrieStorageInteractor, marshal.Marshalizer, chan struct{}, context.Context) error
	getAllHashes(db common.TrieStorageInteractor) ([][]byte, error)
	getNextHashAndKey([]byte) (bool, []byte, []byte)
	getValue() []byte
	getVersion() (core.TrieNodeVersion, error)
	collectLeavesForMigration(migrationArgs vmcommon.ArgsMigrateDataTrieLeaves, db common.TrieStorageInteractor, keyBuilder common.KeyBuilder) (bool, error)

	commitDirty(level byte, maxTrieLevelInMemory uint, originDb common.TrieStorageInteractor, targetDb common.BaseStorer) error
	commitSnapshot(originDb common.TrieStorageInteractor, leavesChan chan core.KeyValueHolder, missingNodesChan chan []byte, ctx context.Context, stats common.TrieStatisticsHandler, idleProvider IdleNodeProvider, depthLevel int) error

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
	sizeInBytes() int
	collectStats(handler common.TrieStatisticsHandler, depthLevel int, db common.TrieStorageInteractor) error

	IsInterfaceNil() bool
}

type dbWithGetFromEpoch interface {
	GetFromEpoch(key []byte, epoch uint32) ([]byte, error)
}

type snapshotNode interface {
	commitSnapshot(originDb common.TrieStorageInteractor, leavesChan chan core.KeyValueHolder, missingNodesChan chan []byte, ctx context.Context, stats common.TrieStatisticsHandler, idleProvider IdleNodeProvider, depthLevel int) error
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
	GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, core.OptionalUint32, error)
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
