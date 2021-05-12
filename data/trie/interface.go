package trie

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
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
	resolveCollapsed(pos byte, db data.DBWriteCacher) error
	hashNode() ([]byte, error)
	hashChildren() error
	tryGet(key []byte, db data.DBWriteCacher) ([]byte, error)
	getNext(key []byte, db data.DBWriteCacher) (node, []byte, error)
	insert(n *leafNode, db data.DBWriteCacher) (node, [][]byte, error)
	delete(key []byte, db data.DBWriteCacher) (bool, node, [][]byte, error)
	reduceNode(pos int) (node, bool, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int, db data.DBWriteCacher)
	getDirtyHashes(data.ModifiedHashes) error
	getChildren(db data.DBWriteCacher) ([]node, error)
	isValid() bool
	setDirty(bool)
	loadChildren(func([]byte) (node, error)) ([][]byte, []node, error)
	getAllLeavesOnChannel(chan core.KeyValueHolder, []byte, data.DBWriteCacher, marshal.Marshalizer, context.Context) error
	getAllHashes(db data.DBWriteCacher) ([][]byte, error)
	getNextHashAndKey([]byte) (bool, []byte, []byte)
	getNumNodes() data.NumNodesDTO

	commitDirty(level byte, maxTrieLevelInMemory uint, originDb data.DBWriteCacher, targetDb data.DBWriteCacher) error
	commitCheckpoint(originDb data.DBWriteCacher, targetDb data.DBWriteCacher) error
	commitSnapshot(originDb data.DBWriteCacher, targetDb data.DBWriteCacher) error

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
	sizeInBytes() int

	IsInterfaceNil() bool
}

type atomicBuffer interface {
	add(rootHash []byte)
	removeAll() [][]byte
	len() int
}

type snapshotNode interface {
	commitCheckpoint(originDb data.DBWriteCacher, targetDb data.DBWriteCacher) error
	commitSnapshot(originDb data.DBWriteCacher, targetDb data.DBWriteCacher) error
}

// RequestHandler defines the methods through which request to data can be made
type RequestHandler interface {
	RequestTrieNodes(destShardID uint32, hashes [][]byte, topic string)
	RequestInterval() time.Duration
	IsInterfaceNil() bool
}
