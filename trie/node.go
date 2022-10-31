//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. node.proto
package trie

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/trie/keyBuilder"
)

const (
	nrOfChildren         = 17
	firstByte            = 0
	hexTerminator        = 16
	nibbleMask           = 0x0f
	pointerSizeInBytes   = 8
	numNodeInnerPointers = 2 // each trie node contains a marshalizer and a hasher
	pollingIdleNode      = time.Millisecond
)

type baseNode struct {
	hash   []byte
	dirty  bool
	marsh  marshal.Marshalizer
	hasher hashing.Hasher
}

type branchNode struct {
	CollapsedBn
	children [nrOfChildren]node
	*baseNode
}

type extensionNode struct {
	CollapsedEn
	child node
	*baseNode
}

type leafNode struct {
	CollapsedLn
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

// encodeNodeAndCommitToDB will encode and save provided node. It returns the node's value in bytes
func encodeNodeAndCommitToDB(n node, db common.DBWriteCacher) (int, error) {
	key, err := computeAndSetNodeHash(n)
	if err != nil {
		return 0, err
	}

	val, err := collapseAndEncodeNode(n)
	if err != nil {
		return 0, err
	}

	// test point encodeNodeAndCommitToDB

	err = db.Put(key, val)

	return len(val), err
}

func collapseAndEncodeNode(n node) ([]byte, error) {
	n, err := n.getCollapsed()
	if err != nil {
		return nil, err
	}

	return n.getEncodedNode()
}

func computeAndSetNodeHash(n node) ([]byte, error) {
	key := n.getHash()
	if len(key) != 0 {
		return key, nil
	}

	err := n.setHash()
	if err != nil {
		return nil, err
	}
	key = n.getHash()

	return key, nil
}

func getNodeFromDBAndDecode(n []byte, db common.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	encChild, err := db.Get(n)
	if err != nil {
		log.Trace(common.GetNodeFromDBErrorString, "error", err, "key", n)
		return nil, fmt.Errorf(common.GetNodeFromDBErrorString+" %w for key %v", err, hex.EncodeToString(n))
	}

	return decodeNode(encChild, marshalizer, hasher)
}

func resolveIfCollapsed(n node, pos byte, db common.DBWriteCacher) error {
	err := n.isEmptyOrNil()
	if err != nil {
		return err
	}

	if n.isPosCollapsed(int(pos)) {
		err = n.resolveCollapsed(pos, db)
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

func decodeNode(encNode []byte, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (node, error) {
	if encNode == nil || len(encNode) < 1 {
		return nil, ErrInvalidEncoding
	}

	nodeType := encNode[len(encNode)-1]
	encNode = encNode[:len(encNode)-1]

	newNode, err := getEmptyNodeOfType(nodeType)
	if err != nil {
		return nil, err
	}

	err = marshalizer.Unmarshal(newNode, encNode)
	if err != nil {
		return nil, err
	}

	newNode.setMarshalizer(marshalizer)
	newNode.setHasher(hasher)

	return newNode, nil
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

// keyBytesToHex transforms key bytes into hex nibbles. The key nibbles are reversed, meaning that the
// last key nibble will be the first in the hex key. A hex terminator is added at the end of the hex key.
func keyBytesToHex(str []byte) []byte {
	hexLength := len(str)*2 + 1
	nibbles := make([]byte, hexLength)

	hexSliceIndex := 0
	nibbles[hexLength-1] = hexTerminator

	for i := hexLength - 2; i > 0; i -= 2 {
		nibbles[i] = str[hexSliceIndex] >> keyBuilder.NibbleSize
		nibbles[i-1] = str[hexSliceIndex] & nibbleMask
		hexSliceIndex++
	}

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

func shouldStopIfContextDone(ctx context.Context, idleProvider IdleNodeProvider) bool {
	for {
		select {
		case <-ctx.Done():
			return true
		default:
		}

		if idleProvider.IsIdle() {
			return false
		}

		select {
		case <-ctx.Done():
			return true
		case <-time.After(pollingIdleNode):
		}
	}
}

func treatCommitSnapshotError(err error, hash []byte, missingNodesChan chan []byte) {
	if errors.IsClosingError(err) {
		log.Debug("context closing", "hash", hash)
		return
	}

	log.Error("error during trie snapshot", "err", err.Error(), "hash", hash)
	missingNodesChan <- hash
}
