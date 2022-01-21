package trie

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
	resolveCollapsed(pos byte, db common.DBWriteCacher) error
	hashNode() ([]byte, error)
	hashChildren() error
	tryGet(key []byte, db common.DBWriteCacher) ([]byte, error)
	getNext(key []byte, db common.DBWriteCacher) (node, []byte, error)
	insert(n *leafNode, db common.DBWriteCacher) (node, [][]byte, error)
	delete(key []byte, db common.DBWriteCacher) (bool, node, [][]byte, error)
	reduceNode(pos int) (node, bool, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int, db common.DBWriteCacher)
	getDirtyHashes(common.ModifiedHashes) error
	getChildren(db common.DBWriteCacher) ([]node, error)
	isValid() bool
	setDirty(bool)
	loadChildren(func([]byte) (node, error)) ([][]byte, []node, error)
	getAllLeavesOnChannel(chan core.KeyValueHolder, []byte, common.DBWriteCacher, marshal.Marshalizer, chan struct{}) error
	getAllHashes(db common.DBWriteCacher) ([][]byte, error)
	getNextHashAndKey([]byte) (bool, []byte, []byte)
	getNumNodes() common.NumNodesDTO
	getValue() []byte

	commitDirty(level byte, maxTrieLevelInMemory uint, originDb common.DBWriteCacher, targetDb common.DBWriteCacher) error
	commitCheckpoint(originDb common.DBWriteCacher, targetDb common.DBWriteCacher, checkpointHashes CheckpointHashesHolder, leavesChan chan core.KeyValueHolder, ctx context.Context, stats common.SnapshotStatisticsHandler) error
	commitSnapshot(originDb common.DBWriteCacher, leavesChan chan core.KeyValueHolder, ctx context.Context, stats common.SnapshotStatisticsHandler) error

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
	sizeInBytes() int

	IsInterfaceNil() bool
}

type snapshotNode interface {
	commitCheckpoint(originDb common.DBWriteCacher, targetDb common.DBWriteCacher, checkpointHashes CheckpointHashesHolder, leavesChan chan core.KeyValueHolder, ctx context.Context, stats common.SnapshotStatisticsHandler) error
	commitSnapshot(originDb common.DBWriteCacher, leavesChan chan core.KeyValueHolder, ctx context.Context, stats common.SnapshotStatisticsHandler) error
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string)
	RequestInterval() time.Duration
	IsInterfaceNil() bool
}

// CheckpointHashesHolder is used to hold the hashes that need to be committed in the future state checkpoint
type CheckpointHashesHolder interface {
	Put(rootHash []byte, hashes common.ModifiedHashes) bool
	RemoveCommitted(lastCommittedRootHash []byte)
	Remove(hash []byte)
	ShouldCommit(hash []byte) bool
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
	common.DBWriteCacher
	GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, error)
	GetFromLastEpoch(key []byte) ([]byte, error)
	PutInEpochWithoutCache(key []byte, data []byte, epoch uint32) error
	GetLatestStorageEpoch() (uint32, error)
	GetFromCurrentEpoch(key []byte) ([]byte, error)
}

// EpochNotifier can notify upon an epoch change and provide the current epoch
type EpochNotifier interface {
	RegisterNotifyHandler(handler vmcommon.EpochSubscriberHandler)
	IsInterfaceNil() bool
}
