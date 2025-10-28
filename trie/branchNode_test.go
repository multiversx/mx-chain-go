package trie

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/trieMetricsCollector"
	"github.com/stretchr/testify/assert"
)

const tenMBSize = uint64(10485760)

var dtmc = trieMetricsCollector.NewDisabledTrieMetricsCollector()

func getTestMarshalizerAndHasher() (marshal.Marshalizer, hashing.Hasher) {
	marsh := &marshal.GogoProtoMarshalizer{}
	hash := &testscommon.KeccakMock{}
	return marsh, hash
}

func getTrieDataWithDefaultVersion(key string, val string) core.TrieData {
	return core.TrieData{
		Key:     []byte(key),
		Value:   []byte(val),
		Version: core.NotSpecified,
	}
}

func getBnAndCollapsedBn(marshalizer marshal.Marshalizer, hasher hashing.Hasher) (*branchNode, *branchNode) {
	var children [nrOfChildren]node
	EncodedChildren := make([][]byte, nrOfChildren)

	children[2], _ = newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), marshalizer, hasher)
	children[6], _ = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"), marshalizer, hasher)
	children[13], _ = newLeafNode(getTrieDataWithDefaultVersion("doge", "doge"), marshalizer, hasher)
	bn, _ := newBranchNode(marshalizer, hasher)
	bn.children = children

	EncodedChildren[2], _ = encodeNodeAndGetHash(children[2])
	EncodedChildren[6], _ = encodeNodeAndGetHash(children[6])
	EncodedChildren[13], _ = encodeNodeAndGetHash(children[13])
	collapsedBn, _ := newBranchNode(marshalizer, hasher)
	collapsedBn.EncodedChildren = EncodedChildren

	return bn, collapsedBn
}

func markNotDirtyBranchNode(bn *branchNode) {
	bn.dirty = false
	for i := range bn.children {
		if bn.children[i] != nil {
			bn.children[i].setDirty(false)
		}
	}
}

func emptyDirtyBranchNode() *branchNode {
	var children [nrOfChildren]node
	encChildren := make([][]byte, nrOfChildren)
	childrenVersion := make([]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: CollapsedBn{
			EncodedChildren: encChildren,
			ChildrenVersion: childrenVersion,
		},
		children: children,
		baseNode: &baseNode{
			dirty: true,
		},
	}
}

func newEmptyTrie() (*patriciaMerkleTrie, *trieStorageManager) {
	args := GetDefaultTrieStorageManagerParameters()
	trieStorage, _ := NewTrieStorageManager(args)
	tr := &patriciaMerkleTrie{
		trieStorage:         trieStorage,
		marshalizer:         args.Marshalizer,
		hasher:              args.Hasher,
		oldHashes:           make([][]byte, 0),
		oldRoot:             make([]byte, 0),
		chanClose:           make(chan struct{}),
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		collapseManager:     collapseManager.NewDisabledCollapseManager(),
	}

	return tr, trieStorage
}

func initTrie() *patriciaMerkleTrie {
	tr, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))

	return tr
}

func getEncodedTrieNodesAndHashes(tr common.Trie) ([][]byte, [][]byte) {
	it, _ := NewDFSIterator(tr)
	encNode, _ := it.MarshalizedNode()

	nodes := make([][]byte, 0)
	nodes = append(nodes, encNode)

	hashes := make([][]byte, 0)
	hash, _ := it.GetHash()
	hashes = append(hashes, hash)

	for it.HasNext() {
		_ = it.Next()
		encNode, _ = it.MarshalizedNode()

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

	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	collapsedBn.dirty = true

	collapsed, err := bn.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, collapsedBn, collapsed)
}

func TestBranchNode_getCollapsedEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	collapsed, err := bn.getCollapsed()
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
	assert.Nil(t, collapsed)
}

func TestBranchNode_getCollapsedNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	collapsed, err := bn.getCollapsed()
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, collapsed)
}

func TestBranchNode_getCollapsedCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	collapsed, err := collapsedBn.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, collapsedBn, collapsed)
}

func TestBranchNode_setHash(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	hash, _ := encodeNodeAndGetHash(collapsedBn)

	err := bn.setHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, bn.hash)
}

func TestBranchNode_setRootHash(t *testing.T) {
	t.Parallel()

	marsh, hsh := getTestMarshalizerAndHasher()

	trieStorage1, _ := NewTrieStorageManager(GetDefaultTrieStorageManagerParameters())
	trieStorage2, _ := NewTrieStorageManager(GetDefaultTrieStorageManagerParameters())

	tr1, _ := NewTrie(trieStorage1, marsh, hsh, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, collapseManager.NewDisabledCollapseManager())
	tr2, _ := NewTrie(trieStorage2, marsh, hsh, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, collapseManager.NewDisabledCollapseManager())

	maxIterations := 10000
	for i := 0; i < maxIterations; i++ {
		val := hsh.Compute(fmt.Sprint(i))
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

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	hash, _ := encodeNodeAndGetHash(collapsedBn)

	err := collapsedBn.setRootHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, collapsedBn.hash)
}

func TestBranchNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	err := bn.setHash()
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
	assert.Nil(t, bn.hash)
}

func TestBranchNode_setHashNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.setHash()
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, bn)
}

func TestBranchNode_setHashCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
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

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

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
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
}

func TestBranchNode_hashChildrenNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.hashChildren()
	assert.True(t, errors.Is(err, ErrNilBranchNode))
}

func TestBranchNode_hashChildrenCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	err := collapsedBn.hashChildren()
	assert.Nil(t, err)

	_, collapsedBn2 := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	assert.Equal(t, collapsedBn2, collapsedBn)
}

func TestBranchNode_hashNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	expectedHash, _ := encodeNodeAndGetHash(collapsedBn)

	hash, err := collapsedBn.hashNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestBranchNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	hash, err := bn.hashNode()
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
	assert.Nil(t, hash)
}

func TestBranchNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	hash, err := bn.hashNode()
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, hash)
}

func TestBranchNode_commit(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	marsh, hasher := getTestMarshalizerAndHasher()
	bn, collapsedBn := getBnAndCollapsedBn(marsh, hasher)

	hash, _ := encodeNodeAndGetHash(collapsedBn)
	_ = bn.setHash()

	err := bn.commitDirty(db, db, dtmc)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	n, _ := decodeNode(encNode, marsh, hasher)
	h1, _ := encodeNodeAndGetHash(collapsedBn)
	h2, _ := encodeNodeAndGetHash(n)
	assert.Equal(t, h1, h2)
}

func TestBranchNode_commitEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	err := bn.commitDirty(nil, nil, nil)
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
}

func TestBranchNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.commitDirty(nil, nil, nil)
	assert.True(t, errors.Is(err, ErrNilBranchNode))
}

func TestBranchNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

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
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
	assert.Nil(t, encNode)
}

func TestBranchNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	encNode, err := bn.getEncodedNode()
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, encNode)
}

func TestBranchNode_resolveCollapsed(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)

	_ = bn.setHash()
	_ = bn.commitDirty(db, db, dtmc)
	resolved, _ := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), bn.marsh, bn.hasher)
	resolved.dirty = false
	resolved.hash = bn.EncodedChildren[childPos]

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	err := collapsedBn.resolveCollapsed(childPos, tmc, db)
	assert.Nil(t, err)
	assert.Equal(t, resolved, collapsedBn.children[childPos])
	assert.Equal(t, collapsedBn.children[childPos].sizeInBytes(), tmc.GetSizeLoadedInMem())
}

func TestBranchNode_resolveCollapsedEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	err := bn.resolveCollapsed(2, nil, nil)
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
}

func TestBranchNode_resolveCollapsedENilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.resolveCollapsed(2, nil, nil)
	assert.True(t, errors.Is(err, ErrNilBranchNode))
}

func TestBranchNode_resolveCollapsedPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	err := bn.resolveCollapsed(17, nil, nil)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_isCollapsed(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	assert.True(t, collapsedBn.isCollapsed())
	assert.False(t, bn.isCollapsed())

	collapsedBn.children[2], _ = newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), bn.marsh, bn.hasher)
	assert.False(t, collapsedBn.isCollapsed())
}

func TestBranchNode_tryGet(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := bn.tryGet(key, tmc, nil)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), tmc.GetMaxDepth())
}

func TestBranchNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	var key []byte

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := bn.tryGet(key, tmc, nil)
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestBranchNode_tryGetChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	key := []byte("dog")

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := bn.tryGet(key, tmc, nil)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestBranchNode_tryGetNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nilChildKey := []byte{3}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := bn.tryGet(nilChildKey, tmc, nil)
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestBranchNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	_ = bn.setHash()
	_ = bn.commitDirty(db, db, dtmc)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := collapsedBn.tryGet(key, tmc, db)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), tmc.GetMaxDepth())
}

func TestBranchNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := bn.tryGet(key, tmc, nil)
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestBranchNode_tryGetNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	val, err := bn.tryGet(key, tmc, nil)
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), tmc.GetMaxDepth())
}

func TestBranchNode_getNext(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nextNode, _ := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), bn.marsh, bn.hasher)
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	n, key, err := bn.getNext(key, dtmc, nil)

	h1, _ := encodeNodeAndGetHash(nextNode)
	h2, _ := encodeNodeAndGetHash(n)
	assert.Equal(t, h1, h2)
	assert.Equal(t, []byte("dog"), key)
	assert.Nil(t, err)
}

func TestBranchNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	key := []byte("dog")

	n, key, err := bn.getNext(key, dtmc, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_getNextNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nilChildPos := byte(4)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	n, key, err := bn.getNext(key, dtmc, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestBranchNode_insert(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nodeKey := []byte{0, 2, 3}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newBn, _, err := bn.insert(getTrieDataWithDefaultVersion(string(nodeKey), "dogs"), tmc, nil)
	assert.NotNil(t, newBn)
	assert.Nil(t, err)
	assert.Equal(t, bn.children[0].sizeInBytes(), tmc.GetSizeLoadedInMem())

	nodeKeyRemainder := nodeKey[1:]
	bn.children[0], _ = newLeafNode(getTrieDataWithDefaultVersion(string(nodeKeyRemainder), "dogs"), bn.marsh, bn.hasher)
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_insertEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	newBn, _, err := bn.insert(getTrieDataWithDefaultVersion("", "dogs"), dtmc, nil)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	newBn, _, err := bn.insert(getTrieDataWithDefaultVersion("dog", "dogs"), dtmc, nil)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	originalSize := bn.children[childPos].sizeInBytes()
	_ = bn.setHash()
	_ = bn.commitDirty(db, db, dtmc)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newBn, _, err := collapsedBn.insert(getTrieDataWithDefaultVersion(string(key), "dogs"), tmc, db)
	assert.NotNil(t, newBn)
	assert.Nil(t, err)

	newSize := collapsedBn.children[childPos].sizeInBytes()
	sizeDiff := newSize - originalSize
	assert.Equal(t, originalSize+sizeDiff, tmc.GetSizeLoadedInMem())

	val, _ := newBn.tryGet(key, trieMetricsCollector.NewTrieMetricsCollector(), db)
	assert.Equal(t, []byte("dogs"), val)
}

func TestBranchNode_insertInStoredBnOnExistingPos(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	originalSize := bn.children[childPos].sizeInBytes()
	_ = bn.setHash()
	markNotDirtyBranchNode(bn)
	bnHash := bn.getHash()
	lnHash := bn.EncodedChildren[childPos]
	expectedHashes := [][]byte{lnHash, bnHash}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, oldHashes, err := bn.insert(getTrieDataWithDefaultVersion(string(key), "dogs"), tmc, db)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)

	newSize := bn.children[childPos].sizeInBytes()
	assert.Equal(t, newSize-originalSize, tmc.GetSizeLoadedInMem())
}

func TestBranchNode_insertInStoredBnOnNilPos(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	_ = bn.setHash()
	markNotDirtyBranchNode(bn)
	bnHash := bn.getHash()
	expectedHashes := [][]byte{bnHash}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, oldHashes, err := bn.insert(getTrieDataWithDefaultVersion(string(key), "dogs"), tmc, db)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
	assert.Equal(t, bn.children[nilChildPos].sizeInBytes(), tmc.GetSizeLoadedInMem())
}

func TestBranchNode_insertInDirtyBnOnNilPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, oldHashes, err := bn.insert(getTrieDataWithDefaultVersion(string(key), "dogs"), tmc, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
	assert.Equal(t, bn.children[nilChildPos].sizeInBytes(), tmc.GetSizeLoadedInMem())
}

func TestBranchNode_insertInDirtyBnOnExistingPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	originalSize := bn.children[childPos].sizeInBytes()
	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	newNode, oldHashes, err := bn.insert(getTrieDataWithDefaultVersion(string(key), "dogs"), tmc, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
	newSize := bn.children[childPos].sizeInBytes()
	assert.Equal(t, newSize-originalSize, tmc.GetSizeLoadedInMem())
}

func TestBranchNode_insertInNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	newBn, _, err := bn.insert(getTrieDataWithDefaultVersion("key", "dogs"), dtmc, nil)
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, newBn)
}

func TestBranchNode_delete(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	var children [nrOfChildren]node
	children[6], _ = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"), bn.marsh, bn.hasher)
	children[13], _ = newLeafNode(getTrieDataWithDefaultVersion("doge", "doge"), bn.marsh, bn.hasher)
	expectedBn, _ := newBranchNode(bn.marsh, bn.hasher)
	expectedBn.children = children

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	originalSize := bn.children[childPos].sizeInBytes()
	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, newBn, _, err := bn.delete(key, tmc, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, -(originalSize + hashSizeInBytes), tmc.GetSizeLoadedInMem())

	_ = expectedBn.setHash()
	_ = newBn.setHash()
	assert.Equal(t, expectedBn.getHash(), newBn.getHash())
}

func TestBranchNode_deleteFromStoredBn(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)

	originalSize := bn.children[childPos].sizeInBytes()
	_ = bn.setHash()
	markNotDirtyBranchNode(bn)
	bnHash := bn.getHash()
	lnHash := bn.EncodedChildren[childPos]
	expectedHashes := [][]byte{lnHash, bnHash}

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, _, oldHashes, err := bn.delete(lnKey, tmc, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
	assert.Equal(t, -(originalSize + hashSizeInBytes), tmc.GetSizeLoadedInMem())
}

func TestBranchNode_deleteFromDirtyBn(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)

	originalSize := bn.children[childPos].sizeInBytes()
	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, _, oldHashes, err := bn.delete(lnKey, tmc, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
	assert.Equal(t, -(originalSize + hashSizeInBytes), tmc.GetSizeLoadedInMem())
}

func TestBranchNode_deleteEmptyNode(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, newBn, _, err := bn.delete(key, tmc, nil)
	assert.False(t, dirty)
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
	assert.Nil(t, newBn)
	assert.Equal(t, 0, tmc.GetSizeLoadedInMem())
}

func TestBranchNode_deleteNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, newBn, _, err := bn.delete(key, tmc, nil)
	assert.False(t, dirty)
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, newBn)
	assert.Equal(t, 0, tmc.GetSizeLoadedInMem())
}

func TestBranchNode_deleteNonexistentNodeFromChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("butterfly")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, newBn, _, err := bn.delete(key, tmc, nil)
	assert.False(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, bn, newBn)
	assert.Equal(t, 0, tmc.GetSizeLoadedInMem())
}

func TestBranchNode_deleteEmptykey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, newBn, _, err := bn.delete([]byte{}, tmc, nil)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
	assert.Equal(t, 0, tmc.GetSizeLoadedInMem())
}

func TestBranchNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	_ = bn.setHash()
	_ = bn.commitDirty(db, db, dtmc)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	tmc := trieMetricsCollector.NewTrieMetricsCollector()
	dirty, newBn, _, err := collapsedBn.delete(key, tmc, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, -hashSizeInBytes, tmc.GetSizeLoadedInMem())

	val, err := newBn.tryGet(key, trieMetricsCollector.NewTrieMetricsCollector(), db)
	assert.Nil(t, val)
	assert.Nil(t, err)
}

func TestBranchNode_deleteAndReduceBn(t *testing.T) {
	t.Parallel()

	bn, _ := newBranchNode(getTestMarshalizerAndHasher())
	var children [nrOfChildren]node
	firstChildPos := byte(2)
	secondChildPos := byte(6)
	children[firstChildPos], _ = newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), bn.marsh, bn.hasher)
	children[secondChildPos], _ = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"), bn.marsh, bn.hasher)
	bn.children = children
	bn.EncodedChildren[firstChildPos], _ = encodeNodeAndGetHash(children[firstChildPos])
	bn.EncodedChildren[secondChildPos], _ = encodeNodeAndGetHash(children[secondChildPos])
	extraLeafData := 1
	expectedSizeInMem := -bn.children[secondChildPos].sizeInBytes() - bn.sizeInBytes() + extraLeafData

	key := append([]byte{firstChildPos}, []byte("dog")...)
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion(string(key), "dog"), bn.marsh, bn.hasher)
	tmc := trieMetricsCollector.NewTrieMetricsCollector()

	key = append([]byte{secondChildPos}, []byte("doe")...)
	dirty, newBn, _, err := bn.delete(key, tmc, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newBn)
	assert.Equal(t, expectedSizeInMem, tmc.GetSizeLoadedInMem())
}

func TestBranchNode_reduceNode(t *testing.T) {
	t.Parallel()

	bn, _ := newBranchNode(getTestMarshalizerAndHasher())
	var children [nrOfChildren]node
	childPos := byte(2)
	children[childPos], _ = newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), bn.marsh, bn.hasher)
	bn.children = children

	key := append([]byte{childPos}, []byte("dog")...)
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion(string(key), "dog"), bn.marsh, bn.hasher)

	n, newChildHash, err := bn.children[childPos].reduceNode(int(childPos), dtmc)
	assert.Equal(t, ln, n)
	assert.Nil(t, err)
	assert.True(t, newChildHash)
}

func TestBranchNode_getChildPosition(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nr, pos := getChildPosition(bn)
	assert.Equal(t, 3, nr)
	assert.Equal(t, 13, pos)
}

func TestBranchNode_clone(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	clone := bn.clone()
	assert.False(t, bn == clone)
	assert.Equal(t, bn, clone)
}

func TestBranchNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()
	assert.Equal(t, ErrEmptyBranchNode, bn.isEmptyOrNil())

	bn = nil
	assert.Equal(t, ErrNilBranchNode, bn.isEmptyOrNil())
}

func TestReduceBranchNodeWithExtensionNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	expectedTr, _ := newEmptyTrie()

	_ = expectedTr.Update([]byte("dog"), []byte("dog"))
	_ = expectedTr.Update([]byte("doll"), []byte("doll"))

	_ = tr.Update([]byte("dog"), []byte("dog"))
	_ = tr.Update([]byte("doll"), []byte("doll"))
	_ = tr.Update([]byte("wolf"), []byte("wolf"))
	_ = tr.Delete([]byte("wolf"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithBranchNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	expectedTr, _ := newEmptyTrie()

	_ = expectedTr.Update([]byte("dog"), []byte("puppy"))
	_ = expectedTr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Delete([]byte("doe"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithLeafNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	expectedTr, _ := newEmptyTrie()

	_ = expectedTr.Update([]byte("doe"), []byte("reindeer"))
	_ = expectedTr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Delete([]byte("dog"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithLeafNodeValueShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	expectedTr, _ := newEmptyTrie()

	_ = expectedTr.Update([]byte("doe"), []byte("reindeer"))
	_ = expectedTr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Delete([]byte("dogglesworth"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()

	assert.Equal(t, expectedHash, hash)
}

func TestBranchNode_getChildren(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	children, err := bn.getChildren(dtmc, nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(children))
}

func TestBranchNode_getChildrenCollapsedBn(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	_ = bn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)

	children, err := collapsedBn.getChildren(dtmc, db)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(children))
}

func TestBranchNode_isValid(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
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

	marsh, hasher := getTestMarshalizerAndHasher()
	tr := initTrie()
	_ = tr.root.setRootHash()
	nodes, _ := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := cache.NewLRUCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], hasher)
		nodesCacher.Put(n.hash, n, len(n.GetSerialized()))
	}

	firstChildIndex := 5
	secondChildIndex := 7

	bn := getCollapsedBn(t, tr.root)

	getNode := func(hash []byte) (node, error) {
		cacheData, _ := nodesCacher.Get(hash)
		return trieNode(cacheData, marsh, hasher)
	}

	missing, _, err := bn.loadChildren(getNode)
	assert.Nil(t, err)
	assert.NotNil(t, bn.children[firstChildIndex])
	assert.NotNil(t, bn.children[secondChildIndex])
	assert.Equal(t, 0, len(missing))
	assert.Equal(t, 6, nodesCacher.Len())
}

func getCollapsedBn(t *testing.T, n node) *branchNode {
	bn, ok := n.(*branchNode)
	assert.True(t, ok)
	for i := 0; i < nrOfChildren; i++ {
		bn.children[i] = nil
	}
	return bn
}

func TestPatriciaMerkleTrie_CommitCollapsedDirtyTrieShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	_ = tr.Update([]byte("aaa"), []byte("aaa"))
	_ = tr.Update([]byte("nnn"), []byte("nnn"))
	_ = tr.Update([]byte("zzz"), []byte("zzz"))
	_ = tr.Commit()

	tr.root, _ = tr.root.getCollapsed()
	_ = tr.Delete([]byte("zzz"))

	assert.True(t, tr.root.isDirty())
	assert.True(t, tr.root.isCollapsed())

	_ = tr.Commit()

	assert.False(t, tr.root.isDirty())
}

func BenchmarkMarshallNodeJson(b *testing.B) {
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	marsh := marshal.JsonMarshalizer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = marsh.Marshal(bn)
	}
}

func TestBranchNode_newBranchNodeNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	bn, err := newBranchNode(nil, &hashingMocks.HasherMock{})
	assert.Nil(t, bn)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestBranchNode_newBranchNodeNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	bn, err := newBranchNode(&marshallerMock.MarshalizerMock{}, nil)
	assert.Nil(t, bn)
	assert.Equal(t, ErrNilHasher, err)
}

func TestBranchNode_newBranchNodeOkVals(t *testing.T) {
	t.Parallel()

	var children [nrOfChildren]node
	marsh, hasher := getTestMarshalizerAndHasher()
	bn, err := newBranchNode(marsh, hasher)

	assert.Nil(t, err)
	assert.Equal(t, make([][]byte, nrOfChildren), bn.EncodedChildren)
	assert.Equal(t, children, bn.children)
	assert.Equal(t, marsh, bn.marsh)
	assert.Equal(t, hasher, bn.hasher)
	assert.True(t, bn.dirty)
}

func TestBranchNode_getMarshalizer(t *testing.T) {
	t.Parallel()

	expectedMarsh := &marshallerMock.MarshalizerMock{}
	bn := &branchNode{
		baseNode: &baseNode{
			marsh: expectedMarsh,
		},
	}

	marsh := bn.getMarshalizer()
	assert.Equal(t, expectedMarsh, marsh)
}

func TestBranchNode_setRootHashCollapsedChildren(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	bn := &branchNode{
		baseNode: &baseNode{
			marsh:  marsh,
			hasher: hasher,
		},
	}

	_, collapsedBn := getBnAndCollapsedBn(marsh, hasher)
	_, collapsedEn := getEnAndCollapsedEn()
	collapsedLn := getLn(marsh, hasher)

	bn.children[0] = collapsedBn
	bn.children[1] = collapsedEn
	bn.children[2] = collapsedLn

	err := bn.setRootHash()
	assert.Nil(t, err)
}

func TestBranchNode_reduceNodeBnChild(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	en, _ := getEnAndCollapsedEn()
	pos := 5
	expectedNode, _ := newExtensionNode([]byte{byte(pos)}, en.child, marsh, hasher)

	newNode, newChildHash, err := en.child.reduceNode(pos, dtmc)
	assert.Nil(t, err)
	assert.Equal(t, expectedNode, newNode)
	assert.False(t, newChildHash)
}

func TestBranchNode_printShouldNotPanicEvenIfNodeIsCollapsed(t *testing.T) {
	t.Parallel()

	bnWriter := bytes.NewBuffer(make([]byte, 0))
	collapsedBnWriter := bytes.NewBuffer(make([]byte, 0))

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	_ = bn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)
	_ = collapsedBn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)

	bn.print(bnWriter, 0, dtmc, db)
	collapsedBn.print(collapsedBnWriter, 0, dtmc, db)

	assert.Equal(t, bnWriter.Bytes(), collapsedBnWriter.Bytes())
}

func TestBranchNode_getDirtyHashesFromCleanNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	_ = bn.commitDirty(db, db, dtmc)
	dirtyHashes := make(common.ModifiedHashes)

	err := bn.getDirtyHashes(dirtyHashes)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(dirtyHashes))
}

func TestBranchNode_getAllHashes(t *testing.T) {
	t.Parallel()

	trieNodes := 4
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	hashes, err := bn.getAllHashes(dtmc, testscommon.NewMemDbMock())
	assert.Nil(t, err)
	assert.Equal(t, trieNodes, len(hashes))
}

func TestBranchNode_getAllHashesResolvesCollapsed(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	_ = bn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)

	hashes, err := collapsedBn.getAllHashes(dtmc, db)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(hashes))
}

func TestBranchNode_getNextHashAndKey(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	proofVerified, nextHash, nextKey := collapsedBn.getNextHashAndKey([]byte{2})

	assert.False(t, proofVerified)
	assert.Equal(t, collapsedBn.EncodedChildren[2], nextHash)
	assert.Equal(t, []byte{}, nextKey)
}

func TestBranchNode_getNextHashAndKeyNilKey(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	proofVerified, nextHash, nextKey := collapsedBn.getNextHashAndKey(nil)

	assert.False(t, proofVerified)
	assert.Nil(t, nextHash)
	assert.Nil(t, nextKey)
}

func TestBranchNode_getNextHashAndKeyNilNode(t *testing.T) {
	t.Parallel()

	var collapsedBn *branchNode
	proofVerified, nextHash, nextKey := collapsedBn.getNextHashAndKey([]byte{2})

	assert.False(t, proofVerified)
	assert.Nil(t, nextHash)
	assert.Nil(t, nextKey)
}

func TestBranchNode_SizeInBytes(t *testing.T) {
	t.Parallel()

	var bn *branchNode
	assert.Equal(t, 0, bn.sizeInBytes())

	collapsed1 := []byte("collapsed1")
	collapsed2 := []byte("collapsed2")
	hash := bytes.Repeat([]byte{1}, 32)
	bn = &branchNode{
		CollapsedBn: CollapsedBn{
			EncodedChildren: [][]byte{collapsed1, collapsed2},
			ChildrenVersion: []byte("version"),
		},
		children: [17]node{},
		baseNode: &baseNode{
			hash:   hash,
			dirty:  false,
			marsh:  nil,
			hasher: nil,
		},
	}
	numChildren := 2
	assert.Equal(t, numChildren*hashSizeInBytes+len(hash)+1+19*pointerSizeInBytes+len(bn.ChildrenVersion), bn.sizeInBytes())
}

func TestBranchNode_commitContextDone(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bn.commitSnapshot(db, nil, nil, ctx, statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)
	assert.Equal(t, core.ErrContextClosing, err)
}

func TestBranchNode_commitSnapshotDbIsClosing(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	db.GetCalled = func(key []byte) ([]byte, error) {
		return nil, core.ErrContextClosing
	}

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	missingNodesChan := make(chan []byte, 10)
	err := collapsedBn.commitSnapshot(db, nil, missingNodesChan, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)
	assert.True(t, core.IsClosingError(err))
	assert.Equal(t, 0, len(missingNodesChan))
}

func TestBranchNode_commitSnapshotChildIsMissingErr(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	db.GetCalled = func(key []byte) ([]byte, error) {
		return nil, core.NewGetNodeFromDBErrWithKey(key, ErrKeyNotFound, "test")
	}

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	missingNodesChan := make(chan []byte, 10)
	err := collapsedBn.commitSnapshot(db, nil, missingNodesChan, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, dtmc)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(missingNodesChan))
}

func TestBranchNode_getVersion(t *testing.T) {
	t.Parallel()

	t.Run("nil ChildrenVersion", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

		version, err := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Nil(t, err)
	})

	t.Run("NotSpecified for all children", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.NotSpecified)
		bn.ChildrenVersion[6] = byte(core.NotSpecified)
		bn.ChildrenVersion[13] = byte(core.NotSpecified)

		version, err := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Nil(t, err)
	})

	t.Run("one child with autoBalanceEnabled", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.NotSpecified)
		bn.ChildrenVersion[6] = byte(core.AutoBalanceEnabled)
		bn.ChildrenVersion[13] = byte(core.NotSpecified)

		version, err := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Nil(t, err)
	})

	t.Run("AutoBalanceEnabled for all children", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.AutoBalanceEnabled)
		bn.ChildrenVersion[6] = byte(core.AutoBalanceEnabled)
		bn.ChildrenVersion[13] = byte(core.AutoBalanceEnabled)

		version, err := bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
		assert.Nil(t, err)
	})
}

func TestBranchNode_getValueReturnsEmptyByteSlice(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	assert.Equal(t, []byte{}, bn.getValue())
}

func TestBranchNode_VerifyChildrenVersionIsSetCorrectlyAfterInsertAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("revert child from version 1 to 0", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.AutoBalanceEnabled)

		childKey := []byte{2, 'd', 'o', 'g'}
		data := core.TrieData{
			Key:     childKey,
			Value:   []byte("value"),
			Version: 0,
		}
		newBn, _, err := bn.insert(data, dtmc, &testscommon.MemDbMock{})
		assert.Nil(t, err)
		assert.Nil(t, newBn.(*branchNode).ChildrenVersion)
	})

	t.Run("remove migrated child", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.AutoBalanceEnabled)
		childKey := []byte{2, 'd', 'o', 'g'}

		_, newBn, _, err := bn.delete(childKey, dtmc, &testscommon.MemDbMock{})
		assert.Nil(t, err)
		assert.Nil(t, newBn.(*branchNode).ChildrenVersion)
	})
}

func TestBranchNode_revertChildrenVersionSliceIfNeeded(t *testing.T) {
	t.Parallel()

	t.Run("nil ChildrenVersion does not panic", func(t *testing.T) {
		t.Parallel()

		bn := &branchNode{}
		bn.revertChildrenVersionSliceIfNeeded()
	})

	t.Run("revert is not needed", func(t *testing.T) {
		t.Parallel()

		childrenVersion := make([]byte, nrOfChildren)
		childrenVersion[5] = byte(core.AutoBalanceEnabled)
		bn := &branchNode{
			CollapsedBn: CollapsedBn{
				ChildrenVersion: childrenVersion,
			},
		}

		bn.revertChildrenVersionSliceIfNeeded()
		assert.Equal(t, nrOfChildren, len(bn.ChildrenVersion))
		assert.Equal(t, byte(core.AutoBalanceEnabled), bn.ChildrenVersion[5])
	})

	t.Run("revert is needed", func(t *testing.T) {
		t.Parallel()

		bn := &branchNode{
			CollapsedBn: CollapsedBn{
				ChildrenVersion: make([]byte, nrOfChildren),
			},
		}

		bn.revertChildrenVersionSliceIfNeeded()
		assert.Nil(t, bn.ChildrenVersion)
	})
}

func TestBranchNode_getNodeData(t *testing.T) {
	t.Parallel()

	t.Run("nil node", func(t *testing.T) {
		t.Parallel()

		var bn *branchNode
		nodeData, err := bn.getNodeData(keyBuilder.NewDisabledKeyBuilder())
		assert.Nil(t, nodeData)
		assert.True(t, errors.Is(err, ErrNilBranchNode))
	})
	t.Run("gets data from all non-nil children", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()
		_ = tr.Update([]byte("111"), []byte("111"))
		_ = tr.Update([]byte("aaa"), []byte("aaa"))
		_ = tr.Commit()

		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)

		hashSize := 32
		keySize := 1
		kb := keyBuilder.NewKeyBuilder()
		nodeData, err := bn.getNodeData(kb)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(nodeData))

		// branch node as child
		firstChildData := nodeData[0]
		assert.Equal(t, uint(1), firstChildData.GetKeyBuilder().Size())
		assert.Equal(t, bn.EncodedChildren[1], firstChildData.GetData())
		assert.Equal(t, uint64(hashSize+keySize), firstChildData.Size())
		assert.False(t, firstChildData.IsLeaf())

		// leaf node as child
		seconChildData := nodeData[1]
		assert.Equal(t, uint(1), seconChildData.GetKeyBuilder().Size())
		assert.Equal(t, bn.EncodedChildren[5], seconChildData.GetData())
		assert.Equal(t, uint64(hashSize+keySize), seconChildData.Size())
		assert.False(t, seconChildData.IsLeaf())

		// extension node as child
		thirdChildData := nodeData[2]
		assert.Equal(t, uint(1), thirdChildData.GetKeyBuilder().Size())
		assert.Equal(t, bn.EncodedChildren[7], thirdChildData.GetData())
		assert.Equal(t, uint64(hashSize+keySize), thirdChildData.Size())
		assert.False(t, thirdChildData.IsLeaf())
	})
}

func TestBranchNode_shouldCollapseChild(t *testing.T) {
	t.Parallel()

	t.Run("empty hexKey", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		shouldCollapseChild := bn.shouldCollapseChild([]byte{}, trieMetricsCollector.NewTrieMetricsCollector())
		assert.False(t, shouldCollapseChild)
	})
	t.Run("invalid hexKey", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		shouldCollapseChild := bn.shouldCollapseChild([]byte{17}, trieMetricsCollector.NewTrieMetricsCollector())
		assert.False(t, shouldCollapseChild)
	})
	t.Run("child is nil", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		shouldCollapseChild := bn.shouldCollapseChild([]byte{4}, trieMetricsCollector.NewTrieMetricsCollector())
		assert.False(t, shouldCollapseChild)
	})
	t.Run("collapse child that is already collapsed", func(t *testing.T) {
		t.Parallel()

		_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		leafChildKey := append([]byte{2}, []byte("dog")...)
		shouldCollapseChild := collapsedBn.shouldCollapseChild(leafChildKey, trieMetricsCollector.NewTrieMetricsCollector())
		assert.False(t, shouldCollapseChild)
	})
	t.Run("successful collapse", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		leafSize := bn.children[2].sizeInBytes()
		leafChildKey := append([]byte{2}, []byte("dog")...)
		tmc := trieMetricsCollector.NewTrieMetricsCollector()
		shouldCollapseChild := bn.shouldCollapseChild(leafChildKey, tmc)
		assert.False(t, shouldCollapseChild)
		assert.Nil(t, bn.children[2])
		assert.Equal(t, -leafSize-pointerSizeInBytes, tmc.GetSizeLoadedInMem())
	})
}
