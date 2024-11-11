package trie

import (
	"bytes"
	"context"
	"errors"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/trieBatchManager"
	"github.com/stretchr/testify/assert"
)

func getTestGoroutinesManager() common.TrieGoroutinesManager {
	th, _ := throttler.NewNumGoRoutinesThrottler(1)
	manager, _ := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	return manager
}

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
	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	tr := &patriciaMerkleTrie{
		trieStorage:             trieStorage,
		marshalizer:             args.Marshalizer,
		hasher:                  args.Hasher,
		maxTrieLevelInMemory:    5,
		chanClose:               make(chan struct{}),
		enableEpochsHandler:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		batchManager:            trieBatchManager.NewTrieBatchManager(),
		goroutinesThrottler:     th,
		RootManager:             NewRootManager(),
		trieOperationInProgress: &atomic.Flag{},
	}

	return tr, trieStorage
}

func initTrie() *patriciaMerkleTrie {
	tr, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.Commit()

	return tr
}

func getEncodedTrieNodesAndHashes(tr common.Trie) ([][]byte, [][]byte) {
	rootHash, _ := tr.RootHash()
	it, _ := NewDFSIterator(tr, rootHash)
	encNode, _ := it.MarshalizedNode()

	nodes := make([][]byte, 0)
	nodes = append(nodes, encNode)

	hashes := make([][]byte, 0)
	hash := it.GetHash()
	hashes = append(hashes, hash)

	for it.HasNext() {
		_ = it.Next()
		encNode, _ = it.MarshalizedNode()

		nodes = append(nodes, encNode)
		hash = it.GetHash()
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
	bn.setHash(getTestGoroutinesManager())
	collapsedBn.dirty = true
	collapsedBn.hash = bn.hash

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
	manager := getTestGoroutinesManager()

	bn.setHash(manager)
	assert.Nil(t, manager.GetError())
	assert.Equal(t, hash, bn.hash)
}

func TestBranchNode_setHashCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	hash, _ := encodeNodeAndGetHash(collapsedBn)
	manager := getTestGoroutinesManager()

	collapsedBn.setHash(manager)
	assert.Nil(t, manager.GetError())
	assert.Equal(t, hash, collapsedBn.hash)
}

func TestBranchNode_setGivenHash(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{}}
	expectedHash := []byte("node hash")

	bn.setGivenHash(expectedHash)
	assert.Equal(t, expectedHash, bn.hash)
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
	bn.setHash(getTestGoroutinesManager())

	err := bn.commitDirty(0, 5, db, db)
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

	err := bn.commitDirty(0, 5, nil, nil)
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
}

func TestBranchNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	err := bn.commitDirty(0, 5, nil, nil)
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

func TestBranchNode_resolveIfCollapsed(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)

	bn.setHash(getTestGoroutinesManager())
	_ = bn.commitDirty(0, 5, db, db)
	resolved, _ := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), bn.marsh, bn.hasher)
	resolved.dirty = false
	resolved.hash = bn.EncodedChildren[childPos]

	childNode, err := collapsedBn.resolveIfCollapsed(childPos, db)
	assert.Nil(t, err)
	assert.Equal(t, resolved, collapsedBn.children[childPos])
	assert.Equal(t, resolved, childNode)
}

func TestBranchNode_resolveCollapsedPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	childNode, err := bn.resolveIfCollapsed(17, nil)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, childNode)
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

	val, maxDepth, err := bn.tryGet(key, 0, nil)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), maxDepth)
}

func TestBranchNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	var key []byte

	val, maxDepth, err := bn.tryGet(key, 0, nil)
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestBranchNode_tryGetChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	key := []byte("dog")

	val, maxDepth, err := bn.tryGet(key, 0, nil)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestBranchNode_tryGetNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nilChildKey := []byte{3}

	val, maxDepth, err := bn.tryGet(nilChildKey, 0, nil)
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestBranchNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	bn.setHash(getTestGoroutinesManager())
	_ = bn.commitDirty(0, 5, db, db)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	val, maxDepth, err := collapsedBn.tryGet(key, 0, db)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), maxDepth)
}

func TestBranchNode_getNext(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nextNode, _ := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), bn.marsh, bn.hasher)
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	n, key, err := bn.getNext(key, nil)

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

	n, key, err := bn.getNext(key, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_getNextNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nilChildPos := byte(4)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	n, key, err := bn.getNext(key, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestBranchNode_insert(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nodeKey := []byte{0, 2, 3}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(nodeKey), "dogs")}

	newBn, _ := bn.insert(data, goRoutinesManager, nil)
	assert.NotNil(t, newBn)
	assert.Nil(t, goRoutinesManager.GetError())

	nodeKeyRemainder := nodeKey[1:]
	bn.children[0], _ = newLeafNode(getTrieDataWithDefaultVersion(string(nodeKeyRemainder), "dogs"), bn.marsh, bn.hasher)
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_insertEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion("", "dogs")}

	newBn, _ := bn.insert(data, goRoutinesManager, nil)
	assert.Equal(t, ErrValueTooShort, goRoutinesManager.GetError())
	assert.Nil(t, newBn)
}

func TestBranchNode_insertChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion("dog", "dogs")}

	newBn, _ := bn.insert(data, goRoutinesManager, nil)
	assert.Equal(t, ErrChildPosOutOfRange, goRoutinesManager.GetError())
	assert.Nil(t, newBn)
}

func TestBranchNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	bn.setHash(getTestGoroutinesManager())
	_ = bn.commitDirty(0, 5, db, db)

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newBn, _ := collapsedBn.insert(data, goRoutinesManager, db)
	assert.NotNil(t, newBn)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, _ := newBn.tryGet(key, 0, db)
	assert.Equal(t, []byte("dogs"), val)
}

func TestBranchNode_insertInStoredBnOnExistingPos(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	bn.setHash(getTestGoroutinesManager())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	_ = bn.commitDirty(0, 5, db, db)
	bnHash := bn.getHash()
	ln, _, _ := bn.getNext(key, db)
	lnHash := ln.getHash()
	expectedHashes := [][]byte{lnHash, bnHash}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newNode, oldHashes := bn.insert(data, goRoutinesManager, db)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestBranchNode_insertInStoredBnOnNilPos(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	bn.setHash(getTestGoroutinesManager())
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	_ = bn.commitDirty(0, 5, db, db)
	bnHash := bn.getHash()
	expectedHashes := [][]byte{bnHash}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newNode, oldHashes := bn.insert(data, goRoutinesManager, db)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestBranchNode_insertInDirtyBnOnNilPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newNode, oldHashes := bn.insert(data, goRoutinesManager, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestBranchNode_insertInDirtyBnOnExistingPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newNode, oldHashes := bn.insert(data, goRoutinesManager, nil)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, oldHashes)
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
	data := []core.TrieData{{Key: key}}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	dirty, newBn, _ := bn.delete(data, goRoutinesManager, nil)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())

	expectedBn.setHash(getTestGoroutinesManager())
	newBn.setHash(getTestGoroutinesManager())
	assert.Equal(t, expectedBn.getHash(), newBn.getHash())
}

func TestBranchNode_deleteFromStoredBn(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	bn.setHash(getTestGoroutinesManager())
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)

	_ = bn.commitDirty(0, 5, db, db)
	bnHash := bn.getHash()
	ln, _, _ := bn.getNext(lnKey, db)
	lnHash := ln.getHash()
	expectedHashes := [][]byte{lnHash, bnHash}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	data := []core.TrieData{{Key: lnKey}}
	dirty, _, oldHashes := bn.delete(data, goRoutinesManager, db)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestBranchNode_deleteFromDirtyBn(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)
	data := []core.TrieData{{Key: lnKey}}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	dirty, _, oldHashes := bn.delete(data, goRoutinesManager, nil)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestBranchNode_deleteNonexistentNodeFromChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("butterfly")...)
	data := []core.TrieData{{Key: key}}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	dirty, newBn, _ := bn.delete(data, goRoutinesManager, nil)
	assert.False(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_deleteEmptykey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	data := []core.TrieData{{Key: []byte{}}}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	dirty, newBn, _ := bn.delete(data, goRoutinesManager, nil)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, goRoutinesManager.GetError())
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	bn.setHash(getTestGoroutinesManager())
	_ = bn.commitDirty(0, 5, db, db)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	data := []core.TrieData{{Key: key}}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	dirty, newBn, _ := collapsedBn.delete(data, goRoutinesManager, db)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, err := newBn.tryGet(key, 0, db)
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

	key := append([]byte{firstChildPos}, []byte("dog")...)
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion(string(key), "dog"), bn.marsh, bn.hasher)

	key = append([]byte{secondChildPos}, []byte("doe")...)
	data := []core.TrieData{{Key: key}}

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	dirty, newBn, _ := bn.delete(data, goRoutinesManager, nil)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, ln, newBn)
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

	n, newChildHash, err := bn.children[childPos].reduceNode(int(childPos), nil)
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
	tr.Delete([]byte("wolf"))

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
	tr.Delete([]byte("doe"))

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
	tr.Delete([]byte("dog"))

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
	tr.Delete([]byte("dogglesworth"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()

	assert.Equal(t, expectedHash, hash)
}

func TestBranchNode_getChildren(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())

	children, err := bn.getChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(children))
}

func TestBranchNode_getChildrenCollapsedBn(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	bn.setHash(getTestGoroutinesManager())
	_ = bn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, 0)

	children, err := collapsedBn.getChildren(db)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(children))
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
	rootNode := tr.GetRootNode()
	rootNode.setHash(getTestGoroutinesManager())
	nodes, _ := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := cache.NewLRUCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], hasher)
		nodesCacher.Put(n.hash, n, len(n.GetSerialized()))
	}

	firstChildIndex := 5
	secondChildIndex := 7

	bn := getCollapsedBn(t, rootNode)

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

	collapsedRoot, _ := tr.GetRootNode().getCollapsed()
	tr.Delete([]byte("zzz"))
	ExecuteUpdatesFromBatch(tr)

	assert.True(t, collapsedRoot.isDirty())

	_ = tr.Commit()

	assert.False(t, collapsedRoot.isDirty())
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
		CollapsedBn: CollapsedBn{
			EncodedChildren: make([][]byte, nrOfChildren),
		},
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

	manager := getTestGoroutinesManager()
	bn.setHash(manager)
	assert.Nil(t, manager.GetError())
}

func TestBranchNode_commitCollapsesTrieIfMaxTrieLevelInMemoryIsReached(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	bn.setHash(getTestGoroutinesManager())
	collapsedBn.setHash(getTestGoroutinesManager())

	err := bn.commitDirty(0, 1, testscommon.NewMemDbMock(), testscommon.NewMemDbMock())
	assert.Nil(t, err)

	assert.Equal(t, collapsedBn.EncodedChildren, bn.EncodedChildren)
	assert.Equal(t, collapsedBn.children, bn.children)
	assert.Equal(t, collapsedBn.hash, bn.hash)
}

func TestBranchNode_reduceNodeBnChild(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	en, _ := getEnAndCollapsedEn()
	pos := 5
	expectedNode, _ := newExtensionNode([]byte{byte(pos)}, en.child, marsh, hasher)

	newNode, newChildHash, err := en.child.reduceNode(pos, nil)
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
	bn.setHash(getTestGoroutinesManager())
	collapsedBn.setHash(getTestGoroutinesManager())
	_ = bn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, 0)
	_ = collapsedBn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, 0)

	bn.print(bnWriter, 0, db)
	collapsedBn.print(collapsedBnWriter, 0, db)

	assert.Equal(t, bnWriter.Bytes(), collapsedBnWriter.Bytes())
}

func TestBranchNode_getDirtyHashesFromCleanNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	bn.setHash(getTestGoroutinesManager())
	_ = bn.commitDirty(0, 5, db, db)
	dirtyHashes := make(common.ModifiedHashes)

	err := bn.getDirtyHashes(dirtyHashes)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(dirtyHashes))
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
	hash := []byte("hash")
	bn = &branchNode{
		CollapsedBn: CollapsedBn{
			EncodedChildren: [][]byte{collapsed1, collapsed2},
		},
		children: [17]node{},
		baseNode: &baseNode{
			hash:   hash,
			dirty:  false,
			marsh:  nil,
			hasher: nil,
		},
	}
	assert.Equal(t, len(collapsed1)+len(collapsed2)+len(hash)+1+19*pointerSizeInBytes, bn.sizeInBytes())
}

func TestBranchNode_commitContextDone(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bn.commitSnapshot(db, nil, nil, ctx, statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, 0)
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
	err := collapsedBn.commitSnapshot(db, nil, missingNodesChan, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, 0)
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
	collapsedBn.setHash(getTestGoroutinesManager())
	missingNodesChan := make(chan []byte, 10)
	err := collapsedBn.commitSnapshot(db, nil, missingNodesChan, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, 0)
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

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		childKey := []byte{2, 'd', 'o', 'g'}
		data := core.TrieData{
			Key:     childKey,
			Value:   []byte("value"),
			Version: 0,
		}

		newBn, _ := bn.insert([]core.TrieData{data}, goRoutinesManager, &testscommon.MemDbMock{})
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Nil(t, newBn.(*branchNode).ChildrenVersion)
	})

	t.Run("remove migrated child", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.AutoBalanceEnabled)
		childKey := []byte{2, 'd', 'o', 'g'}
		data := []core.TrieData{{Key: childKey}}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		_, newBn, _ := bn.delete(data, goRoutinesManager, &testscommon.MemDbMock{})
		assert.Nil(t, goRoutinesManager.GetError())
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

func TestBranchNode_splitDataForChildren(t *testing.T) {
	t.Parallel()

	t.Run("empty array returns err", func(t *testing.T) {
		t.Parallel()

		var newData []core.TrieData
		data, err := splitDataForChildren(newData)
		assert.Nil(t, data)
		assert.Equal(t, ErrValueTooShort, err)
	})
	t.Run("empty key returns err", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{Key: []byte{2, 3, 4}},
			{Key: []byte{}},
		}

		data, err := splitDataForChildren(newData)
		assert.Nil(t, data)
		assert.Equal(t, ErrValueTooShort, err)
	})
	t.Run("child pos out of range returns err", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{Key: []byte{2, 3, 4}},
			{Key: []byte{17, 2, 3}},
		}

		data, err := splitDataForChildren(newData)
		assert.Nil(t, data)
		assert.Equal(t, ErrChildPosOutOfRange, err)
	})
	t.Run("one child on last pos should work", func(t *testing.T) {
		t.Parallel()

		childPos := byte(16)
		newData := []core.TrieData{
			{Key: []byte{childPos}},
		}

		data, err := splitDataForChildren(newData)
		assert.True(t, len(data) == nrOfChildren)
		assert.Nil(t, err)
		assert.Equal(t, newData, data[childPos])
	})
	t.Run("all children have same pos should work", func(t *testing.T) {
		t.Parallel()

		childPos := byte(2)
		newData := []core.TrieData{
			{Key: []byte{childPos, 3, 4}},
			{Key: []byte{childPos, 5, 6}},
			{Key: []byte{childPos, 7, 8}},
		}

		data, err := splitDataForChildren(newData)
		assert.True(t, len(data) == nrOfChildren)
		assert.Nil(t, err)
		for i := range data {
			if i != int(childPos) {
				assert.Nil(t, data[i])
				continue
			}
			assert.Equal(t, newData, data[i])
		}
	})
	t.Run("all children have different pos should work", func(t *testing.T) {
		t.Parallel()

		childPos1 := byte(2)
		childPos2 := byte(6)
		childPos3 := byte(13)
		newData := []core.TrieData{
			{Key: []byte{childPos1, 3, 4}},
			{Key: []byte{childPos2, 5, 6}},
			{Key: []byte{childPos3, 7, 8}},
		}

		data, err := splitDataForChildren(newData)
		assert.True(t, len(data) == nrOfChildren)
		assert.Nil(t, err)
		for i := range data {
			if i == int(childPos1) {
				assert.Equal(t, []core.TrieData{newData[0]}, data[i])
			} else if i == int(childPos2) {
				assert.Equal(t, []core.TrieData{newData[1]}, data[i])
			} else if i == int(childPos3) {
				assert.Equal(t, []core.TrieData{newData[2]}, data[i])
			} else {
				assert.Nil(t, data[i])
			}
		}
	})
	t.Run("some children have same pos should work", func(t *testing.T) {
		t.Parallel()

		childPos1 := byte(2)
		childPos2 := byte(6)
		newData := []core.TrieData{
			{Key: []byte{childPos1, 3, 4}},
			{Key: []byte{childPos1, 5, 6}},
			{Key: []byte{childPos2, 7, 8}},
		}

		data, err := splitDataForChildren(newData)
		assert.True(t, len(data) == nrOfChildren)
		assert.Nil(t, err)
		for i := range data {
			if i == int(childPos1) {
				assert.Equal(t, []core.TrieData{newData[0], newData[1]}, data[i])
			} else if i == int(childPos2) {
				assert.Equal(t, []core.TrieData{newData[2]}, data[i])
			} else {
				assert.Nil(t, data[i])
			}
		}
	})
	t.Run("child pos is removed from key", func(t *testing.T) {
		t.Parallel()

		childPos := byte(2)
		newData := []core.TrieData{
			{Key: []byte{childPos, 3, 4}},
			{Key: []byte{childPos, 5, 6}},
		}

		data, err := splitDataForChildren(newData)
		assert.True(t, len(data) == nrOfChildren)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(data[childPos]))
		assert.Equal(t, []byte{3, 4}, newData[0].Key)
		assert.Equal(t, []byte{5, 6}, newData[1].Key)
	})
}

func TestBranchNode_insertOnNilChild(t *testing.T) {
	t.Parallel()

	t.Run("empty data should err", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		var data []core.TrieData

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		newNode, modifiedHashes := bn.insertOnNilChild(data, goRoutinesManager, nil)
		assert.Equal(t, 0, len(modifiedHashes))
		assert.Equal(t, ErrValueTooShort, goRoutinesManager.GetError())
		assert.Nil(t, newNode)
	})
	t.Run("insert one child in !dirty node", func(t *testing.T) {
		t.Parallel()

		db := testscommon.NewMemDbMock()
		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		bn.setHash(getTestGoroutinesManager())
		err := bn.commitDirty(0, 5, db, db)
		assert.Nil(t, err)
		assert.False(t, bn.dirty)
		originalHash := bn.getHash()
		assert.True(t, len(originalHash) > 0)
		newData := []core.TrieData{
			{
				Key:     []byte{1, 2, 3},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		newNode, modifiedHashes := bn.insertOnNilChild(newData, goRoutinesManager, db)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Equal(t, 0, len(modifiedHashes))
		assert.NotNil(t, newNode)
		assert.True(t, newNode.isDirty())
	})
	t.Run("insert one child in dirty node", func(t *testing.T) {
		t.Parallel()

		db := testscommon.NewMemDbMock()
		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		assert.True(t, bn.dirty)
		newData := []core.TrieData{
			{
				Key:     []byte{1, 2, 3},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		newNode, modifiedHashes := bn.insertOnNilChild(newData, goRoutinesManager, db)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Equal(t, 0, len(modifiedHashes))
		assert.NotNil(t, newNode)
		assert.True(t, newNode.isDirty())
	})
	t.Run("insert multiple children", func(t *testing.T) {
		t.Parallel()

		db := testscommon.NewMemDbMock()
		bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
		newData := []core.TrieData{
			{
				Key:     []byte{1, 2, 3},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
			{
				Key:     []byte{1, 2, 4},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		newNode, modifiedHashes := bn.insertOnNilChild(newData, goRoutinesManager, db)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Equal(t, 0, len(modifiedHashes))
		_, ok := newNode.(*extensionNode)
		assert.True(t, ok)
		assert.NotNil(t, newNode)
	})
}

func TestBranchNode_insertOnExistingChild(t *testing.T) {
	t.Parallel()

	t.Run("insert on existing child multiple children", func(t *testing.T) {
		t.Parallel()

		childPos := byte(2)
		db := testscommon.NewMemDbMock()
		var children [nrOfChildren]node
		marshaller, hasher := getTestMarshalizerAndHasher()
		children[2], _ = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 5}), "dog"), marshaller, hasher)
		children[6], _ = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"), marshaller, hasher)
		bn, _ := newBranchNode(marshaller, hasher)
		bn.children = children
		bn.setHash(getTestGoroutinesManager())
		newData := []core.TrieData{
			{
				Key:     []byte{1, 2, 3},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
			{
				Key:     []byte{1, 2, 4},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
		}
		err := bn.commitDirty(0, 5, db, db)
		assert.Nil(t, err)
		assert.False(t, bn.dirty)
		originalHash := bn.getHash()
		assert.True(t, len(originalHash) > 0)
		originalChildHash := bn.children[childPos].getHash()
		assert.True(t, len(originalChildHash) > 0)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, _ := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		newNode, modifiedHashes := bn.insertOnChild(newData, int(childPos), goRoutinesManager, db)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.NotNil(t, newNode)
		assert.Equal(t, 1, len(modifiedHashes))
		assert.Equal(t, originalChildHash, modifiedHashes[0])
		_, ok := newNode.(*extensionNode)
		assert.True(t, ok)
	})
	t.Run("insert on existing child same node", func(t *testing.T) {
		t.Parallel()

		childPos := byte(2)
		db := testscommon.NewMemDbMock()
		var children [nrOfChildren]node
		marshaller, hasher := getTestMarshalizerAndHasher()
		key := []byte{1, 2, 5}
		value := "dog"
		children[2], _ = newLeafNode(getTrieDataWithDefaultVersion(string(key), value), marshaller, hasher)
		children[6], _ = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"), marshaller, hasher)
		bn, _ := newBranchNode(marshaller, hasher)
		bn.children = children
		bn.setHash(getTestGoroutinesManager())
		newData := []core.TrieData{
			{
				Key:     key,
				Value:   []byte(value),
				Version: core.NotSpecified,
			},
		}
		err := bn.commitDirty(0, 5, db, db)
		assert.Nil(t, err)
		assert.False(t, bn.dirty)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		newNode, modifiedHashes := bn.insertOnChild(newData, int(childPos), goRoutinesManager, db)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Nil(t, newNode)
		assert.False(t, bn.dirty)
		assert.Equal(t, 0, len(modifiedHashes))
		_, ok := bn.children[childPos].(*leafNode)
		assert.True(t, ok)
	})
}

func TestBranchNode_insertBatch(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	var children [nrOfChildren]node
	marshaller, hasher := getTestMarshalizerAndHasher()
	children[2], _ = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{3, 4, 5}), "dog"), marshaller, hasher)
	children[6], _ = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{7, 8, 9}), "doe"), marshaller, hasher)
	bn, _ := newBranchNode(marshaller, hasher)
	bn.children = children
	bn.setHash(getTestGoroutinesManager())

	newData := []core.TrieData{
		{
			Key:   []byte{1, 2, 3},
			Value: []byte("value1"),
		},
		{
			Key:   []byte{6, 7, 8, 16},
			Value: []byte("value2"),
		},
		{
			Key:   []byte{6, 10, 11, 16},
			Value: []byte("value3"),
		},
		{
			Key:   []byte{16},
			Value: []byte("value4"),
		},
	}
	err := bn.commitDirty(0, 5, db, db)
	assert.Nil(t, err)
	assert.False(t, bn.dirty)

	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
	assert.Nil(t, err)

	newNode, modifiedHashes := bn.insert(newData, goRoutinesManager, db)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, 2, len(modifiedHashes))
	assert.True(t, newNode.isDirty())

	bn, ok := newNode.(*branchNode)
	assert.True(t, ok)
	assert.True(t, bn.dirty)
	assert.False(t, bn.children[2].isDirty())
	_, ok = bn.children[1].(*leafNode)
	assert.True(t, ok)
	_, ok = bn.children[6].(*branchNode)
	assert.True(t, ok)
	_, ok = bn.children[16].(*leafNode)
	assert.True(t, ok)

}

func getNewBn() *branchNode {
	marsh, hasher := getTestMarshalizerAndHasher()
	var children [nrOfChildren]node
	childBn, _ := newBranchNode(marsh, hasher)
	childBn.children[1], _ = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{3, 4, 5}), "dog"), marsh, hasher)
	childBn.children[3], _ = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{7, 8, 9}), "doe"), marsh, hasher)

	children[4], _ = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{3, 4, 5}), "dog"), marsh, hasher)
	children[7], _ = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{7, 8, 9}), "doe"), marsh, hasher)
	children[9] = childBn

	bn, _ := newBranchNode(marsh, hasher)
	bn.children = children
	bn.setHash(getTestGoroutinesManager())
	_ = bn.commitDirty(0, 5, testscommon.NewMemDbMock(), testscommon.NewMemDbMock())
	return bn
}

func TestBranchNode_deleteBatch(t *testing.T) {
	t.Parallel()

	t.Run("delete multiple children", func(t *testing.T) {
		t.Parallel()

		bn := getNewBn()
		assert.False(t, bn.dirty)

		data := []core.TrieData{
			{
				Key: []byte{4, 3, 4, 5},
			},
			{
				Key: []byte{9, 1, 3, 4, 5},
			},
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		dirty, newNode, modifiedHashes := bn.delete(data, goRoutinesManager, testscommon.NewMemDbMock())
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, dirty)
		assert.True(t, newNode.isDirty())
		assert.Equal(t, 5, len(modifiedHashes))
		bn, ok := newNode.(*branchNode)
		assert.True(t, ok)

		assert.Nil(t, bn.children[4])
		assert.False(t, bn.children[7].isDirty())
		_, ok = bn.children[9].(*leafNode)
		assert.True(t, ok)

	})
	t.Run("reduce node after delete batch", func(t *testing.T) {
		t.Parallel()

		bn := getNewBn()
		assert.False(t, bn.dirty)

		data := []core.TrieData{
			{
				Key: []byte{4, 3, 4, 5},
			},
			{
				Key: []byte{7, 7, 8, 9},
			},
			{
				Key: []byte{9, 1, 3, 4, 5},
			},
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		dirty, newNode, modifiedHashes := bn.delete(data, goRoutinesManager, testscommon.NewMemDbMock())
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, dirty)
		assert.True(t, newNode.isDirty())
		assert.Equal(t, 6, len(modifiedHashes))
		ln, ok := newNode.(*leafNode)
		assert.True(t, ok)
		assert.Equal(t, []byte{9, 3, 7, 8, 9}, ln.Key)
	})
	t.Run("delete all children", func(t *testing.T) {
		t.Parallel()

		bn := getNewBn()
		assert.False(t, bn.dirty)

		data := []core.TrieData{
			{
				Key: []byte{4, 3, 4, 5},
			},
			{
				Key: []byte{7, 7, 8, 9},
			},
			{
				Key: []byte{9, 1, 3, 4, 5},
			},
			{
				Key: []byte{9, 3, 7, 8, 9},
			},
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}))
		assert.Nil(t, err)

		dirty, newNode, modifiedHashes := bn.delete(data, goRoutinesManager, testscommon.NewMemDbMock())
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, dirty)
		assert.Nil(t, newNode)
		assert.Equal(t, 6, len(modifiedHashes))
	})
}
