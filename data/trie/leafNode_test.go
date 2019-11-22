package trie

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	protobuf "github.com/ElrondNetwork/elrond-go/data/trie/proto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/stretchr/testify/assert"
)

func getLn(marsh marshal.Marshalizer, hasher hashing.Hasher) *leafNode {
	newLn, _ := newLeafNode([]byte("dog"), []byte("dog"), marsh, hasher)
	return newLn
}

func TestLeafNode_newLeafNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()
	expectedLn := &leafNode{
		CollapsedLn: protobuf.CollapsedLn{
			Key:   []byte("dog"),
			Value: []byte("dog"),
		},
		baseNode: &baseNode{
			dirty:  true,
			marsh:  marsh,
			hasher: hasher,
		},
	}
	ln, _ := newLeafNode([]byte("dog"), []byte("dog"), marsh, hasher)
	assert.Equal(t, expectedLn, ln)
}

func TestLeafNode_getHash(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{hash: []byte("test hash")}}
	assert.Equal(t, ln.hash, ln.getHash())
}

func TestLeafNode_isDirty(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{dirty: true}}
	assert.Equal(t, true, ln.isDirty())

	ln = &leafNode{baseNode: &baseNode{dirty: false}}
	assert.Equal(t, false, ln.isDirty())
}

func TestLeafNode_getCollapsed(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())

	collapsed, err := ln.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, ln, collapsed)
}

func TestLeafNode_setHash(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	hash, _ := encodeNodeAndGetHash(ln)

	err := ln.setHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, ln.hash)
}

func TestLeafNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}

	err := ln.setHash()
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, ln.hash)
}

func TestLeafNode_setHashNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	err := ln.setHash()
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, ln)
}

func TestLeafNode_setGivenHash(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}
	expectedHash := []byte("node hash")

	ln.setGivenHash(expectedHash)
	assert.Equal(t, expectedHash, ln.hash)
}

func TestLeafNode_hashChildren(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())

	assert.Nil(t, ln.hashChildren())
}

func TestLeafNode_hashNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	expectedHash, _ := encodeNodeAndGetHash(ln)

	hash, err := ln.hashNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestLeafNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	hash, err := ln.hashNode()
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, hash)
}

func TestLeafNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	hash, err := ln.hashNode()
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, hash)
}

func TestLeafNode_commit(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshAndHasher())
	hash, _ := encodeNodeAndGetHash(ln)
	_ = ln.setHash()

	err := ln.commit(false, 0, db, db)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	node, _ := decodeNode(encNode, ln.marsh, ln.hasher)
	ln = getLn(ln.marsh, ln.hasher)
	ln.dirty = false
	assert.Equal(t, ln, node)
}

func TestLeafNode_commitEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	err := ln.commit(false, 0, nil, nil)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestLeafNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	err := ln.commit(false, 0, nil, nil)
	assert.Equal(t, ErrNilNode, err)
}

func TestLeafNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	expectedEncodedNode, _ := ln.marsh.Marshal(ln)
	expectedEncodedNode = append(expectedEncodedNode, leaf)

	encNode, err := ln.getEncodedNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestLeafNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	encNode, err := ln.getEncodedNode()
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, encNode)
}

func TestLeafNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	encNode, err := ln.getEncodedNode()
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, encNode)
}

func TestLeafNode_resolveCollapsed(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())

	assert.Nil(t, ln.resolveCollapsed(0, nil))
}

func TestLeafNode_isCollapsed(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	assert.False(t, ln.isCollapsed())
}

func TestLeafNode_tryGet(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	key := []byte("dog")

	val, err := ln.tryGet(key, nil)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestLeafNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	wrongKey := []byte{1, 2, 3}

	val, err := ln.tryGet(wrongKey, nil)
	assert.Nil(t, val)
	assert.Nil(t, err)
}

func TestLeafNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	key := []byte("dog")
	val, err := ln.tryGet(key, nil)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, val)
}

func TestLeafNode_tryGetNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	key := []byte("dog")

	val, err := ln.tryGet(key, nil)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, val)
}

func TestLeafNode_getNext(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	key := []byte("dog")

	node, key, err := ln.getNext(key, nil)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Nil(t, err)
}

func TestLeafNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	wrongKey := append([]byte{2}, []byte("dog")...)

	node, key, err := ln.getNext(wrongKey, nil)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestLeafNode_getNextNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	key := []byte("dog")

	node, key, err := ln.getNext(key, nil)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNilNode, err)
}

func TestLeafNode_insertAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	key := []byte("dog")
	expectedVal := []byte("dogs")
	node, _ := newLeafNode(key, expectedVal, ln.marsh, ln.hasher)

	dirty, newNode, _, err := ln.insert(node, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, _ := newNode.tryGet(key, nil)
	assert.Equal(t, expectedVal, val)
}

func TestLeafNode_insertAtDifferentKey(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()

	lnKey := []byte{2, 100, 111, 103}
	ln, _ := newLeafNode(lnKey, []byte("dog"), marsh, hasher)

	nodeKey := []byte{3, 4, 5}
	nodeVal := []byte{3, 4, 5}
	node, _ := newLeafNode(nodeKey, nodeVal, marsh, hasher)

	dirty, newNode, _, err := ln.insert(node, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, _ := newNode.tryGet(nodeKey, nil)
	assert.Equal(t, nodeVal, val)
	assert.IsType(t, &branchNode{}, newNode)
}

func TestLeafNode_insertInStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshAndHasher())
	node, _ := newLeafNode([]byte("dog"), []byte("dogs"), ln.marsh, ln.hasher)
	_ = ln.commit(false, 0, db, db)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.insert(node, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_insertInStoredLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()
	ln, _ := newLeafNode([]byte{1, 2, 3}, []byte("dog"), marsh, hasher)
	node, _ := newLeafNode([]byte{4, 5, 6}, []byte("dogs"), marsh, hasher)
	_ = ln.commit(false, 0, db, db)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.insert(node, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_insertInDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	node, _ := newLeafNode([]byte("dog"), []byte("dogs"), ln.marsh, ln.hasher)

	dirty, _, oldHashes, err := ln.insert(node, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInDirtyLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()
	ln, _ := newLeafNode([]byte{1, 2, 3}, []byte("dog"), marsh, hasher)
	node, _ := newLeafNode([]byte{4, 5, 6}, []byte("dogs"), marsh, hasher)

	dirty, _, oldHashes, err := ln.insert(node, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	dirty, newNode, _, err := ln.insert(&leafNode{}, nil)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newNode)
}

func TestLeafNode_deletePresent(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())

	dirty, newNode, _, err := ln.delete([]byte("dog"), nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Nil(t, newNode)
}

func TestLeafNode_deleteFromStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshAndHasher())
	_ = ln.commit(false, 0, db, db)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_deleteFromLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshAndHasher())
	_ = ln.commit(false, 0, db, db)
	wrongKey := []byte{1, 2, 3}

	dirty, _, oldHashes, err := ln.delete(wrongKey, db)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_deleteFromDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_deleteNotPresent(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	wrongKey := []byte{1, 2, 3}

	dirty, newNode, _, err := ln.delete(wrongKey, nil)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newNode)
}

func TestLeafNode_reduceNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()
	ln, _ := newLeafNode([]byte{100, 111, 103}, nil, marsh, hasher)
	expected, _ := newLeafNode([]byte{2, 100, 111, 103}, nil, marsh, hasher)
	expected.dirty = true

	node, err := ln.reduceNode(2)
	assert.Equal(t, expected, node)
	assert.Nil(t, err)
}

func TestLeafNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	assert.Equal(t, ErrEmptyNode, ln.isEmptyOrNil())

	ln = nil
	assert.Equal(t, ErrNilNode, ln.isEmptyOrNil())
}

func TestLeafNode_getChildren(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())

	children, err := ln.getChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(children))
}

func TestLeafNode_isValid(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshAndHasher())
	assert.True(t, ln.isValid())

	ln.Value = []byte{}
	assert.False(t, ln.isValid())
}

func TestLeafNode_setDirty(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}
	ln.setDirty(true)

	assert.True(t, ln.dirty)
}

func TestLeafNode_loadChildren(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()
	tr := initTrie()
	nodes, hashes := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := lrucache.NewCache(100)

	resolver := &mock.TrieNodesResolverStub{}
	for i := range nodes {
		node, _ := NewInterceptedTrieNode(nodes[i], marsh, hasher)
		nodesCacher.Put(node.hash, node)
	}
	syncer, _ := NewTrieSyncer(resolver, nodesCacher, tr, time.Second)
	syncer.interceptedNodes.RegisterHandler(func(key []byte) {
		syncer.chRcvTrieNodes <- true
	})

	lnPosition := 5
	ln := &leafNode{baseNode: &baseNode{hash: hashes[lnPosition]}}
	err := ln.loadChildren(syncer)
	assert.Nil(t, err)
	assert.Equal(t, 5, nodesCacher.Len())
}

//------- deepClone

func TestLeafNode_deepCloneWithNilHashShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}
	ln.dirty = true
	ln.hash = nil
	ln.Value = getRandomByteSlice()
	ln.Key = getRandomByteSlice()

	cloned := ln.deepClone().(*leafNode)

	testSameLeafNodeContent(t, ln, cloned)
}

func TestLeafNode_deepCloneWithNilValueShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}
	ln.dirty = true
	ln.hash = getRandomByteSlice()
	ln.Value = nil
	ln.Key = getRandomByteSlice()

	cloned := ln.deepClone().(*leafNode)

	testSameLeafNodeContent(t, ln, cloned)
}

func TestLeafNode_deepCloneWithNilKeyShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}
	ln.dirty = true
	ln.hash = getRandomByteSlice()
	ln.Value = getRandomByteSlice()
	ln.Key = nil

	cloned := ln.deepClone().(*leafNode)

	testSameLeafNodeContent(t, ln, cloned)
}

func TestLeafNode_deepCloneShouldWork(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}
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
