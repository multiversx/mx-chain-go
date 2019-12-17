package trie

import (
	"io"
	"sync"

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
	commit(force bool, level byte, originDb data.DBWriteCacher, targetDb data.DBWriteCacher) error
	resolveCollapsed(pos byte, db data.DBWriteCacher) error
	hashNode() ([]byte, error)
	hashChildren() error
	tryGet(key []byte, db data.DBWriteCacher) ([]byte, error)
	getNext(key []byte, db data.DBWriteCacher) (node, []byte, error)
	insert(n *leafNode, db data.DBWriteCacher) (bool, node, [][]byte, error)
	delete(key []byte, db data.DBWriteCacher) (bool, node, [][]byte, error)
	reduceNode(pos int) (node, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int)
	deepClone() node
	getDirtyHashes() ([][]byte, error)
	getChildren(db data.DBWriteCacher) ([]node, error)
	isValid() bool
	setDirty(bool)
	loadChildren(*trieSyncer) error

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
}

type snapshotsBuffer interface {
	add([]byte, bool)
	len() int
	removeFirst()
	getFirst() *snapshotsQueueEntry
	clone() snapshotsBuffer
}
