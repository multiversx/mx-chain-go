package trie

import (
	"bytes"
	"context"
	"errors"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/trieMetricsCollector"
	"github.com/stretchr/testify/assert"
)

func getLn(marsh marshal.Marshalizer, hasher hashing.Hasher) *leafNode {
	newLn, _ := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), marsh, hasher)
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
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), marsh, hasher)
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

	db := testscommon.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	hash, _ := encodeNodeAndGetHash(ln)
	_ = ln.setHash()

	err := ln.commitDirty(db, db, dtmc)
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

	err := ln.commitDirty(nil, nil, nil)
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
}

func TestLeafNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	err := ln.commitDirty(nil, nil, nil)
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

	assert.Nil(t, ln.resolveCollapsed(0, nil, nil))
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

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := ln.tryGet(key, tmc, nil)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestLeafNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	wrongKey := []byte{1, 2, 3}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := ln.tryGet(wrongKey, tmc, nil)
	assert.Nil(t, val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestLeafNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	key := []byte("dog")
	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := ln.tryGet(key, tmc, nil)
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestLeafNode_tryGetNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	key := []byte("dog")

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := ln.tryGet(key, tmc, nil)
	assert.True(t, errors.Is(err, ErrNilLeafNode))
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestLeafNode_getNext(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	key := []byte("dog")

	n, key, err := ln.getNext(key, nil, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Nil(t, err)
}

func TestLeafNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	wrongKey := append([]byte{2}, []byte("dog")...)

	n, key, err := ln.getNext(wrongKey, nil, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestLeafNode_getNextNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	key := []byte("dog")

	n, key, err := ln.getNext(key, nil, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.True(t, errors.Is(err, ErrNilLeafNode))
}

func TestLeafNode_insertAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	key := "dog"
	expectedVal := "dogs"

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, _, err := ln.insert(getTrieDataWithDefaultVersion(key, expectedVal), tmc, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, 1, tmc.GetSizeLoadedInMem())

	val, _ := newNode.tryGet([]byte(key), dtmc, nil)
	assert.Equal(t, []byte(expectedVal), val)
}

func TestLeafNode_insertAtDifferentKey(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()

	lnKey := []byte{2, 100, 111, 103}
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion(string(lnKey), "dog"), marsh, hasher)

	nodeKey := []byte{3, 4, 5}
	nodeVal := []byte{3, 4, 5}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, _, err := ln.insert(getTrieDataWithDefaultVersion(string(nodeKey), string(nodeVal)), tmc, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	expectedSize := newNode.sizeInBytes() + newNode.(*branchNode).children[3].sizeInBytes() - 1
	assert.Equal(t, expectedSize, tmc.GetSizeLoadedInMem())

	val, _ := newNode.tryGet(nodeKey, trieMetricsCollector.NewTrieMetricsCollector(), nil)
	assert.Equal(t, nodeVal, val)
	assert.IsType(t, &branchNode{}, newNode)
}

func TestLeafNode_insertInStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commitDirty(db, db, dtmc)
	lnHash := ln.getHash()

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, oldHashes, err := ln.insert(getTrieDataWithDefaultVersion("dog", "dogs"), tmc, db)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
	assert.Equal(t, 1, tmc.GetSizeLoadedInMem())
}

func TestLeafNode_insertInStoredLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	marsh, hasher := getTestMarshalizerAndHasher()
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3}), "dog"), marsh, hasher)
	_ = ln.commitDirty(db, db, dtmc)
	lnHash := ln.getHash()

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, oldHashes, err := ln.insert(getTrieDataWithDefaultVersion(string([]byte{4, 5, 6}), "dogs"), tmc, db)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
	expectedSize := newNode.sizeInBytes() + newNode.(*branchNode).children[4].sizeInBytes() - 1
	assert.Equal(t, expectedSize, tmc.GetSizeLoadedInMem())
}

func TestLeafNode_insertInDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())

	newNode, oldHashes, err := ln.insert(getTrieDataWithDefaultVersion("dog", "dogs"), dtmc, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInDirtyLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3}), "dog"), marsh, hasher)

	newNode, oldHashes, err := ln.insert(getTrieDataWithDefaultVersion(string([]byte{4, 5, 6}), "dogs"), dtmc, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_insertInNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	newNode, _, err := ln.insert(getTrieDataWithDefaultVersion("dog", "dogs"), nil, nil)
	assert.Nil(t, newNode)
	assert.True(t, errors.Is(err, ErrNilLeafNode))
	assert.Nil(t, newNode)
}

func TestLeafNode_deletePresent(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, newNode, _, err := ln.delete([]byte("dog"), tmc, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Nil(t, newNode)
	assert.Equal(t, -ln.sizeInBytes(), tmc.GetSizeLoadedInMem())
}

func TestLeafNode_deleteFromStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commitDirty(db, db, dtmc)
	lnHash := ln.getHash()

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), dtmc, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{lnHash}, oldHashes)
}

func TestLeafNode_deleteFromLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commitDirty(db, db, nil)
	wrongKey := []byte{1, 2, 3}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, _, oldHashes, err := ln.delete(wrongKey, tmc, db)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
	assert.Equal(t, 0, tmc.GetSizeLoadedInMem())
}

func TestLeafNode_deleteFromDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())

	dirty, _, oldHashes, err := ln.delete([]byte("dog"), dtmc, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestLeafNode_deleteNotPresent(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	wrongKey := []byte{1, 2, 3}

	dirty, newNode, _, err := ln.delete(wrongKey, dtmc, nil)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newNode)
}

func TestLeafNode_reduceNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{100, 111, 103}), ""), marsh, hasher)
	expected, _ := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{2, 100, 111, 103}), ""), marsh, hasher)
	expected.dirty = true

	n, newChildHash, err := ln.reduceNode(2, dtmc)
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

	children, err := ln.getChildren(nil, nil)
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

	_, hasher := getTestMarshalizerAndHasher()
	tr := initTrie()
	nodes, hashes := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := cache.NewLRUCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], hasher)
		nodesCacher.Put(n.hash, n, len(n.GetSerialized()))
	}

	lnPosition := 5
	ln := &leafNode{baseNode: &baseNode{hash: hashes[lnPosition]}}
	missing, _, err := ln.loadChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 6, nodesCacher.Len())
	assert.Equal(t, 0, len(missing))
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

	tr, _ := newEmptyTrie()
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

	tr, _ := newEmptyTrie()
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

	ln, err := newLeafNode(getTrieDataWithDefaultVersion("key", "val"), nil, &hashingMocks.HasherMock{})
	assert.Nil(t, ln)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestLeafNode_newLeafNodeNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	ln, err := newLeafNode(getTrieDataWithDefaultVersion("key", "val"), &marshallerMock.MarshalizerMock{}, nil)
	assert.Nil(t, ln)
	assert.Equal(t, ErrNilHasher, err)
}

func TestLeafNode_newLeafNodeOkVals(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	key := []byte("key")
	val := []byte("val")
	ln, err := newLeafNode(getTrieDataWithDefaultVersion("key", "val"), marsh, hasher)

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
	hashes, err := ln.getAllHashes(nil, testscommon.NewMemDbMock())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hashes))
	assert.Equal(t, ln.hash, hashes[0])
}

func TestLeafNode_getNextHashAndKey(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	proofVerified, nextHash, nextKey := ln.getNextHashAndKey([]byte("dog"))

	assert.True(t, proofVerified)
	assert.Nil(t, nextHash)
	assert.Nil(t, nextKey)
}

func TestLeafNode_getNextHashAndKeyNilNode(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	proofVerified, nextHash, nextKey := ln.getNextHashAndKey([]byte("dog"))

	assert.False(t, proofVerified)
	assert.Nil(t, nextHash)
	assert.Nil(t, nextKey)
}

func TestLeafNode_SizeInBytes(t *testing.T) {
	t.Parallel()

	var ln *leafNode
	assert.Equal(t, 0, ln.sizeInBytes())

	value := []byte("value")
	key := []byte("key")
	hash := bytes.Repeat([]byte{1}, 32)
	ln = &leafNode{
		CollapsedLn: CollapsedLn{
			Key:   key,
			Value: value,
		},
		baseNode: &baseNode{
			hash:   hash,
			dirty:  false,
			marsh:  nil,
			hasher: nil,
		},
	}
	assert.Equal(t, len(key)+len(value)+len(hash)+1+2*pointerSizeInBytes+nodeVersionSizeInBytes, ln.sizeInBytes())
}

func TestLeafNode_writeNodeOnChannel(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.setHash()
	leavesChannel := make(chan core.KeyValueHolder, 2)

	err := writeNodeOnChannel(ln, leavesChannel)
	assert.Nil(t, err)

	retrievedLn := <-leavesChannel
	assert.Equal(t, ln.getHash(), retrievedLn.Key())
	assert.Equal(t, ln.Value, retrievedLn.Value())
}

func TestLeafNode_commitContextDone(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	ln := getLn(marshalizer, hasherMock)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ln.commitSnapshot(db, nil, nil, ctx, statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)
	assert.Equal(t, core.ErrContextClosing, err)
}

func TestLeafNode_getValue(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	assert.Equal(t, ln.Value, ln.getValue())
}

func TestLeafNode_getVersion(t *testing.T) {
	t.Parallel()

	t.Run("invalid node version", func(t *testing.T) {
		t.Parallel()

		ln := getLn(getTestMarshalizerAndHasher())
		ln.Version = math.MaxUint8 + 1

		version, err := ln.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Equal(t, ErrInvalidNodeVersion, err)
	})

	t.Run("NotSpecified version", func(t *testing.T) {
		t.Parallel()

		ln := getLn(getTestMarshalizerAndHasher())

		version, err := ln.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Nil(t, err)
	})

	t.Run("AutoBalanceEnabled version", func(t *testing.T) {
		t.Parallel()

		ln := getLn(getTestMarshalizerAndHasher())
		ln.Version = uint32(core.AutoBalanceEnabled)

		version, err := ln.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
		assert.Nil(t, err)
	})
}

func TestLeafNode_getNodeData(t *testing.T) {
	t.Parallel()

	t.Run("nil node", func(t *testing.T) {
		t.Parallel()

		var ln *leafNode
		nodeData, err := ln.getNodeData(keyBuilder.NewDisabledKeyBuilder())
		assert.Nil(t, nodeData)
		assert.True(t, errors.Is(err, ErrNilLeafNode))
	})
	t.Run("gets data from child", func(t *testing.T) {
		t.Parallel()

		ln := getLn(getTestMarshalizerAndHasher())
		val := []byte("dog")

		kb := keyBuilder.NewKeyBuilder()
		nodeData, err := ln.getNodeData(kb)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(nodeData))

		assert.Equal(t, uint(3), nodeData[0].GetKeyBuilder().Size())
		assert.Equal(t, val, nodeData[0].GetData())
		assert.Equal(t, uint64(len(val)+len(val)), nodeData[0].Size())
		assert.True(t, nodeData[0].IsLeaf())
	})
}
