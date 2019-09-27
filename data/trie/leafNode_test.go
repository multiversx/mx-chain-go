package trie

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	protobuf "github.com/ElrondNetwork/elrond-go/data/trie/proto"
	"github.com/stretchr/testify/assert"
)

func getLn() *leafNode {
	return newLeafNode([]byte("dog"), []byte("dog"))
}

func TestLeafNode_newLeafNode(t *testing.T) {
	t.Parallel()

	expectedLn := &leafNode{
		CollapsedLn: protobuf.CollapsedLn{
			Key:   []byte("dog"),
			Value: []byte("dog"),
		},
		hash:  nil,
		dirty: true,
	}
	ln := newLeafNode([]byte("dog"), []byte("dog"))
	assert.Equal(t, expectedLn, ln)
}

func TestLeafNode_getHash(t *testing.T) {
	t.Parallel()

	ln := &leafNode{hash: []byte("test hash")}
	assert.Equal(t, ln.hash, ln.getHash())
}

func TestLeafNode_isDirty(t *testing.T) {
	t.Parallel()

	ln := &leafNode{dirty: true}
	assert.Equal(t, true, ln.isDirty())

	ln = &leafNode{dirty: false}
	assert.Equal(t, false, ln.isDirty())
}

func TestLeafNode_getCollapsed(t *testing.T) {
	t.Parallel()

	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()

	collapsed, err := ln.getCollapsed(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, ln, collapsed)
}

func TestLeafNode_setHash(t *testing.T) {
	t.Parallel()

	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(ln, marsh, hasher)

	err := ln.setHash(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, hash, ln.hash)
}

func TestLeafNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()
	ln := &leafNode{}

	err := ln.setHash(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, ln.hash)
}

func TestLeafNode_setHashNilNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()
	var ln *leafNode

	err := ln.setHash(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, ln)
}

func TestLeafNode_setGivenHash(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	expectedHash := []byte("node hash")

	ln.setGivenHash(expectedHash)

	assert.Equal(t, expectedHash, ln.hash)
}

func TestLeafNode_hashChildren(t *testing.T) {
	t.Parallel()

	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()
	assert.Nil(t, ln.hashChildren(marsh, hasher))
}

func TestLeafNode_hashNode(t *testing.T) {
	t.Parallel()

	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()

	expectedHash, _ := encodeNodeAndGetHash(ln, marsh, hasher)
	hash, err := ln.hashNode(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestLeafNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	marsh, hasher := getTestMarshAndHasher()

	hash, err := ln.hashNode(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, hash)
}

func TestLeafNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	marsh, hasher := getTestMarshAndHasher()

	hash, err := ln.hashNode(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, hash)
}

func TestLeafNode_commit(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(ln, marsh, hasher)
	_ = ln.setHash(marsh, hasher)

	err := ln.commit(0, db, marsh, hasher)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	node, _ := decodeNode(encNode, marsh)
	ln = getLn()
	ln.dirty = false
	assert.Equal(t, ln, node)
}

func TestLeafNode_commitEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	db := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()

	err := ln.commit(0, db, marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestLeafNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	db := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()

	err := ln.commit(0, db, marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
}

func TestLeafNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	expectedEncodedNode, _ := marsh.Marshal(ln)
	expectedEncodedNode = append(expectedEncodedNode, leaf)

	encNode, err := ln.getEncodedNode(marsh)
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestLeafNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	marsh, _ := getTestMarshAndHasher()

	encNode, err := ln.getEncodedNode(marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, encNode)
}

func TestLeafNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	marsh, _ := getTestMarshAndHasher()

	encNode, err := ln.getEncodedNode(marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, encNode)
}

func TestLeafNode_resolveCollapsed(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	assert.Nil(t, ln.resolveCollapsed(0, db, marsh))
}

func TestLeafNode_isCollapsed(t *testing.T) {
	t.Parallel()

	ln := getLn()
	assert.False(t, ln.isCollapsed())
}

func TestLeafNode_tryGet(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte("dog")
	val, err := ln.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestLeafNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	wrongKey := []byte{1, 2, 3}
	val, err := ln.tryGet(wrongKey, db, marsh)
	assert.Nil(t, val)
	assert.Nil(t, err)
}

func TestLeafNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := &leafNode{}
	marsh, _ := getTestMarshAndHasher()

	key := []byte("dog")
	val, err := ln.tryGet(key, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, val)
}

func TestLeafNode_tryGetNilNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	var ln *leafNode
	marsh, _ := getTestMarshAndHasher()

	key := []byte("dog")
	val, err := ln.tryGet(key, db, marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, val)
}

func TestLeafNode_getNext(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()
	key := []byte("dog")

	node, key, err := ln.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Nil(t, err)
}

func TestLeafNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()
	wrongKey := append([]byte{2}, []byte("dog")...)

	node, key, err := ln.getNext(wrongKey, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestLeafNode_getNextNilNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	var ln *leafNode
	marsh, _ := getTestMarshAndHasher()
	key := []byte("dog")

	node, key, err := ln.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNilNode, err)
}

func TestLeafNode_insertAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	key := []byte("dog")
	expectedVal := []byte("dogs")

	node := newLeafNode(key, expectedVal)
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, _, err := ln.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newNode.tryGet(key, db, marsh)
	assert.Equal(t, expectedVal, val)
}

func TestLeafNode_insertAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()

	lnKey := []byte{2, 100, 111, 103}
	ln := newLeafNode(lnKey, []byte("dog"))

	nodeKey := []byte{3, 4, 5}
	nodeVal := []byte{3, 4, 5}
	node := newLeafNode(nodeKey, nodeVal)
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, _, err := ln.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newNode.tryGet(nodeKey, db, marsh)
	assert.Equal(t, nodeVal, val)
	assert.IsType(t, &branchNode{}, newNode)
}

func TestLeafNode_insertInStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()
	node := newLeafNode([]byte("dog"), []byte("dogs"))
	_ = ln.commit(0, db, marsh, hasher)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.insert(node, db, marsh)

	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_insertInStoredLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()
	ln := newLeafNode([]byte{1, 2, 3}, []byte("dog"))
	node := newLeafNode([]byte{4, 5, 6}, []byte("dogs"))
	_ = ln.commit(0, db, marsh, hasher)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.insert(node, db, marsh)

	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_insertInDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()
	node := newLeafNode([]byte("dog"), []byte("dogs"))

	dirty, _, oldHashes, err := ln.insert(node, db, marsh)

	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInDirtyLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	marsh, _ := getTestMarshAndHasher()
	ln := newLeafNode([]byte{1, 2, 3}, []byte("dog"))
	node := newLeafNode([]byte{4, 5, 6}, []byte("dogs"))

	dirty, _, oldHashes, err := ln.insert(node, db, marsh)

	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInNilNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	var ln *leafNode
	node := newLeafNode([]byte("dog"), []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, _, err := ln.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newNode)
}

func TestLeafNode_deletePresent(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, _, err := ln.delete([]byte("dog"), db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Nil(t, newNode)
}

func TestLeafNode_deleteFromStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()
	_ = ln.commit(0, db, marsh, hasher)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), db, marsh)

	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_deleteFromLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()
	ln := getLn()
	_ = ln.commit(0, db, marsh, hasher)

	wrongKey := []byte{1, 2, 3}
	dirty, _, oldHashes, err := ln.delete(wrongKey, db, marsh)

	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_deleteFromDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	marsh, _ := getTestMarshAndHasher()
	ln := getLn()

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), db, marsh)

	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_deleteNotPresent(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	wrongKey := []byte{1, 2, 3}
	dirty, newNode, _, err := ln.delete(wrongKey, db, marsh)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newNode)
}

func TestLeafNode_reduceNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{CollapsedLn: protobuf.CollapsedLn{Key: []byte{100, 111, 103}}}
	expected := &leafNode{CollapsedLn: protobuf.CollapsedLn{Key: []byte{2, 100, 111, 103}}, dirty: true}
	node := ln.reduceNode(2)
	assert.Equal(t, expected, node)
}

func TestLeafNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	assert.Equal(t, ErrEmptyNode, ln.isEmptyOrNil())

	ln = nil
	assert.Equal(t, ErrNilNode, ln.isEmptyOrNil())
}

//------- deepClone

func TestLeafNode_deepCloneWithNilHashShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	ln.dirty = true
	ln.hash = nil
	ln.Value = getRandomByteSlice()
	ln.Key = getRandomByteSlice()

	cloned := ln.deepClone().(*leafNode)

	testSameLeafNodeContent(t, ln, cloned)
}

func TestLeafNode_deepCloneWithNilValueShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	ln.dirty = true
	ln.hash = getRandomByteSlice()
	ln.Value = nil
	ln.Key = getRandomByteSlice()

	cloned := ln.deepClone().(*leafNode)

	testSameLeafNodeContent(t, ln, cloned)
}

func TestLeafNode_deepCloneWithNilKeyShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	ln.dirty = true
	ln.hash = getRandomByteSlice()
	ln.Value = getRandomByteSlice()
	ln.Key = nil

	cloned := ln.deepClone().(*leafNode)

	testSameLeafNodeContent(t, ln, cloned)
}

func TestLeafNode_deepCloneShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	ln.dirty = true
	ln.hash = getRandomByteSlice()
	ln.Value = getRandomByteSlice()
	ln.Key = getRandomByteSlice()

	cloned := ln.deepClone().(*leafNode)

	testSameLeafNodeContent(t, ln, cloned)
}

func testSameLeafNodeContent(t *testing.T, expected *leafNode, actual *leafNode) {
	if !reflect.DeepEqual(expected, actual) {
		assert.Fail(t, "not equal content")
		fmt.Printf(
			"expected:\n %s, got: \n%s",
			getLeafNodeContents(expected),
			getLeafNodeContents(actual),
		)
	}
	assert.False(t, expected == actual)
}

func getRandomByteSlice() []byte {
	maxChars := 32
	buff := make([]byte, maxChars)
	_, _ = rand.Reader.Read(buff)

	return buff
}

func getLeafNodeContents(lf *leafNode) string {
	str := fmt.Sprintf(`leaf node:
   key: %s
   value: %s
   hash: %s
   dirty: %v
`,
		hex.EncodeToString(lf.Key),
		hex.EncodeToString(lf.Value),
		hex.EncodeToString(lf.hash),
		lf.dirty)

	return str
}
