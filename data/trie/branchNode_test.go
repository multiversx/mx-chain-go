package trie

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	protobuf "github.com/ElrondNetwork/elrond-go/data/trie/proto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/stretchr/testify/assert"
)

func getTestDbMarshAndHasher() (data.DBWriteCacher, marshal.Marshalizer, hashing.Hasher) {
	marsh := &mock.ProtobufMarshalizerMock{}
	hasher := &mock.KeccakMock{}
	return mock.NewMemDbMock(), marsh, hasher
}

func getBnAndCollapsedBn(db data.DBWriteCacher, marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*branchNode, *branchNode) {
	var children [nrOfChildren]node
	EncodedChildren := make([][]byte, nrOfChildren)

	children[2], _ = newLeafNode([]byte("dog"), []byte("dog"), db, marshalizer, hasher)
	children[6], _ = newLeafNode([]byte("doe"), []byte("doe"), db, marshalizer, hasher)
	children[13], _ = newLeafNode([]byte("doge"), []byte("doge"), db, marshalizer, hasher)
	bn, _ := newBranchNode(db, marshalizer, hasher)
	bn.children = children

	EncodedChildren[2], _ = encodeNodeAndGetHash(children[2])
	EncodedChildren[6], _ = encodeNodeAndGetHash(children[6])
	EncodedChildren[13], _ = encodeNodeAndGetHash(children[13])
	collapsedBn, _ := newBranchNode(db, marshalizer, hasher)
	collapsedBn.EncodedChildren = EncodedChildren

	return bn, collapsedBn
}

func newEmptyTrie(db data.DBWriteCacher, marsh marshal.Marshalizer, hsh hashing.Hasher) *patriciaMerkleTrie {
	cacheSize := 100
	tr, _ := NewTrie(db, marsh, hsh, mock.NewMemDbMock(), cacheSize, config.DBConfig{})
	return tr
}

func initTrie() *patriciaMerkleTrie {
	tr := newEmptyTrie(getTestDbMarshAndHasher())
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	return tr
}

func getEncodedTrieNodesAndHashes(tr data.Trie) ([][]byte, [][]byte) {
	it, _ := NewIterator(tr)
	encNode, _ := it.GetMarshalizedNode()

	nodes := make([][]byte, 0)
	nodes = append(nodes, encNode)

	hashes := make([][]byte, 0)
	hash, _ := it.GetHash()
	hashes = append(hashes, hash)

	for it.HasNext() {
		_ = it.Next()
		encNode, _ = it.GetMarshalizedNode()

		nodes = append(nodes, encNode)
		hash, _ = it.GetHash()
		hashes = append(hashes, hash)
	}

	return nodes, hashes
}

func TestBranchNode_getHash(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{hash: []byte("test hash")}}
	assert.Equal(t, bn.hash, bn.getHash())
}

func TestBranchNode_isDirty(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{dirty: true}}
	assert.Equal(t, true, bn.isDirty())

	bn = &branchNode{baseNode: &baseNode{dirty: false}}
	assert.Equal(t, false, bn.isDirty())
}

func TestBranchNode_getCollapsed(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	collapsedBn.dirty = true

	collapsed, err := bn.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, collapsedBn, collapsed)
}

func TestBranchNode_getCollapsedEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	collapsed, err := bn.getCollapsed()
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, collapsed)
}

func TestBranchNode_getCollapsedNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	collapsed, err := bn.getCollapsed()
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, collapsed)
}

func TestBranchNode_getCollapsedCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	collapsed, err := collapsedBn.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, collapsedBn, collapsed)
}

func TestBranchNode_setHash(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	hash, _ := encodeNodeAndGetHash(collapsedBn)

	err := bn.setHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, bn.hash)
}

func TestBranchNode_setRootHash(t *testing.T) {
	t.Parallel()

	cfg := config.DBConfig{}
	db, marsh, hsh := getTestDbMarshAndHasher()
	cacheSize := 100

	tr1, _ := NewTrie(db, marsh, hsh, mock.NewMemDbMock(), cacheSize, cfg)
	tr2, _ := NewTrie(db, marsh, hsh, mock.NewMemDbMock(), cacheSize, cfg)

	for i := 0; i < 100000; i++ {
		val := hsh.Compute(string(i))
		_ = tr1.Update(val, val)
		_ = tr2.Update(val, val)
	}

	err := tr1.root.setRootHash()
	_ = tr2.root.setHash()
	assert.Nil(t, err)
	assert.Equal(t, tr1.root.getHash(), tr2.root.getHash())
}

func TestBranchNode_setRootHashCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	hash, _ := encodeNodeAndGetHash(collapsedBn)

	err := collapsedBn.setRootHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, collapsedBn.hash)
}

func TestBranchNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	err := bn.setHash()
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, bn.hash)
}

func TestBranchNode_setHashNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.setHash()
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, bn)
}

func TestBranchNode_setHashCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	hash, _ := encodeNodeAndGetHash(collapsedBn)

	err := collapsedBn.setHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, collapsedBn.hash)
}

func TestBranchNode_setGivenHash(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{}}
	expectedHash := []byte("node hash")

	bn.setGivenHash(expectedHash)
	assert.Equal(t, expectedHash, bn.hash)
}

func TestBranchNode_hashChildren(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	for i := range bn.children {
		if bn.children[i] != nil {
			assert.Nil(t, bn.children[i].getHash())
		}
	}
	err := bn.hashChildren()
	assert.Nil(t, err)

	for i := range bn.children {
		if bn.children[i] != nil {
			childHash, _ := encodeNodeAndGetHash(bn.children[i])
			assert.Equal(t, childHash, bn.children[i].getHash())
		}
	}
}

func TestBranchNode_hashChildrenEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	err := bn.hashChildren()
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_hashChildrenNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.hashChildren()
	assert.Equal(t, ErrNilNode, err)
}

func TestBranchNode_hashChildrenCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	err := collapsedBn.hashChildren()
	assert.Nil(t, err)

	_, collapsedBn2 := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	assert.Equal(t, collapsedBn2, collapsedBn)
}

func TestBranchNode_hashNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	expectedHash, _ := encodeNodeAndGetHash(collapsedBn)

	hash, err := collapsedBn.hashNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestBranchNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	hash, err := bn.hashNode()
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, hash)
}

func TestBranchNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	hash, err := bn.hashNode()
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, hash)
}

func TestBranchNode_commit(t *testing.T) {
	t.Parallel()

	db, marsh, hasher := getTestDbMarshAndHasher()
	bn, collapsedBn := getBnAndCollapsedBn(db, marsh, hasher)

	hash, _ := encodeNodeAndGetHash(collapsedBn)
	_ = bn.setHash()

	err := bn.commit(false, 0, bn.db)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	node, _ := decodeNode(encNode, db, marsh, hasher)
	h1, _ := encodeNodeAndGetHash(collapsedBn)
	h2, _ := encodeNodeAndGetHash(node)
	assert.Equal(t, h1, h2)
}

func TestBranchNode_commitEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	err := bn.commit(false, 0, bn.db)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.commit(false, 0, nil)
	assert.Equal(t, ErrNilNode, err)
}

func TestBranchNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	expectedEncodedNode, _ := bn.marsh.Marshal(bn)
	expectedEncodedNode = append(expectedEncodedNode, branch)

	encNode, err := bn.getEncodedNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestBranchNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	encNode, err := bn.getEncodedNode()
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, encNode)
}

func TestBranchNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	encNode, err := bn.getEncodedNode()
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, encNode)
}

func TestBranchNode_resolveCollapsed(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	childPos := byte(2)

	_ = bn.setHash()
	_ = bn.commit(false, 0, bn.db)
	resolved, _ := newLeafNode([]byte("dog"), []byte("dog"), bn.db, bn.marsh, bn.hasher)
	resolved.dirty = false
	resolved.hash = bn.EncodedChildren[childPos]

	err := collapsedBn.resolveCollapsed(childPos)
	assert.Nil(t, err)
	assert.Equal(t, resolved, collapsedBn.children[childPos])
}

func TestBranchNode_resolveCollapsedEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	err := bn.resolveCollapsed(2)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_resolveCollapsedENilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.resolveCollapsed(2)
	assert.Equal(t, ErrNilNode, err)
}

func TestBranchNode_resolveCollapsedPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	err := bn.resolveCollapsed(17)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_isCollapsed(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	assert.True(t, collapsedBn.isCollapsed())
	assert.False(t, bn.isCollapsed())

	collapsedBn.children[2], _ = newLeafNode([]byte("dog"), []byte("dog"), bn.db, bn.marsh, bn.hasher)
	assert.False(t, collapsedBn.isCollapsed())
}

func TestBranchNode_tryGet(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	val, err := bn.tryGet(key)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestBranchNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	var key []byte

	val, err := bn.tryGet(key)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	key := []byte("dog")

	val, err := bn.tryGet(key)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	nilChildKey := []byte{3}

	val, err := bn.tryGet(nilChildKey)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	_ = bn.setHash()
	_ = bn.commit(false, 0, bn.db)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	val, err := collapsedBn.tryGet(key)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestBranchNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	val, err := bn.tryGet(key)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	val, err := bn.tryGet(key)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, val)
}

func TestBranchNode_getNext(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	nextNode, _ := newLeafNode([]byte("dog"), []byte("dog"), bn.db, bn.marsh, bn.hasher)
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	node, key, err := bn.getNext(key)

	h1, _ := encodeNodeAndGetHash(nextNode)
	h2, _ := encodeNodeAndGetHash(node)
	assert.Equal(t, h1, h2)
	assert.Equal(t, []byte("dog"), key)
	assert.Nil(t, err)
}

func TestBranchNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	key := []byte("dog")

	node, key, err := bn.getNext(key)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_getNextNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	nilChildPos := byte(4)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	node, key, err := bn.getNext(key)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestBranchNode_insert(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	nodeKey := []byte{0, 2, 3}
	node, _ := newLeafNode(nodeKey, []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	dirty, newBn, _, err := bn.insert(node)
	nodeKeyRemainder := nodeKey[1:]

	bn.children[0], _ = newLeafNode(nodeKeyRemainder, []byte("dogs"), bn.db, bn.marsh, bn.hasher)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_insertEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	node, _ := newLeafNode([]byte{}, []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	dirty, newBn, _, err := bn.insert(node)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	node, _ := newLeafNode([]byte("dog"), []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	dirty, newBn, _, err := bn.insert(node)
	assert.False(t, dirty)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	node, _ := newLeafNode(key, []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	_ = bn.setHash()
	_ = bn.commit(false, 0, bn.db)

	dirty, newBn, _, err := collapsedBn.insert(node)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, _ := newBn.tryGet(key)
	assert.Equal(t, []byte("dogs"), val)
}

func TestBranchNode_insertInStoredBnOnExistingPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	node, _ := newLeafNode(key, []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	_ = bn.commit(false, 0, bn.db)
	bnHash := bn.getHash()
	ln, _, _ := bn.getNext(key)
	lnHash := ln.getHash()
	expectedHashes := [][]byte{lnHash, bnHash}

	dirty, _, oldHashes, err := bn.insert(node)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestBranchNode_insertInStoredBnOnNilPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)
	node, _ := newLeafNode(key, []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	_ = bn.commit(false, 0, bn.db)
	bnHash := bn.getHash()
	expectedHashes := [][]byte{bnHash}

	dirty, _, oldHashes, err := bn.insert(node)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestBranchNode_insertInDirtyBnOnNilPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)
	node, _ := newLeafNode(key, []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	dirty, _, oldHashes, err := bn.insert(node)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestBranchNode_insertInDirtyBnOnExistingPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	node, _ := newLeafNode(key, []byte("dogs"), bn.db, bn.marsh, bn.hasher)

	dirty, _, oldHashes, err := bn.insert(node)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestBranchNode_insertInNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	dirty, newBn, _, err := bn.insert(&leafNode{})
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_delete(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	var children [nrOfChildren]node
	children[6], _ = newLeafNode([]byte("doe"), []byte("doe"), bn.db, bn.marsh, bn.hasher)
	children[13], _ = newLeafNode([]byte("doge"), []byte("doge"), bn.db, bn.marsh, bn.hasher)
	expectedBn, _ := newBranchNode(bn.db, bn.marsh, bn.hasher)
	expectedBn.children = children

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	dirty, newBn, _, err := bn.delete(key)
	assert.True(t, dirty)
	assert.Nil(t, err)

	_ = expectedBn.setHash()
	_ = newBn.setHash()
	assert.Equal(t, expectedBn.getHash(), newBn.getHash())
}

func TestBranchNode_deleteFromStoredBn(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)

	_ = bn.commit(false, 0, bn.db)
	bnHash := bn.getHash()
	ln, _, _ := bn.getNext(lnKey)
	lnHash := ln.getHash()
	expectedHashes := [][]byte{lnHash, bnHash}

	dirty, _, oldHashes, err := bn.delete(lnKey)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestBranchNode_deleteFromDirtyBn(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)

	dirty, _, oldHashes, err := bn.delete(lnKey)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestBranchNode_deleteEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	dirty, newBn, _, err := bn.delete(key)
	assert.False(t, dirty)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	dirty, newBn, _, err := bn.delete(key)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteEmptykey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	dirty, newBn, _, err := bn.delete([]byte{})
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	_ = bn.setHash()
	_ = bn.commit(false, 0, bn.db)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	dirty, newBn, _, err := collapsedBn.delete(key)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, err := newBn.tryGet(key)
	assert.Nil(t, val)
	assert.Nil(t, err)
}

func TestBranchNode_deleteAndReduceBn(t *testing.T) {
	t.Parallel()

	bn, _ := newBranchNode(getTestDbMarshAndHasher())
	var children [nrOfChildren]node
	firstChildPos := byte(2)
	secondChildPos := byte(6)
	children[firstChildPos], _ = newLeafNode([]byte("dog"), []byte("dog"), bn.db, bn.marsh, bn.hasher)
	children[secondChildPos], _ = newLeafNode([]byte("doe"), []byte("doe"), bn.db, bn.marsh, bn.hasher)
	bn.children = children

	key := append([]byte{firstChildPos}, []byte("dog")...)
	ln, _ := newLeafNode(key, []byte("dog"), bn.db, bn.marsh, bn.hasher)

	key = append([]byte{secondChildPos}, []byte("doe")...)
	dirty, newBn, _, err := bn.delete(key)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newBn)
}

func TestBranchNode_reduceNode(t *testing.T) {
	t.Parallel()

	bn, _ := newBranchNode(getTestDbMarshAndHasher())
	var children [nrOfChildren]node
	childPos := byte(2)
	children[childPos], _ = newLeafNode([]byte("dog"), []byte("dog"), bn.db, bn.marsh, bn.hasher)
	bn.children = children

	key := append([]byte{childPos}, []byte("dog")...)
	ln, _ := newLeafNode(key, []byte("dog"), bn.db, bn.marsh, bn.hasher)

	node, err := bn.children[childPos].reduceNode(int(childPos))
	assert.Equal(t, ln, node)
	assert.Nil(t, err)
}

func TestBranchNode_getChildPosition(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	nr, pos := getChildPosition(bn)
	assert.Equal(t, 3, nr)
	assert.Equal(t, 13, pos)
}

func TestBranchNode_clone(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	clone := bn.clone()
	assert.False(t, bn == clone)
	assert.Equal(t, bn, clone)
}

func TestBranchNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()
	assert.Equal(t, ErrEmptyNode, bn.isEmptyOrNil())

	bn = nil
	assert.Equal(t, ErrNilNode, bn.isEmptyOrNil())
}

func TestReduceBranchNodeWithExtensionNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr := newEmptyTrie(getTestDbMarshAndHasher())
	expectedTr := newEmptyTrie(getTestDbMarshAndHasher())

	_ = expectedTr.Update([]byte("dog"), []byte("dog"))
	_ = expectedTr.Update([]byte("doll"), []byte("doll"))

	_ = tr.Update([]byte("dog"), []byte("dog"))
	_ = tr.Update([]byte("doll"), []byte("doll"))
	_ = tr.Update([]byte("wolf"), []byte("wolf"))
	_ = tr.Delete([]byte("wolf"))

	expectedHash, _ := expectedTr.Root()
	hash, _ := tr.Root()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithBranchNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr := newEmptyTrie(getTestDbMarshAndHasher())
	expectedTr := newEmptyTrie(getTestDbMarshAndHasher())

	_ = expectedTr.Update([]byte("dog"), []byte("puppy"))
	_ = expectedTr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Delete([]byte("doe"))

	expectedHash, _ := expectedTr.Root()
	hash, _ := tr.Root()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithLeafNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr := newEmptyTrie(getTestDbMarshAndHasher())
	expectedTr := newEmptyTrie(getTestDbMarshAndHasher())

	_ = expectedTr.Update([]byte("doe"), []byte("reindeer"))
	_ = expectedTr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Delete([]byte("dog"))

	expectedHash, _ := expectedTr.Root()
	hash, _ := tr.Root()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithLeafNodeValueShouldWork(t *testing.T) {
	t.Parallel()

	tr := newEmptyTrie(getTestDbMarshAndHasher())
	expectedTr := newEmptyTrie(getTestDbMarshAndHasher())

	_ = expectedTr.Update([]byte("doe"), []byte("reindeer"))
	_ = expectedTr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Delete([]byte("dogglesworth"))

	expectedHash, _ := expectedTr.Root()
	hash, _ := tr.Root()

	assert.Equal(t, expectedHash, hash)
}

func TestBranchNode_getChildren(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())

	children, err := bn.getChildren()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(children))
}

func TestBranchNode_getChildrenCollapsedBn(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	_ = bn.commit(true, 0, collapsedBn.db)

	children, err := collapsedBn.getChildren()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(children))
}

func TestBranchNode_isValid(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	assert.True(t, bn.isValid())

	bn.children[2] = nil
	bn.children[6] = nil
	assert.False(t, bn.isValid())
}

func TestBranchNode_setDirty(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{}}
	bn.setDirty(true)

	assert.True(t, bn.dirty)
}

func TestBranchNode_loadChildren(t *testing.T) {
	t.Parallel()

	_, marsh, hasher := getTestDbMarshAndHasher()
	tr := initTrie()
	nodes, hashes := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := lrucache.NewCache(100)

	resolver := &mock.TrieNodesResolverStub{
		RequestDataFromHashCalled: func(hash []byte) error {
			for i := range nodes {
				node, _ := NewInterceptedTrieNode(nodes[i], tr.GetDatabase(), marsh, hasher)
				nodesCacher.Put(node.hash, node)
			}
			return nil
		},
	}
	syncer, _ := NewTrieSyncer(resolver, nodesCacher, tr, time.Second)
	syncer.interceptedNodes.RegisterHandler(func(key []byte) {
		syncer.chRcvTrieNodes <- true
	})

	bnHashPosition := 1
	firstChildPos := 5
	firstChildHash := 2
	secondChildPos := 7
	secondChildHash := 3
	encodedChildren := make([][]byte, nrOfChildren)
	encodedChildren[firstChildPos] = hashes[firstChildHash]
	encodedChildren[secondChildPos] = hashes[secondChildHash]
	bn := &branchNode{
		CollapsedBn: protobuf.CollapsedBn{
			EncodedChildren: encodedChildren,
		},
		baseNode: &baseNode{
			hash: hashes[bnHashPosition],
		},
	}

	err := bn.loadChildren(syncer)
	assert.Nil(t, err)
	assert.NotNil(t, bn.children[firstChildPos])
	assert.NotNil(t, bn.children[secondChildPos])

	assert.Equal(t, 5, nodesCacher.Len())
}

//------- deepClone

func TestBranchNode_deepCloneWithNilHashShouldWork(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{}}
	bn.dirty = true
	bn.hash = nil
	bn.EncodedChildren = make([][]byte, len(bn.children))
	bn.EncodedChildren[4] = getRandomByteSlice()
	bn.EncodedChildren[5] = getRandomByteSlice()
	bn.EncodedChildren[12] = getRandomByteSlice()
	bn.children[4] = &leafNode{baseNode: &baseNode{}}
	bn.children[5] = &leafNode{baseNode: &baseNode{}}
	bn.children[12] = &leafNode{baseNode: &baseNode{}}

	cloned := bn.deepClone().(*branchNode)

	testSameBranchNodeContent(t, bn, cloned)
}

func TestBranchNode_deepCloneShouldWork(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{}}
	bn.dirty = true
	bn.hash = getRandomByteSlice()
	bn.EncodedChildren = make([][]byte, len(bn.children))
	bn.EncodedChildren[4] = getRandomByteSlice()
	bn.EncodedChildren[5] = getRandomByteSlice()
	bn.EncodedChildren[12] = getRandomByteSlice()
	bn.children[4] = &leafNode{baseNode: &baseNode{}}
	bn.children[5] = &leafNode{baseNode: &baseNode{}}
	bn.children[12] = &leafNode{baseNode: &baseNode{}}

	cloned := bn.deepClone().(*branchNode)

	testSameBranchNodeContent(t, bn, cloned)
}

func testSameBranchNodeContent(t *testing.T, expected *branchNode, actual *branchNode) {
	if !reflect.DeepEqual(expected, actual) {
		assert.Fail(t, "not equal content")
		fmt.Printf(
			"expected:\n %s, got: \n%s",
			getBranchNodeContents(expected),
			getBranchNodeContents(actual),
		)
	}
	assert.False(t, expected == actual)
}

func getBranchNodeContents(bn *branchNode) string {
	encodedChildsString := ""
	for i := 0; i < len(bn.EncodedChildren); i++ {
		if i > 0 {
			encodedChildsString += ", "
		}

		if bn.EncodedChildren[i] == nil {
			encodedChildsString += "<nil>"
			continue
		}

		encodedChildsString += hex.EncodeToString(bn.EncodedChildren[i])
	}

	childsString := ""
	for i := 0; i < len(bn.children); i++ {
		if i > 0 {
			childsString += ", "
		}

		if bn.children[i] == nil {
			childsString += "<nil>"
			continue
		}

		childsString += fmt.Sprintf("%p", bn.children[i])
	}

	str := fmt.Sprintf(`extension node:
  		encoded child: %s
  		hash: %s
 		child: %s,
  		dirty: %v
`,
		encodedChildsString,
		hex.EncodeToString(bn.hash),
		childsString,
		bn.dirty)

	return str
}

func BenchmarkDecodeBranchNode(b *testing.B) {
	db, marsh, hsh := getTestDbMarshAndHasher()
	tr := newEmptyTrie(db, marsh, hsh)

	nrValuesInTrie := 100000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		_ = tr.Update(values[i], values[i])
	}

	proof, _ := tr.Prove(values[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decodeNode(proof[0], db, marsh, hsh)
	}
}

func BenchmarkMarshallNodeCapnp(b *testing.B) {
	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	marsh := &marshal.CapnpMarshalizer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = marsh.Marshal(bn)
	}
}

func BenchmarkMarshallNodeJson(b *testing.B) {
	bn, _ := getBnAndCollapsedBn(getTestDbMarshAndHasher())
	marsh := marshal.JsonMarshalizer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = marsh.Marshal(bn)
	}
}
