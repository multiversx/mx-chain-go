package trie

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state/hashesCollector"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/trieBatchManager"
	"github.com/stretchr/testify/assert"
)

const initialModifiedHashesCapacity = 10

func getTestGoroutinesManager() common.TrieGoroutinesManager {
	th, _ := throttler.NewNumGoRoutinesThrottler(5)
	goRoutinesManager, _ := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")

	return goRoutinesManager
}

func getTestMarshalizerAndHasher() (marshal.Marshalizer, hashing.Hasher) {
	marsh := &marshal.GogoProtoMarshalizer{}
	hash := &testscommon.KeccakMock{}
	return marsh, hash
}

func getDefaultTrieContext() common.TrieContext {
	args := GetDefaultTrieStorageManagerParameters()
	trieStorage, _ := NewTrieStorageManager(args)
	return &trieContext{
		StorageManager: trieStorage,
		Marshalizer:    &marshal.GogoProtoMarshalizer{},
		Hasher:         &testscommon.KeccakMock{},
	}
}

func getTrieContextWithCustomStorage(storage common.StorageManager) common.TrieContext {
	return &trieContext{
		StorageManager: storage,
		Marshalizer:    &marshal.GogoProtoMarshalizer{},
		Hasher:         &testscommon.KeccakMock{},
	}
}

func getTrieDataWithDefaultVersion(key string, val string) core.TrieData {
	return core.TrieData{
		Key:     []byte(key),
		Value:   []byte(val),
		Version: core.NotSpecified,
	}
}

func getBnAndCollapsedBn() (*branchNode, *branchNode) {
	var children [nrOfChildren]node
	ChildrenHashes := make([][]byte, nrOfChildren)

	children[2] = newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
	children[6] = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"))
	children[13] = newLeafNode(getTrieDataWithDefaultVersion("doge", "doge"))
	bn := newBranchNode()
	bn.children = children

	trieCtx := &trieContext{
		Marshalizer: &marshal.GogoProtoMarshalizer{},
		Hasher:      &testscommon.KeccakMock{},
	}
	ChildrenHashes[2], _ = encodeNodeAndGetHash(children[2], trieCtx)
	ChildrenHashes[6], _ = encodeNodeAndGetHash(children[6], trieCtx)
	ChildrenHashes[13], _ = encodeNodeAndGetHash(children[13], trieCtx)
	collapsedBn := newBranchNode()
	collapsedBn.ChildrenHashes = ChildrenHashes
	bn.ChildrenHashes = ChildrenHashes

	return bn, collapsedBn
}

func emptyDirtyBranchNode() *branchNode {
	var children [nrOfChildren]node
	encChildren := make([][]byte, nrOfChildren)
	childrenVersion := make([]byte, nrOfChildren)

	return &branchNode{
		CollapsedBn: CollapsedBn{
			ChildrenHashes:  encChildren,
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
	thr, _ := throttler.NewNumGoRoutinesThrottler(10)
	tr := &patriciaMerkleTrie{
		TrieContext: &trieContext{
			StorageManager: trieStorage,
			Marshalizer:    args.Marshalizer,
			Hasher:         args.Hasher,
		},
		maxTrieLevelInMemory:    5,
		chanClose:               make(chan struct{}),
		enableEpochsHandler:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		batchManager:            trieBatchManager.NewTrieBatchManager(""),
		goRoutinesManager:       getTestGoroutinesManager(),
		RootManager:             NewRootManager(),
		trieOperationInProgress: &atomic.Flag{},
		throttler:               thr,
	}

	return tr, trieStorage
}

func initTrie() *patriciaMerkleTrie {
	tr, _ := newEmptyTrie()
	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

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

func TestBranchNode_isDirty(t *testing.T) {
	t.Parallel()

	bn := &branchNode{baseNode: &baseNode{dirty: true}}
	assert.Equal(t, true, bn.isDirty())

	bn = &branchNode{baseNode: &baseNode{dirty: false}}
	assert.Equal(t, false, bn.isDirty())
}

func TestBranchNode_commit(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	bn, collapsedBn := getBnAndCollapsedBn()
	hash, _ := encodeNodeAndGetHash(collapsedBn, trieCtx)

	manager := getTestGoroutinesManager()
	bn.commitDirty(0, 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
	assert.Nil(t, manager.GetError())

	encNode, _ := trieCtx.Get(hash)
	n, _ := decodeNode(encNode, trieCtx)
	h1, _ := encodeNodeAndGetHash(collapsedBn, trieCtx)
	h2, _ := encodeNodeAndGetHash(n, trieCtx)
	assert.Equal(t, h1, h2)
}

func TestBranchNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getTrieContextWithCustomStorage(nil)
	bn, _ := getBnAndCollapsedBn()

	expectedEncodedNode, _ := trieCtx.Marshal(bn)
	expectedEncodedNode = append(expectedEncodedNode, branch)

	encNode, err := bn.getEncodedNode(trieCtx)
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestBranchNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()

	bn := emptyDirtyBranchNode()

	encNode, err := bn.getEncodedNode(getTrieContextWithCustomStorage(nil))
	assert.True(t, errors.Is(err, ErrEmptyBranchNode))
	assert.Nil(t, encNode)
}

func TestBranchNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var bn *branchNode

	encNode, err := bn.getEncodedNode(getTrieContextWithCustomStorage(nil))
	assert.True(t, errors.Is(err, ErrNilBranchNode))
	assert.Nil(t, encNode)
}

func TestBranchNode_resolveIfCollapsed(t *testing.T) {
	t.Parallel()

	t.Run("child pos out of range", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()

		childNode, childHash, err := bn.resolveIfCollapsed(17, getTrieContextWithCustomStorage(nil))
		assert.Equal(t, ErrChildPosOutOfRange, err)
		assert.Nil(t, childNode)
		assert.Nil(t, childHash)
	})
	t.Run("resolve collapsed node", func(t *testing.T) {
		t.Parallel()

		bn, collapsedBn := getBnAndCollapsedBn()
		childPos := byte(2)

		trieCtx := getDefaultTrieContext()
		bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
		resolved := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
		resolved.dirty = false
		resolvedHash, _ := encodeNodeAndGetHash(resolved, trieCtx)

		childNode, childHash, err := collapsedBn.resolveIfCollapsed(childPos, trieCtx)
		assert.Nil(t, err)
		assert.Equal(t, resolved, collapsedBn.children[childPos])
		assert.Equal(t, resolved, childNode)
		assert.Equal(t, childHash, resolvedHash)
	})
	t.Run("invalid node state", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()
		bn.ChildrenHashes = make([][]byte, nrOfChildren)

		childNode, childHash, err := bn.resolveIfCollapsed(2, getTrieContextWithCustomStorage(nil))
		assert.Equal(t, ErrInvalidNodeState, err)
		assert.Nil(t, childNode)
		assert.Nil(t, childHash)
	})
	t.Run("node is not collapsed", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()

		resolved := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
		resolvedHash, _ := encodeNodeAndGetHash(resolved, getTrieContextWithCustomStorage(nil))

		childNode, childHash, err := bn.resolveIfCollapsed(2, getTrieContextWithCustomStorage(nil))
		assert.Nil(t, err)
		assert.Equal(t, resolved, childNode)
		assert.Equal(t, childHash, resolvedHash)
	})
}

func TestBranchNode_tryGet(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	val, maxDepth, err := bn.tryGet(key, 0, getTrieContextWithCustomStorage(nil))
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), maxDepth)
}

func TestBranchNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	var key []byte

	val, maxDepth, err := bn.tryGet(key, 0, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestBranchNode_tryGetChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	key := []byte("dog")

	val, maxDepth, err := bn.tryGet(key, 0, getTrieContextWithCustomStorage(nil))
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestBranchNode_tryGetNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	nilChildKey := []byte{3}

	val, maxDepth, err := bn.tryGet(nilChildKey, 0, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestBranchNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn()
	trieCtx := getDefaultTrieContext()
	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	val, maxDepth, err := collapsedBn.tryGet(key, 0, trieCtx)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), maxDepth)
}

func TestBranchNode_getNext(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	nextNode := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	trieCtx := getDefaultTrieContext()
	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	data, err := bn.getNext(key, trieCtx)
	assert.NotNil(t, data)

	h1, _ := encodeNodeAndGetHash(nextNode, trieCtx)
	h2, _ := encodeNodeAndGetHash(data.currentNode, trieCtx)
	nextNodeBytes, _ := nextNode.getEncodedNode(trieCtx)
	assert.Equal(t, nextNodeBytes, data.encodedNode)
	assert.Equal(t, h1, h2)
	assert.Equal(t, []byte("dog"), data.hexKey)
	assert.Nil(t, err)
}

func TestBranchNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	key := []byte("dog")

	data, err := bn.getNext(key, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, data)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_getNextNilChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	nilChildPos := byte(4)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	data, err := bn.getNext(key, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, data)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestBranchNode_insert(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	nodeKey := []byte{0, 2, 3}

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(nodeKey), "dogs")}

	newBn := bn.insert(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newBn)
	assert.Nil(t, goRoutinesManager.GetError())

	nodeKeyRemainder := nodeKey[1:]
	bn.children[0] = newLeafNode(getTrieDataWithDefaultVersion(string(nodeKeyRemainder), "dogs"))
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_insertEmptyKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion("", "dogs")}

	newBn := bn.insert(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.Equal(t, ErrValueTooShort, goRoutinesManager.GetError())
	assert.Nil(t, newBn)
}

func TestBranchNode_insertChildPosOutOfRange(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion("dog", "dogs")}

	newBn := bn.insert(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.Equal(t, ErrChildPosOutOfRange, goRoutinesManager.GetError())
	assert.Nil(t, newBn)
}

func TestBranchNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	trieCtx := getDefaultTrieContext()
	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newBn := collapsedBn.insert(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), trieCtx)
	assert.NotNil(t, newBn)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, _ := newBn.tryGet(key, 0, trieCtx)
	assert.Equal(t, []byte("dogs"), val)
}

func TestBranchNode_insertInStoredBnOnExistingPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	trieCtx := getDefaultTrieContext()

	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	expectedHashes := [][]byte{bn.ChildrenHashes[2]}

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := bn.insert(data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())

	assert.True(t, slicesContainSameElements(expectedHashes, modifiedHashes.Get()))
}

func slicesContainSameElements(s1 [][]byte, s2 [][]byte) bool {
	if len(s1) != len(s2) {
		return false
	}

	for _, e1 := range s1 {
		found := false
		for _, e2 := range s2 {
			if bytes.Equal(e1, e2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func TestBranchNode_insertInStoredBnOnNilPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)
	trieCtx := getDefaultTrieContext()

	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	var expectedHashes [][]byte

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := bn.insert(data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.True(t, slicesContainSameElements(expectedHashes, modifiedHashes.Get()))
}

func TestBranchNode_insertInDirtyBnOnNilPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	nilChildPos := byte(11)
	key := append([]byte{nilChildPos}, []byte("dog")...)

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := bn.insert(data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestBranchNode_insertInDirtyBnOnExistingPos(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := bn.insert(data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestBranchNode_delete(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()

	expectedBn, _ := getBnAndCollapsedBn()
	expectedBn.children[2] = nil
	expectedBn.ChildrenHashes[2] = nil

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	data := []core.TrieData{{Key: key}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newBn := bn.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())

	trieCtx := getTrieContextWithCustomStorage(nil)
	expectedBnHash, _ := encodeNodeAndGetHash(expectedBn, trieCtx)
	newBnHash, _ := encodeNodeAndGetHash(newBn, trieCtx)
	assert.Equal(t, expectedBnHash, newBnHash)
}

func TestBranchNode_deleteFromStoredBn(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)
	trieCtx := getDefaultTrieContext()

	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	lnHash := bn.ChildrenHashes[2]
	expectedHashes := [][]byte{lnHash}

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{{Key: lnKey}}
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	dirty, _ := bn.delete(data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.True(t, slicesContainSameElements(expectedHashes, modifiedHashes.Get()))
}

func TestBranchNode_deleteFromDirtyBn(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	childPos := byte(2)
	lnKey := append([]byte{childPos}, []byte("dog")...)
	data := []core.TrieData{{Key: lnKey}}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	dirty, _ := bn.delete(data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestBranchNode_deleteNonexistentNodeFromChild(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("butterfly")...)
	data := []core.TrieData{{Key: key}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newBn := bn.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.False(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_deleteEmptykey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	data := []core.TrieData{{Key: []byte{}}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newBn := bn.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, goRoutinesManager.GetError())
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	bn, collapsedBn := getBnAndCollapsedBn()
	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)

	childPos := byte(2)
	key := append([]byte{childPos}, []byte("dog")...)
	data := []core.TrieData{{Key: key}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newBn := collapsedBn.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), trieCtx)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, err := newBn.tryGet(key, 0, trieCtx)
	assert.Nil(t, val)
	assert.Nil(t, err)
}

func TestBranchNode_deleteAndReduceBn(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	bn.ChildrenHashes[13] = nil
	var children [nrOfChildren]node
	firstChildPos := byte(2)
	secondChildPos := byte(6)
	children[firstChildPos] = newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
	children[secondChildPos] = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"))
	bn.children = children

	key := append([]byte{firstChildPos}, []byte("dog")...)
	ln := newLeafNode(getTrieDataWithDefaultVersion(string(key), "dog"))

	key = append([]byte{secondChildPos}, []byte("doe")...)
	data := []core.TrieData{{Key: key}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newBn := bn.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, ln, newBn)
}

func TestBranchNode_reduceNode(t *testing.T) {
	t.Parallel()

	bn := newBranchNode()
	var children [nrOfChildren]node
	childPos := byte(2)
	children[childPos] = newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
	bn.children = children

	key := append([]byte{childPos}, []byte("dog")...)
	ln := newLeafNode(getTrieDataWithDefaultVersion(string(key), "dog"))

	n, newChildHash, err := bn.children[childPos].reduceNode(int(childPos), bn.ChildrenHashes[childPos], getTrieContextWithCustomStorage(nil))
	assert.Equal(t, ln, n)
	assert.Nil(t, err)
	assert.True(t, newChildHash)
}

func TestBranchNode_getChildPosition(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	nr, pos := getChildPosition(bn)
	assert.Equal(t, 3, nr)
	assert.Equal(t, 13, pos)
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

	expectedTr.Update([]byte("dog"), []byte("dog"))
	expectedTr.Update([]byte("doll"), []byte("doll"))

	tr.Update([]byte("dog"), []byte("dog"))
	tr.Update([]byte("doll"), []byte("doll"))
	tr.Update([]byte("wolf"), []byte("wolf"))
	tr.Delete([]byte("wolf"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithBranchNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	expectedTr, _ := newEmptyTrie()

	expectedTr.Update([]byte("dog"), []byte("puppy"))
	expectedTr.Update([]byte("dogglesworth"), []byte("cat"))

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))
	tr.Delete([]byte("doe"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithLeafNodeChildShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	expectedTr, _ := newEmptyTrie()

	expectedTr.Update([]byte("doe"), []byte("reindeer"))
	expectedTr.Update([]byte("dogglesworth"), []byte("cat"))

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))
	tr.Delete([]byte("dog"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()
	assert.Equal(t, expectedHash, hash)
}

func TestReduceBranchNodeWithLeafNodeValueShouldWork(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	expectedTr, _ := newEmptyTrie()

	expectedTr.Update([]byte("doe"), []byte("reindeer"))
	expectedTr.Update([]byte("dog"), []byte("puppy"))

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))
	tr.Delete([]byte("dogglesworth"))

	expectedHash, _ := expectedTr.RootHash()
	hash, _ := tr.RootHash()

	assert.Equal(t, expectedHash, hash)
}

func TestBranchNode_getChildren(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()

	children, err := bn.getChildren(getTrieContextWithCustomStorage(nil))
	assert.Nil(t, err)
	assert.Equal(t, 3, len(children))
}

func TestBranchNode_getChildrenCollapsedBn(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	bn, collapsedBn := getBnAndCollapsedBn()
	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)

	children, err := collapsedBn.getChildren(trieCtx)
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

	_, hasher := getTestMarshalizerAndHasher()
	tr := initTrie()
	rootNode := tr.GetRootNode()
	nodes, _ := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := cache.NewLRUCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], hasher)
		nodesCacher.Put(n.hash, n, len(n.GetSerialized()))
	}

	firstChildIndex := 5
	secondChildIndex := 7

	bn := getCollapsedBn(t, rootNode)
	trieCtx := getDefaultTrieContext()

	getNode := func(hash []byte) (node, error) {
		cacheData, _ := nodesCacher.Get(hash)
		return trieNode(cacheData, trieCtx)
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
	tr.Update([]byte("aaa"), []byte("aaa"))
	tr.Update([]byte("nnn"), []byte("nnn"))
	tr.Update([]byte("zzz"), []byte("zzz"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()
	collapsedTrie, _ := tr.recreate(rootHash, "", tr.TrieContext)
	collapsedRoot := collapsedTrie.GetRootNode()

	collapsedTrie.Delete([]byte("zzz"))
	ExecuteUpdatesFromBatch(collapsedTrie)

	assert.True(t, collapsedRoot.isDirty())

	_ = collapsedTrie.Commit(hashesCollector.NewDisabledHashesCollector())

	assert.False(t, collapsedRoot.isDirty())
}

func BenchmarkMarshallNodeJson(b *testing.B) {
	bn, _ := getBnAndCollapsedBn()
	marsh := marshal.JsonMarshalizer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = marsh.Marshal(bn)
	}
}

func TestBranchNode_newBranchNodeOkVals(t *testing.T) {
	t.Parallel()

	var children [nrOfChildren]node
	bn := newBranchNode()
	assert.Equal(t, make([][]byte, nrOfChildren), bn.ChildrenHashes)
	assert.Equal(t, children, bn.children)
	assert.True(t, bn.dirty)
}

func TestBranchNode_commitCollapsesTrieIfMaxTrieLevelInMemoryIsReached(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn()

	manager := getTestGoroutinesManager()
	trieCtx := getDefaultTrieContext()
	bn.commitDirty(0, 1, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
	assert.Nil(t, manager.GetError())

	assert.Equal(t, collapsedBn.ChildrenHashes, bn.ChildrenHashes)
	assert.Equal(t, collapsedBn.children, bn.children)
}

func TestBranchNode_reduceNodeBnChild(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	pos := 5
	expectedNode, _ := newExtensionNode([]byte{byte(pos)}, en.child)
	expectedNode.ChildHash = en.ChildHash

	newNode, newChildHash, err := en.child.reduceNode(pos, en.ChildHash, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, err)
	assert.Equal(t, expectedNode, newNode)
	assert.False(t, newChildHash)
}

func TestBranchNode_printShouldNotPanicEvenIfNodeIsCollapsed(t *testing.T) {
	t.Parallel()

	bnWriter := bytes.NewBuffer(make([]byte, 0))
	collapsedBnWriter := bytes.NewBuffer(make([]byte, 0))

	bn, collapsedBn := getBnAndCollapsedBn()
	trieCtx := getDefaultTrieContext()
	collapsedBn.setDirty(false)
	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)

	bn.print(bnWriter, 0, trieCtx)
	collapsedBn.print(collapsedBnWriter, 0, trieCtx)

	assert.Equal(t, bnWriter.Bytes(), collapsedBnWriter.Bytes())
}

func TestBranchNode_getNextHashAndKey(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn()
	proofVerified, nextHash, nextKey := collapsedBn.getNextHashAndKey([]byte{2})

	assert.False(t, proofVerified)
	assert.Equal(t, collapsedBn.ChildrenHashes[2], nextHash)
	assert.Equal(t, []byte{}, nextKey)
}

func TestBranchNode_getNextHashAndKeyNilKey(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn()
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
	bn = &branchNode{
		CollapsedBn: CollapsedBn{
			ChildrenHashes:  [][]byte{collapsed1, collapsed2},
			ChildrenVersion: []byte{1},
		},
		children: [17]node{},
		baseNode: &baseNode{
			dirty: false,
		},
	}
	assert.Equal(t, len(collapsed1)+len(collapsed2)+1+17*pointerSizeInBytes+len(bn.ChildrenVersion)+18*mutexSizeInBytes, bn.sizeInBytes())
}

func TestBranchNode_commitSnapshotContextDone(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bn.commitSnapshot(getDefaultTrieContext(), nil, nil, ctx, statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, []byte("nodeBytes"), 0)
	assert.Equal(t, core.ErrContextClosing, err)
}

func TestBranchNode_commitSnapshotDbIsClosing(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	db.GetCalled = func(key []byte) ([]byte, error) {
		return nil, core.ErrContextClosing
	}
	args := GetDefaultTrieStorageManagerParameters()
	args.MainStorer = db
	trieStorage, _ := NewTrieStorageManager(args)
	trieCtx := getTrieContextWithCustomStorage(trieStorage)

	_, collapsedBn := getBnAndCollapsedBn()
	nodeBytes, _ := collapsedBn.getEncodedNode(trieCtx)
	missingNodesChan := make(chan []byte, 10)
	err := collapsedBn.commitSnapshot(trieCtx, nil, missingNodesChan, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, nodeBytes, 0)
	assert.True(t, core.IsClosingError(err))
	assert.Equal(t, 0, len(missingNodesChan))
}

func TestBranchNode_commitSnapshotChildIsMissingErr(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	db.GetCalled = func(key []byte) ([]byte, error) {
		return nil, core.NewGetNodeFromDBErrWithKey(key, ErrKeyNotFound, "test")
	}
	args := GetDefaultTrieStorageManagerParameters()
	args.MainStorer = db
	trieStorage, _ := NewTrieStorageManager(args)
	trieCtx := getTrieContextWithCustomStorage(trieStorage)

	_, collapsedBn := getBnAndCollapsedBn()
	missingNodesChan := make(chan []byte, 10)
	nodeBytes, _ := collapsedBn.getEncodedNode(trieCtx)
	err := collapsedBn.commitSnapshot(trieCtx, nil, missingNodesChan, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, nodeBytes, 0)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(missingNodesChan))
}

func TestBranchNode_getVersion(t *testing.T) {
	t.Parallel()

	t.Run("nil ChildrenVersion", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()

		version, err := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Nil(t, err)
	})

	t.Run("NotSpecified for all children", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()
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

		bn, _ := getBnAndCollapsedBn()
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

		bn, _ := getBnAndCollapsedBn()
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

	bn, _ := getBnAndCollapsedBn()
	assert.Equal(t, []byte{}, bn.getValue())
}

func TestBranchNode_VerifyChildrenVersionIsSetCorrectlyAfterInsertAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("revert child from version 1 to 0", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.AutoBalanceEnabled)

		goRoutinesManager := getTestGoroutinesManager()
		childKey := []byte{2, 'd', 'o', 'g'}
		data := core.TrieData{
			Key:     childKey,
			Value:   []byte("value"),
			Version: 0,
		}

		newBn := bn.insert([]core.TrieData{data}, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getDefaultTrieContext())
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Nil(t, newBn.(*branchNode).ChildrenVersion)
	})

	t.Run("remove migrated child", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()
		bn.ChildrenVersion = make([]byte, nrOfChildren)
		bn.ChildrenVersion[2] = byte(core.AutoBalanceEnabled)
		childKey := []byte{2, 'd', 'o', 'g'}
		data := []core.TrieData{{Key: childKey}}

		goRoutinesManager := getTestGoroutinesManager()
		_, newBn := bn.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getDefaultTrieContext())
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
		tr.Update([]byte("111"), []byte("111"))
		tr.Update([]byte("aaa"), []byte("aaa"))
		_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())

		bn, ok := tr.GetRootNode().(*branchNode)
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
		assert.Equal(t, bn.ChildrenHashes[1], firstChildData.GetData())
		assert.Equal(t, uint64(hashSize+keySize), firstChildData.Size())
		assert.False(t, firstChildData.IsLeaf())

		// leaf node as child
		seconChildData := nodeData[1]
		assert.Equal(t, uint(1), seconChildData.GetKeyBuilder().Size())
		assert.Equal(t, bn.ChildrenHashes[5], seconChildData.GetData())
		assert.Equal(t, uint64(hashSize+keySize), seconChildData.Size())
		assert.False(t, seconChildData.IsLeaf())

		// extension node as child
		thirdChildData := nodeData[2]
		assert.Equal(t, uint(1), thirdChildData.GetKeyBuilder().Size())
		assert.Equal(t, bn.ChildrenHashes[7], thirdChildData.GetData())
		assert.Equal(t, uint64(hashSize+keySize), thirdChildData.Size())
		assert.False(t, thirdChildData.IsLeaf())
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

		bn, _ := getBnAndCollapsedBn()
		var data []core.TrieData

		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		bnModified := &atomic.Flag{}
		bn.insertOnNilChild(data, 0, goRoutinesManager, modifiedHashes, bnModified, nil)
		expectedNumTrieNodesChanged := 0
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		assert.Equal(t, ErrValueTooShort, goRoutinesManager.GetError())
		assert.False(t, bnModified.IsSet())
	})
	t.Run("insert one child in !dirty node", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		bn.commitDirty(0, 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())
		assert.False(t, bn.dirty)
		newData := []core.TrieData{
			{
				Key:     []byte{1, 2, 3},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
		}
		childPos := byte(0)
		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		bnModified := &atomic.Flag{}
		bn.insertOnNilChild(newData, childPos, goRoutinesManager, modifiedHashes, bnModified, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Equal(t, 0, len(modifiedHashes.Get()))
		assert.True(t, bn.dirty)
		assert.NotNil(t, bn.children[childPos])
		assert.Equal(t, byte(core.AutoBalanceEnabled), bn.ChildrenVersion[childPos])
		assert.True(t, bnModified.IsSet())
	})
	t.Run("insert one child in dirty node", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()
		assert.True(t, bn.dirty)
		newData := []core.TrieData{
			{
				Key:     []byte{1, 2, 3},
				Value:   []byte("value"),
				Version: core.AutoBalanceEnabled,
			},
		}
		childPos := byte(0)
		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		bnModified := &atomic.Flag{}
		bn.insertOnNilChild(newData, childPos, goRoutinesManager, modifiedHashes, bnModified, getDefaultTrieContext())
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 0
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		assert.True(t, bn.dirty)
		assert.NotNil(t, bn.children[childPos])
		assert.Equal(t, byte(core.AutoBalanceEnabled), bn.ChildrenVersion[childPos])
		assert.True(t, bnModified.IsSet())
	})
	t.Run("insert multiple children", func(t *testing.T) {
		t.Parallel()

		bn, _ := getBnAndCollapsedBn()
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
		childPos := byte(0)
		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		bnModified := &atomic.Flag{}
		bn.insertOnNilChild(newData, childPos, goRoutinesManager, modifiedHashes, bnModified, getDefaultTrieContext())
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 0
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		assert.True(t, bn.dirty)
		assert.NotNil(t, bn.children[childPos])
		assert.Equal(t, byte(core.AutoBalanceEnabled), bn.ChildrenVersion[childPos])
		_, ok := bn.children[0].(*extensionNode)
		assert.True(t, ok)
		assert.True(t, bnModified.IsSet())
	})
}

func TestBranchNode_insertOnExistingChild(t *testing.T) {
	t.Parallel()

	t.Run("insert on existing child multiple children", func(t *testing.T) {
		t.Parallel()

		trieCtx := getDefaultTrieContext()
		childPos := byte(2)
		var children [nrOfChildren]node
		children[2] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 5}), "dog"))
		children[6] = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"))
		bn := newBranchNode()
		bn.children = children
		bn.ChildrenHashes[2], _ = encodeNodeAndGetHash(bn.children[2], trieCtx)
		bn.ChildrenHashes[6], _ = encodeNodeAndGetHash(bn.children[6], trieCtx)
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
		manager := getTestGoroutinesManager()
		bn.commitDirty(0, 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())
		assert.False(t, bn.dirty)
		originalHash, _ := encodeNodeAndGetHash(bn, trieCtx)
		assert.True(t, len(originalHash) > 0)
		originalChildHash := bn.ChildrenHashes[childPos]
		assert.True(t, len(originalChildHash) > 0)

		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		bnModified := &atomic.Flag{}
		bn.insertOnChild(newData, int(childPos), goRoutinesManager, modifiedHashes, bnModified, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, bnModified.IsSet())
		assert.True(t, bn.dirty)
		expectedNumTrieNodesChanged := 1
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))

		originalChildHashPresent := false
		for _, hash := range modifiedHashes.Get() {
			if bytes.Equal(hash, originalChildHash) {
				originalChildHashPresent = true
			}
		}
		assert.True(t, originalChildHashPresent)

		_, ok := bn.children[childPos].(*extensionNode)
		assert.True(t, ok)
	})
	t.Run("insert on existing child same node", func(t *testing.T) {
		t.Parallel()

		trieCtx := getDefaultTrieContext()
		childPos := byte(2)
		var children [nrOfChildren]node
		key := []byte{1, 2, 5}
		value := "dog"
		children[2] = newLeafNode(getTrieDataWithDefaultVersion(string(key), value))
		children[6] = newLeafNode(getTrieDataWithDefaultVersion("doe", "doe"))
		bn := newBranchNode()
		bn.children = children
		bn.ChildrenHashes[2], _ = encodeNodeAndGetHash(bn.children[2], trieCtx)
		bn.ChildrenHashes[6], _ = encodeNodeAndGetHash(bn.children[6], trieCtx)
		newData := []core.TrieData{
			{
				Key:     key,
				Value:   []byte(value),
				Version: core.NotSpecified,
			},
		}
		manager := getTestGoroutinesManager()
		bn.commitDirty(0, 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())
		assert.False(t, bn.dirty)

		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		bnModified := &atomic.Flag{}
		bn.insertOnChild(newData, int(childPos), goRoutinesManager, modifiedHashes, bnModified, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.False(t, bnModified.IsSet())
		assert.False(t, bn.dirty)
		expectedNumTrieNodesChanged := 0
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		_, ok := bn.children[childPos].(*leafNode)
		assert.True(t, ok)
	})
}

func TestBranchNode_insertBatch(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	var children [nrOfChildren]node
	children[2] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{3, 4, 5}), "dog"))
	children[6] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{7, 8, 9}), "doe"))
	bn := newBranchNode()
	bn.children = children
	childHash1, _ := encodeNodeAndGetHash(children[2], trieCtx)
	childHash2, _ := encodeNodeAndGetHash(children[6], trieCtx)
	bn.ChildrenHashes[2] = childHash1
	bn.ChildrenHashes[6] = childHash2

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
	manager := getTestGoroutinesManager()
	bn.commitDirty(0, 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
	assert.Nil(t, manager.GetError())
	assert.False(t, bn.dirty)

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := bn.insert(newData, goRoutinesManager, modifiedHashes, trieCtx)
	assert.Nil(t, goRoutinesManager.GetError())
	expectedNumTrieNodesChanged := 1
	assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
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
	trieCtx := getDefaultTrieContext()
	var children [nrOfChildren]node
	childBn := newBranchNode()
	childBn.children[1] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{3, 4, 5}), "dog"))
	childBn.children[3] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{7, 8, 9}), "doe"))
	childBn.ChildrenHashes[1], _ = encodeNodeAndGetHash(childBn.children[1], trieCtx)
	childBn.ChildrenHashes[3], _ = encodeNodeAndGetHash(childBn.children[3], trieCtx)

	children[4] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{3, 4, 5}), "dog"))
	children[7] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{7, 8, 9}), "doe"))
	children[9] = childBn
	bn := newBranchNode()
	bn.children = children
	bn.ChildrenHashes[4], _ = encodeNodeAndGetHash(bn.children[4], trieCtx)
	bn.ChildrenHashes[7], _ = encodeNodeAndGetHash(bn.children[7], trieCtx)
	bn.ChildrenHashes[9], _ = encodeNodeAndGetHash(childBn, trieCtx)

	bn.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
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

		trieCtx := getDefaultTrieContext()
		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := bn.delete(data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, dirty)
		assert.True(t, newNode.isDirty())
		expectedNumTrieNodesChanged := 4
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
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

		trieCtx := getDefaultTrieContext()
		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := bn.delete(data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, dirty)
		assert.True(t, newNode.isDirty())
		expectedNumTrieNodesChanged := 5
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
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

		trieCtx := getDefaultTrieContext()
		goRoutinesManager := getTestGoroutinesManager()
		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := bn.delete(data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, dirty)
		assert.Nil(t, newNode)
		expectedNumTrieNodesChanged := 5
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
	})
}
