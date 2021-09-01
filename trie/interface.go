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

	commitDirty(level byte, maxTrieLevelInMemory uint, originDb common.DBWriteCacher, targetDb common.DBWriteCacher) error
	commitCheckpoint(originDb common.DBWriteCacher, targetDb common.DBWriteCacher, checkpointHashes CheckpointHashesHolder, leavesChan chan core.KeyValueHolder, ctx context.Context) error
	commitSnapshot(originDb common.DBWriteCacher, leavesChan chan core.KeyValueHolder, ctx context.Context) error

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
	sizeInBytes() int

	IsInterfaceNil() bool
}

type snapshotNode interface {
	commitCheckpoint(originDb common.DBWriteCacher, targetDb common.DBWriteCacher, checkpointHashes CheckpointHashesHolder, leavesChan chan core.KeyValueHolder, ctx context.Context) error
	commitSnapshot(originDb common.DBWriteCacher, leavesChan chan core.KeyValueHolder, ctx context.Context) error
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
