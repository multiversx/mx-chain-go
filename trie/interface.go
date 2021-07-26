package trie

import (
	"io"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
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
	resolveCollapsed(pos byte, db temporary.DBWriteCacher) error
	hashNode() ([]byte, error)
	hashChildren() error
	tryGet(key []byte, db temporary.DBWriteCacher) ([]byte, error)
	getNext(key []byte, db temporary.DBWriteCacher) (node, []byte, error)
	insert(n *leafNode, db temporary.DBWriteCacher) (node, [][]byte, error)
	delete(key []byte, db temporary.DBWriteCacher) (bool, node, [][]byte, error)
	reduceNode(pos int) (node, bool, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int, db temporary.DBWriteCacher)
	getDirtyHashes(temporary.ModifiedHashes) error
	getChildren(db temporary.DBWriteCacher) ([]node, error)
	isValid() bool
	setDirty(bool)
	loadChildren(func([]byte) (node, error)) ([][]byte, []node, error)
	getAllLeavesOnChannel(chan core.KeyValueHolder, []byte, temporary.DBWriteCacher, marshal.Marshalizer, chan struct{}) error
	getAllHashes(db temporary.DBWriteCacher) ([][]byte, error)
	getNextHashAndKey([]byte) (bool, []byte, []byte)
	getNumNodes() temporary.NumNodesDTO

	commitDirty(level byte, maxTrieLevelInMemory uint, originDb temporary.DBWriteCacher, targetDb temporary.DBWriteCacher) error
	commitCheckpoint(originDb temporary.DBWriteCacher, targetDb temporary.DBWriteCacher, checkpointHashes temporary.CheckpointHashesHolder, leavesChan chan core.KeyValueHolder) error
	commitSnapshot(originDb temporary.DBWriteCacher, targetDb temporary.DBWriteCacher, leavesChan chan core.KeyValueHolder) error

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
	sizeInBytes() int

	IsInterfaceNil() bool
}

type snapshotNode interface {
	commitCheckpoint(originDb temporary.DBWriteCacher, targetDb temporary.DBWriteCacher, checkpointHashes temporary.CheckpointHashesHolder, leavesChan chan core.KeyValueHolder) error
	commitSnapshot(originDb temporary.DBWriteCacher, targetDb temporary.DBWriteCacher, leavesChan chan core.KeyValueHolder) error
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string)
	RequestInterval() time.Duration
	IsInterfaceNil() bool
}
