package trie

import (
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
	db, _ := mock.NewMemDbMock()
	ln := getLn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(ln, marsh, hasher)
	ln.setHash(marsh, hasher)

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
	db, _ := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()

	err := ln.commit(0, db, marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestLeafNode_commitNilNode(t *testing.T) {
	t.Parallel()
	var ln *leafNode
	db, _ := mock.NewMemDbMock()
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
	db, _ := mock.NewMemDbMock()
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
	db, _ := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 111, 103}
	val, err := ln.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestLeafNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{1, 2, 3}
	val, err := ln.tryGet(key, db, marsh)
	assert.Nil(t, val)
	assert.Nil(t, err)
}

func TestLeafNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := &leafNode{}
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 111, 103}
	val, err := ln.tryGet(key, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, val)
}

func TestLeafNode_tryGetNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var ln *leafNode
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 111, 103}
	val, err := ln.tryGet(key, db, marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, val)
}

func TestLeafNode_getNext(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()
	key := []byte{100, 111, 103}

	node, key, err := ln.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Nil(t, err)
}

func TestLeafNode_getNextWrongKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()
	key := []byte{2, 100, 111, 103}

	node, key, err := ln.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestLeafNode_getNextNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var ln *leafNode
	marsh, _ := getTestMarshAndHasher()
	key := []byte{2, 100, 111, 103}

	node, key, err := ln.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNilNode, err)
}

func TestLeafNode_insertAtSameKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := getLn()

	node := newLeafNode([]byte{100, 111, 103}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := ln.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newNode.tryGet([]byte{100, 111, 103}, db, marsh)
	assert.Equal(t, []byte("dogs"), val)
}

func TestLeafNode_insertAtDifferentKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := newLeafNode([]byte{2, 100, 111, 103}, []byte{100, 111, 103})
	node := newLeafNode([]byte{3, 4, 5}, []byte{3, 4, 5})
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := ln.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newNode.tryGet([]byte{3, 4, 5}, db, marsh)
	assert.Equal(t, []byte{3, 4, 5}, val)
	assert.IsType(t, &branchNode{}, newNode)
}

func TestLeafNode_insertInNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var ln *leafNode
	node := newLeafNode([]byte{0, 2, 3}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := ln.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newNode)
}

func TestLeafNode_deletePresent(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := ln.delete([]byte{100, 111, 103}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Nil(t, newNode)
}

func TestLeafNode_deleteNotPresent(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	ln := getLn()
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := ln.delete([]byte{1, 2, 3}, db, marsh)
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
