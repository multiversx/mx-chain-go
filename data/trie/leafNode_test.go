package trie

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
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

	marsh, hasher := getTestMarshalizerAndHasher()
	expectedLn := &leafNode{
		CollapsedLn: CollapsedLn{
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

	ln := getLn(getTestMarshalizerAndHasher())

	collapsed, err := ln.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, ln, collapsed)
}

func TestLeafNode_setHash(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	hash, _ := encodeNodeAndGetHash(ln)

	err := ln.setHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, ln.hash)
}

func TestLeafNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{}}

	err := ln.setHash()
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
	assert.Nil(t, ln.hash)
}

func TestLeafNode_setHashNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	err := ln.setHash()
	assert.True(t, errors.Is(err, ErrNilLeafNode))
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

	ln := getLn(getTestMarshalizerAndHasher())

	assert.Nil(t, ln.hashChildren())
}

func TestLeafNode_hashNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	expectedHash, _ := encodeNodeAndGetHash(ln)

	hash, err := ln.hashNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestLeafNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	hash, err := ln.hashNode()
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
	assert.Nil(t, hash)
}

func TestLeafNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	hash, err := ln.hashNode()
	assert.True(t, errors.Is(err, ErrNilLeafNode))
	assert.Nil(t, hash)
}

func TestLeafNode_commit(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	hash, _ := encodeNodeAndGetHash(ln)
	_ = ln.setHash()

	err := ln.commit(false, 0, 5, db, db)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	n, _ := decodeNode(encNode, ln.marsh, ln.hasher)
	ln = getLn(ln.marsh, ln.hasher)
	ln.dirty = false
	assert.Equal(t, ln, n)
}

func TestLeafNode_commitEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	err := ln.commit(false, 0, 5, nil, nil)
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
}

func TestLeafNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	err := ln.commit(false, 0, 5, nil, nil)
	assert.True(t, errors.Is(err, ErrNilLeafNode))
}

func TestLeafNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
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
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
	assert.Nil(t, encNode)
}

func TestLeafNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	encNode, err := ln.getEncodedNode()
	assert.True(t, errors.Is(err, ErrNilLeafNode))
	assert.Nil(t, encNode)
}

func TestLeafNode_resolveCollapsed(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())

	assert.Nil(t, ln.resolveCollapsed(0, nil))
}

func TestLeafNode_isCollapsed(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	assert.False(t, ln.isCollapsed())
}

func TestLeafNode_tryGet(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	key := []byte("dog")

	val, err := ln.tryGet(key, nil)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestLeafNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
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
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
	assert.Nil(t, val)
}

func TestLeafNode_tryGetNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	key := []byte("dog")

	val, err := ln.tryGet(key, nil)
	assert.True(t, errors.Is(err, ErrNilLeafNode))
	assert.Nil(t, val)
}

func TestLeafNode_getNext(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	key := []byte("dog")

	n, key, err := ln.getNext(key, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Nil(t, err)
}

func TestLeafNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	wrongKey := append([]byte{2}, []byte("dog")...)

	n, key, err := ln.getNext(wrongKey, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestLeafNode_getNextNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	key := []byte("dog")

	n, key, err := ln.getNext(key, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.True(t, errors.Is(err, ErrNilLeafNode))
}

func TestLeafNode_insertAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	key := []byte("dog")
	expectedVal := []byte("dogs")
	n, _ := newLeafNode(key, expectedVal, ln.marsh, ln.hasher)

	dirty, newNode, _, err := ln.insert(n, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, _ := newNode.tryGet(key, nil)
	assert.Equal(t, expectedVal, val)
}

func TestLeafNode_insertAtDifferentKey(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()

	lnKey := []byte{2, 100, 111, 103}
	ln, _ := newLeafNode(lnKey, []byte("dog"), marsh, hasher)

	nodeKey := []byte{3, 4, 5}
	nodeVal := []byte{3, 4, 5}
	n, _ := newLeafNode(nodeKey, nodeVal, marsh, hasher)

	dirty, newNode, _, err := ln.insert(n, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, _ := newNode.tryGet(nodeKey, nil)
	assert.Equal(t, nodeVal, val)
	assert.IsType(t, &branchNode{}, newNode)
}

func TestLeafNode_insertInStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	n, _ := newLeafNode([]byte("dog"), []byte("dogs"), ln.marsh, ln.hasher)
	_ = ln.commit(false, 0, 5, db, db)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.insert(n, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_insertInStoredLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	marsh, hasher := getTestMarshalizerAndHasher()
	ln, _ := newLeafNode([]byte{1, 2, 3}, []byte("dog"), marsh, hasher)
	n, _ := newLeafNode([]byte{4, 5, 6}, []byte("dogs"), marsh, hasher)
	_ = ln.commit(false, 0, 5, db, db)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.insert(n, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_insertInDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	n, _ := newLeafNode([]byte("dog"), []byte("dogs"), ln.marsh, ln.hasher)

	dirty, _, oldHashes, err := ln.insert(n, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInDirtyLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	ln, _ := newLeafNode([]byte{1, 2, 3}, []byte("dog"), marsh, hasher)
	n, _ := newLeafNode([]byte{4, 5, 6}, []byte("dogs"), marsh, hasher)

	dirty, _, oldHashes, err := ln.insert(n, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	dirty, newNode, _, err := ln.insert(&leafNode{}, nil)
	assert.False(t, dirty)
	assert.True(t, errors.Is(err, ErrNilLeafNode))
	assert.Nil(t, newNode)
}

func TestLeafNode_deletePresent(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())

	dirty, newNode, _, err := ln.delete([]byte("dog"), nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Nil(t, newNode)
}

func TestLeafNode_deleteFromStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commit(false, 0, 5, db, db)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_deleteFromLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commit(false, 0, 5, db, db)
	wrongKey := []byte{1, 2, 3}

	dirty, _, oldHashes, err := ln.delete(wrongKey, db)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_deleteFromDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_deleteNotPresent(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	wrongKey := []byte{1, 2, 3}

	dirty, newNode, _, err := ln.delete(wrongKey, nil)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newNode)
}

func TestLeafNode_reduceNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	ln, _ := newLeafNode([]byte{100, 111, 103}, nil, marsh, hasher)
	expected, _ := newLeafNode([]byte{2, 100, 111, 103}, nil, marsh, hasher)
	expected.dirty = true

	n, newChildHash, err := ln.reduceNode(2)
	assert.Equal(t, expected, n)
	assert.Nil(t, err)
	assert.True(t, newChildHash)
}

func TestLeafNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}
	assert.Equal(t, ErrEmptyLeafNode, ln.isEmptyOrNil())

	ln = nil
	assert.Equal(t, ErrNilLeafNode, ln.isEmptyOrNil())
}

func TestLeafNode_getChildren(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())

	children, err := ln.getChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(children))
}

func TestLeafNode_isValid(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
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

	marsh, hasher := getTestMarshalizerAndHasher()
	tr := initTrie()
	nodes, hashes := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := lrucache.NewCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], marsh, hasher)
		nodesCacher.Put(n.hash, n, len(n.EncodedNode()))
	}

	lnPosition := 5
	ln := &leafNode{baseNode: &baseNode{hash: hashes[lnPosition]}}
	missing, _, err := ln.loadChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 6, nodesCacher.Len())
	assert.Equal(t, 0, len(missing))
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

func TestInsertSameNodeShouldNotSetDirtyBnRoot(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash := tr.root.getHash()

	_ = tr.Update([]byte("dog"), []byte("puppy"))
	assert.False(t, tr.root.isDirty())
	assert.Equal(t, rootHash, tr.root.getHash())
	assert.Equal(t, [][]byte{}, tr.oldHashes)
}

func TestInsertSameNodeShouldNotSetDirtyEnRoot(t *testing.T) {
	t.Parallel()

	tr, _, _ := newEmptyTrie()
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("log"), []byte("wood"))
	_ = tr.Commit()
	rootHash := tr.root.getHash()

	_ = tr.Update([]byte("dog"), []byte("puppy"))
	assert.False(t, tr.root.isDirty())
	assert.Equal(t, rootHash, tr.root.getHash())
	assert.Equal(t, [][]byte{}, tr.oldHashes)
}

func TestInsertSameNodeShouldNotSetDirtyLnRoot(t *testing.T) {
	t.Parallel()

	tr, _, _ := newEmptyTrie()
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Commit()
	rootHash := tr.root.getHash()

	_ = tr.Update([]byte("dog"), []byte("puppy"))
	assert.False(t, tr.root.isDirty())
	assert.Equal(t, rootHash, tr.root.getHash())
	assert.Equal(t, [][]byte{}, tr.oldHashes)
}

func TestLeafNode_deleteDifferentKeyShouldNotModifyTrie(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash := tr.root.getHash()

	_ = tr.Update([]byte("ddoe"), []byte{})
	assert.False(t, tr.root.isDirty())
	assert.Equal(t, rootHash, tr.root.getHash())
	assert.Equal(t, [][]byte{}, tr.oldHashes)
}

func TestLeafNode_newLeafNodeNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	ln, err := newLeafNode([]byte("key"), []byte("val"), nil, mock.HasherMock{})
	assert.Nil(t, ln)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestLeafNode_newLeafNodeNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	ln, err := newLeafNode([]byte("key"), []byte("val"), &mock.MarshalizerMock{}, nil)
	assert.Nil(t, ln)
	assert.Equal(t, ErrNilHasher, err)
}

func TestLeafNode_newLeafNodeOkVals(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	key := []byte("key")
	val := []byte("val")
	ln, err := newLeafNode(key, val, marsh, hasher)

	assert.Nil(t, err)
	assert.Equal(t, key, ln.Key)
	assert.Equal(t, val, ln.Value)
	assert.Equal(t, hasher, ln.hasher)
	assert.Equal(t, marsh, ln.marsh)
	assert.True(t, ln.dirty)
}

func TestLeafNode_getMarshalizer(t *testing.T) {
	t.Parallel()

	marsh, _ := getTestMarshalizerAndHasher()
	ln := &leafNode{
		baseNode: &baseNode{
			marsh: marsh,
		},
	}

	assert.Equal(t, marsh, ln.getMarshalizer())
}

func TestLeafNode_getAllHashes(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	hashes, err := ln.getAllHashes(mock.NewMemDbMock())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hashes))
	assert.Equal(t, ln.hash, hashes[0])
}
