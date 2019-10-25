package trie

import (
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	protobuf "github.com/ElrondNetwork/elrond-go/data/trie/proto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

const nrOfChildren = 17
const firstByte = 0
const maxTrieLevelAfterCommit = 6
const hexTerminator = 16

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
	commit(force bool, level byte, db data.DBWriteCacher) error
	resolveCollapsed(pos byte) error
	hashNode() ([]byte, error)
	hashChildren() error
	tryGet(key []byte) ([]byte, error)
	getNext(key []byte) (node, []byte, error)
	insert(n *leafNode) (bool, node, [][]byte, error)
	delete(key []byte) (bool, node, [][]byte, error)
	reduceNode(pos int) (node, error)
	isEmptyOrNil() error
	print(writer io.Writer, index int)
	deepClone() node
	getDirtyHashes() ([][]byte, error)

	getMarshalizer() marshal.Marshalizer
	setMarshalizer(marshal.Marshalizer)
	getHasher() hashing.Hasher
	setHasher(hashing.Hasher)
	getDb() data.DBWriteCacher
	setDb(data.DBWriteCacher)
}

type baseNode struct {
	hash   []byte
	dirty  bool
	db     data.DBWriteCacher
	marsh  marshal.Marshalizer
	hasher hashing.Hasher
}

type branchNode struct {
	protobuf.CollapsedBn
	children [nrOfChildren]node
	*baseNode
}

type extensionNode struct {
	protobuf.CollapsedEn
	child node
	*baseNode
}

type leafNode struct {
	protobuf.CollapsedLn
	*baseNode
}

func hashChildrenAndNode(n node) ([]byte, error) {
	err := n.hashChildren()
	if err != nil {
		return nil, err
	}

	hashed, err := n.hashNode()
	if err != nil {
		return nil, err
	}

	return hashed, nil
}

func encodeNodeAndGetHash(n node) ([]byte, error) {
	encNode, err := n.getEncodedNode()
	if err != nil {
		return nil, err
	}

	hash := n.getHasher().Compute(string(encNode))

	return hash, nil
}

func encodeNodeAndCommitToDB(n node, db data.DBWriteCacher) error {
	key := n.getHash()
	if key == nil {
		err := n.setHash()
		if err != nil {
			return err
		}
		key = n.getHash()
	}

	n, err := n.getCollapsed()
	if err != nil {
		return err
	}

	val, err := n.getEncodedNode()
	if err != nil {
		return err
	}

	err = db.Put(key, val)

	return err
}

func getNodeFromDBAndDecode(n []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	encChild, err := db.Get(n)
	if err != nil {
		return nil, err
	}

	node, err := decodeNode(encChild, db, marshalizer, hasher)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func resolveIfCollapsed(n node, pos byte) error {
	err := n.isEmptyOrNil()
	if err != nil {
		return err
	}

	if n.isPosCollapsed(int(pos)) {
		err = n.resolveCollapsed(pos)
		if err != nil {
			return err
		}
	}

	return nil
}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)

	return r
}

func hasValidHash(n node) (bool, error) {
	err := n.isEmptyOrNil()
	if err != nil {
		return false, err
	}

	childHash := n.getHash()
	childIsDirty := n.isDirty()
	if childHash == nil || childIsDirty {
		return false, nil
	}

	return true, nil
}

func decodeNode(encNode []byte, db data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	if encNode == nil || len(encNode) < 1 {
		return nil, ErrInvalidEncoding
	}

	nodeType := encNode[len(encNode)-1]
	encNode = encNode[:len(encNode)-1]

	node, err := getEmptyNodeOfType(nodeType)
	if err != nil {
		return nil, err
	}

	err = marshalizer.Unmarshal(node, encNode)
	if err != nil {
		return nil, err
	}

	node.setDb(db)
	node.setMarshalizer(marshalizer)
	node.setHasher(hasher)

	return node, nil
}

func getEmptyNodeOfType(t byte) (node, error) {
	switch t {
	case extension:
		return &extensionNode{baseNode: &baseNode{}}, nil
	case leaf:
		return &leafNode{baseNode: &baseNode{}}, nil
	case branch:
		return &branchNode{baseNode: &baseNode{}}, nil
	default:
		return nil, ErrInvalidNode
	}
}

func childPosOutOfRange(pos byte) bool {
	return pos >= nrOfChildren
}

// keyBytesToHex transforms key bytes into hex nibbles
func keyBytesToHex(str []byte) []byte {
	length := len(str)*2 + 1
	nibbles := make([]byte, length)
	for i, b := range str {
		nibbles[i*2] = b / hexTerminator
		nibbles[i*2+1] = b % hexTerminator
	}
	nibbles[length-1] = hexTerminator

	return nibbles
}

// prefixLen returns the length of the common prefix of a and b.
func prefixLen(a, b []byte) int {
	i := 0
	length := len(a)
	if len(b) < length {
		length = len(b)
	}

	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}

	return i
}
