package trie2

import (
	"sync"

	protobuf "github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/proto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

const nrOfChildren = 17
const firstByte = 0
const maxTrieLevelAfterCommit = 6

type node interface {
	getHash() []byte
	setHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error
	setHashConcurrent(marshalizer marshal.Marshalizer, hasher hashing.Hasher, wg *sync.WaitGroup, c chan error)
	setRootHash(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error
	getCollapsed(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) // a collapsed node is a node that instead of the children holds the children hashes
	isCollapsed() bool
	isPosCollapsed(pos int) bool
	isDirty() bool
	getEncodedNode(marshal.Marshalizer) ([]byte, error)
	commit(level byte, dbw DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error
	resolveCollapsed(pos byte, dbw DBWriteCacher, marshalizer marshal.Marshalizer) error
	hashNode(marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error)
	hashChildren(marshalizer marshal.Marshalizer, hasher hashing.Hasher) error
	tryGet(key []byte, dbw DBWriteCacher, marshalizer marshal.Marshalizer) ([]byte, error)
	getNext(key []byte, dbw DBWriteCacher, marshalizer marshal.Marshalizer) (node, []byte, error)
	insert(n *leafNode, dbw DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error)
	delete(key []byte, dbw DBWriteCacher, marshalizer marshal.Marshalizer) (bool, node, error)
	reduceNode(pos int) node
	isEmptyOrNil() error
}

type branchNode struct {
	protobuf.CollapsedBn
	children [nrOfChildren]node
	hash     []byte
	dirty    bool
}

type extensionNode struct {
	protobuf.CollapsedEn
	child node
	hash  []byte
	dirty bool
}

type leafNode struct {
	protobuf.CollapsedLn
	hash  []byte
	dirty bool
}

func hashChildrenAndNode(n node, marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	err := n.hashChildren(marshalizer, hasher)
	if err != nil {
		return nil, err
	}
	hashed, err := n.hashNode(marshalizer, hasher)
	if err != nil {
		return nil, err
	}
	return hashed, nil
}

func encodeNodeAndGetHash(n node, marshalizer marshal.Marshalizer, hasher hashing.Hasher) ([]byte, error) {
	encNode, err := n.getEncodedNode(marshalizer)
	if err != nil {
		return nil, err
	}
	hash := hasher.Compute(string(encNode))
	return hash, nil
}

func encodeNodeAndCommitToDB(n node, db DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) error {
	key := n.getHash()
	if key == nil {
		err := n.setHash(marshalizer, hasher)
		if err != nil {
			return err
		}
		key = n.getHash()
	}
	n, err := n.getCollapsed(marshalizer, hasher)
	if err != nil {
		return err
	}
	val, err := n.getEncodedNode(marshalizer)
	if err != nil {
		return err
	}
	err = db.Put(key, val)
	return err
}

func getNodeFromDBAndDecode(n []byte, db DBWriteCacher, marshalizer marshal.Marshalizer) (node, error) {
	encChild, err := db.Get(n)
	if err != nil {
		return nil, err
	}
	node, err := decodeNode(encChild, marshalizer)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func resolveIfCollapsed(n node, pos byte, db DBWriteCacher, marshalizer marshal.Marshalizer) error {
	err := n.isEmptyOrNil()
	if err != nil {
		return err
	}
	if n.isPosCollapsed(int(pos)) {
		err := n.resolveCollapsed(pos, db, marshalizer)
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

func decodeNode(encNode []byte, marshalizer marshal.Marshalizer) (node, error) {
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
	return node, nil
}

func getEmptyNodeOfType(t byte) (node, error) {
	var decNode node
	switch t {
	case extension:
		decNode = &extensionNode{}
	case leaf:
		decNode = &leafNode{}
	case branch:
		decNode = newBranchNode()
	default:
		return nil, ErrInvalidNode
	}
	return decNode, nil
}

func childPosOutOfRange(pos byte) bool {
	if pos >= nrOfChildren {
		return true
	}
	return false
}
