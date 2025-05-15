package trie

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/state/hashesCollector"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/stretchr/testify/assert"
)

func getLn() *leafNode {
	newLn := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
	return newLn
}

func TestLeafNode_newLeafNode(t *testing.T) {
	t.Parallel()

	expectedLn := &leafNode{
		CollapsedLn: CollapsedLn{
			Key:   []byte("dog"),
			Value: []byte("dog"),
		},
		baseNode: &baseNode{
			dirty: true,
		},
	}
	ln := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"))
	assert.Equal(t, expectedLn, ln)
}

func TestLeafNode_isDirty(t *testing.T) {
	t.Parallel()

	ln := &leafNode{baseNode: &baseNode{dirty: true}}
	assert.Equal(t, true, ln.isDirty())

	ln = &leafNode{baseNode: &baseNode{dirty: false}}
	assert.Equal(t, false, ln.isDirty())
}

func TestLeafNode_commit(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	ln := getLn()
	hash, _ := encodeNodeAndGetHash(ln, trieCtx)

	manager := getTestGoroutinesManager()
	ln.commitDirty(0, 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
	assert.Nil(t, manager.GetError())

	encNode, _ := trieCtx.Get(hash)
	n, _ := decodeNode(encNode, trieCtx)
	ln = getLn()
	ln.dirty = false
	assert.Equal(t, ln, n)
}

func TestLeafNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getTrieContextWithCustomStorage(nil)
	ln := getLn()
	expectedEncodedNode, _ := trieCtx.Marshal(ln)
	expectedEncodedNode = append(expectedEncodedNode, leaf)

	encNode, err := ln.getEncodedNode(trieCtx)
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestLeafNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()

	ln := &leafNode{}

	encNode, err := ln.getEncodedNode(getTrieContextWithCustomStorage(nil))
	assert.True(t, errors.Is(err, ErrEmptyLeafNode))
	assert.Nil(t, encNode)
}

func TestLeafNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var ln *leafNode

	encNode, err := ln.getEncodedNode(getTrieContextWithCustomStorage(nil))
	assert.True(t, errors.Is(err, ErrNilLeafNode))
	assert.Nil(t, encNode)
}

func TestLeafNode_tryGet(t *testing.T) {
	t.Parallel()

	ln := getLn()
	key := []byte("dog")

	val, maxDepth, err := ln.tryGet(key, 0, getTrieContextWithCustomStorage(nil))
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestLeafNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn()
	wrongKey := []byte{1, 2, 3}

	val, maxDepth, err := ln.tryGet(wrongKey, 0, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestLeafNode_getNext(t *testing.T) {
	t.Parallel()

	ln := getLn()
	key := []byte("dog")

	data, err := ln.getNext(key, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, data)
	assert.Nil(t, err)
}

func TestLeafNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	ln := getLn()
	wrongKey := append([]byte{2}, []byte("dog")...)

	data, err := ln.getNext(wrongKey, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, data)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestLeafNode_insertAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn()
	key := "dog"
	expectedVal := "dogs"
	data := []core.TrieData{getTrieDataWithDefaultVersion(key, expectedVal)}

	goRoutinesManager := getTestGoroutinesManager()
	newNode := ln.insert(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, _ := newNode.tryGet([]byte(key), 0, getTrieContextWithCustomStorage(nil))
	assert.Equal(t, []byte(expectedVal), val)
}

func TestLeafNode_insertAtDifferentKey(t *testing.T) {
	t.Parallel()

	lnKey := []byte{2, 100, 111, 103}
	ln := newLeafNode(getTrieDataWithDefaultVersion(string(lnKey), "dog"))

	nodeKey := []byte{3, 4, 5}
	nodeVal := []byte{3, 4, 5}
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(nodeKey), string(nodeVal))}

	goRoutinesManager := getTestGoroutinesManager()
	newNode := ln.insert(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, _ := newNode.tryGet(nodeKey, 0, getTrieContextWithCustomStorage(nil))
	assert.Equal(t, nodeVal, val)
	assert.IsType(t, &branchNode{}, newNode)
}

func TestLeafNode_insertInStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	ln := getLn()
	ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	data := []core.TrieData{getTrieDataWithDefaultVersion("dog", "dogs")}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := ln.insert(data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestLeafNode_insertInStoredLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3}), "dog"))
	ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	data := []core.TrieData{getTrieDataWithDefaultVersion(string([]byte{4, 5, 6}), "dogs")}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := ln.insert(data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestLeafNode_insertInDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn()
	data := []core.TrieData{getTrieDataWithDefaultVersion("dog", "dogs")}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := ln.insert(data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestLeafNode_insertInDirtyLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3}), "dog"))
	data := []core.TrieData{getTrieDataWithDefaultVersion(string([]byte{4, 5, 6}), "dogs")}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := ln.insert(data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestLeafNode_deletePresent(t *testing.T) {
	t.Parallel()

	ln := getLn()
	data := []core.TrieData{{Key: []byte("dog")}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newNode := ln.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Nil(t, newNode)
}

func TestLeafNode_deleteFromStoredLnAtSameKey(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	ln := getLn()
	ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	data := []core.TrieData{{Key: []byte("dog")}}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	dirty, _ := ln.delete(data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestLeafNode_deleteFromLnAtDifferentKey(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	ln := getLn()
	ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
	wrongKey := []byte{1, 2, 3}
	data := []core.TrieData{{Key: wrongKey}}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	dirty, _ := ln.delete(data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.False(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestLeafNode_deleteFromDirtyLnAtSameKey(t *testing.T) {
	t.Parallel()

	ln := getLn()
	data := []core.TrieData{{Key: []byte("dog")}}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	dirty, _ := ln.delete(data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestLeafNode_deleteNotPresent(t *testing.T) {
	t.Parallel()

	ln := getLn()
	wrongKey := []byte{1, 2, 3}
	data := []core.TrieData{{Key: wrongKey}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newNode := ln.delete(data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), nil)
	assert.False(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, ln, newNode)
}

func TestLeafNode_reduceNode(t *testing.T) {
	t.Parallel()

	ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{100, 111, 103}), ""))
	expected := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{2, 100, 111, 103}), ""))
	expected.dirty = true

	n, newChildHash, err := ln.reduceNode(2, nil, nil)
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

	ln := getLn()

	children, err := ln.getChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(children))
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
	nodes, _ := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := cache.NewLRUCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], hasher)
		nodesCacher.Put(n.hash, n, len(n.GetSerialized()))
	}

	ln := &leafNode{baseNode: &baseNode{}}
	missing, _, err := ln.loadChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 6, nodesCacher.Len())
	assert.Equal(t, 0, len(missing))
}

func TestInsertSameNodeShouldNotSetDirtyBnRoot(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash := tr.GetRootHash()

	tr.Update([]byte("dog"), []byte("puppy"))
	ExecuteUpdatesFromBatch(tr)
	assert.False(t, tr.GetRootNode().isDirty())
	assert.Equal(t, rootHash, tr.GetRootHash())
	assert.Equal(t, [][]byte{}, tr.GetOldHashes())
}

func TestInsertSameNodeShouldNotSetDirtyEnRoot(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("log"), []byte("wood"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash := tr.GetRootHash()

	tr.Update([]byte("dog"), []byte("puppy"))
	ExecuteUpdatesFromBatch(tr)
	assert.False(t, tr.GetRootNode().isDirty())
	assert.Equal(t, rootHash, tr.GetRootHash())
	assert.Equal(t, [][]byte{}, tr.GetOldHashes())
}

func TestInsertSameNodeShouldNotSetDirtyLnRoot(t *testing.T) {
	t.Parallel()

	tr, _ := newEmptyTrie()
	tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash := tr.GetRootHash()

	tr.Update([]byte("dog"), []byte("puppy"))
	ExecuteUpdatesFromBatch(tr)
	assert.False(t, tr.GetRootNode().isDirty())
	assert.Equal(t, rootHash, tr.GetRootHash())
	assert.Equal(t, [][]byte{}, tr.GetOldHashes())
}

func TestLeafNode_deleteDifferentKeyShouldNotModifyTrie(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash := tr.GetRootHash()

	tr.Update([]byte("ddoe"), []byte{})
	ExecuteUpdatesFromBatch(tr)
	assert.False(t, tr.GetRootNode().isDirty())
	assert.Equal(t, rootHash, tr.GetRootHash())
	assert.Equal(t, [][]byte{}, tr.GetOldHashes())
}

func TestLeafNode_newLeafNodeOkVals(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	val := []byte("val")
	ln := newLeafNode(getTrieDataWithDefaultVersion("key", "val"))

	assert.Equal(t, key, ln.Key)
	assert.Equal(t, val, ln.Value)
	assert.True(t, ln.dirty)
}

func TestLeafNode_getNextHashAndKey(t *testing.T) {
	t.Parallel()

	ln := getLn()
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
	ln = &leafNode{
		CollapsedLn: CollapsedLn{
			Key:     key,
			Value:   value,
			Version: 1,
		},
		baseNode: &baseNode{
			dirty: false,
		},
	}
	assert.Equal(t, len(key)+len(value)+1+4, ln.sizeInBytes())
}

func TestLeafNode_writeNodeOnChannel(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	ln := getLn()
	lnHash, _ := encodeNodeAndGetHash(ln, trieCtx)
	leavesChannel := make(chan core.KeyValueHolder, 2)

	err := writeNodeOnChannel(ln, leavesChannel, trieCtx)
	assert.Nil(t, err)

	retrievedLn := <-leavesChannel
	assert.Equal(t, lnHash, retrievedLn.Key())
	assert.Equal(t, ln.Value, retrievedLn.Value())
}

func TestLeafNode_commitSnapshotContextDone(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	ln := getLn()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ln.commitSnapshot(trieCtx, nil, nil, ctx, statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, []byte("nodeBytes"), 0)
	assert.Equal(t, core.ErrContextClosing, err)
}

func TestLeafNode_getValue(t *testing.T) {
	t.Parallel()

	ln := getLn()
	assert.Equal(t, ln.Value, ln.getValue())
}

func TestLeafNode_getVersion(t *testing.T) {
	t.Parallel()

	t.Run("invalid node version", func(t *testing.T) {
		t.Parallel()

		ln := getLn()
		ln.Version = math.MaxUint8 + 1

		version, err := ln.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Equal(t, ErrInvalidNodeVersion, err)
	})

	t.Run("NotSpecified version", func(t *testing.T) {
		t.Parallel()

		ln := getLn()

		version, err := ln.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Nil(t, err)
	})

	t.Run("AutoBalanceEnabled version", func(t *testing.T) {
		t.Parallel()

		ln := getLn()
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

		ln := getLn()
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

func TestLeafNode_insertBatch(t *testing.T) {
	t.Parallel()

	t.Run("insert in same leaf node different val", func(t *testing.T) {
		t.Parallel()

		ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dog"))
		trieCtx := getDefaultTrieContext()
		newData := []core.TrieData{getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dogs")}
		ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.False(t, ln.dirty)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := ln.insert(newData, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.True(t, newNode.isDirty())
		assert.Equal(t, [][]byte{}, modifiedHashes.Get())
		assert.Equal(t, []byte("dogs"), newNode.(*leafNode).Value)
	})
	t.Run("insert in same leaf node same val", func(t *testing.T) {
		t.Parallel()

		ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dog"))
		trieCtx := getDefaultTrieContext()
		newData := []core.TrieData{getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dog")}
		ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.False(t, ln.dirty)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := ln.insert(newData, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Nil(t, newNode)
		assert.Equal(t, 0, len(modifiedHashes.Get()))
	})
	t.Run("branch at the beginning after insert", func(t *testing.T) {
		t.Parallel()

		ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dog"))
		trieCtx := getDefaultTrieContext()
		newData := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dogs"),
			getTrieDataWithDefaultVersion(string([]byte{2, 3, 4, 5, 16}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{3, 4, 5, 6, 16}), "dog"),
		}
		ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.False(t, ln.dirty)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := ln.insert(newData, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Equal(t, [][]byte{}, modifiedHashes.Get())
		bn, ok := newNode.(*branchNode)
		assert.True(t, ok)
		assert.Equal(t, []byte("dogs"), bn.children[1].(*leafNode).Value)
		assert.NotNil(t, []byte("dog"), bn.children[2].(*leafNode).Value)
		assert.NotNil(t, []byte("dog"), bn.children[3].(*leafNode).Value)
		assert.True(t, bn.dirty)
		bnHash, _ := encodeNodeAndGetHash(bn, trieCtx)
		assert.NotEqual(t, 0, len(bnHash))
	})
	t.Run("extension node at the beginning after insert ", func(t *testing.T) {
		t.Parallel()

		ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dog"))
		trieCtx := getDefaultTrieContext()
		newData := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dogs"),
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 4, 5, 16}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 5, 6, 16}), "dog"),
		}
		ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.False(t, ln.dirty)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := ln.insert(newData, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Equal(t, [][]byte{}, modifiedHashes.Get())
		en, ok := newNode.(*extensionNode)
		assert.True(t, ok)
		assert.Equal(t, []byte{1, 2}, en.Key)
		assert.True(t, en.dirty)
		enHash, _ := encodeNodeAndGetHash(en, trieCtx)
		assert.NotEqual(t, 0, len(enHash))
	})
}

func TestLeafNode_deleteBatch(t *testing.T) {
	t.Parallel()

	t.Run("delete existing", func(t *testing.T) {
		t.Parallel()

		ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dog"))
		trieCtx := getDefaultTrieContext()
		newData := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), ""),
			getTrieDataWithDefaultVersion(string([]byte{2, 2, 3, 4, 16}), ""),
			getTrieDataWithDefaultVersion(string([]byte{3, 2, 3, 4, 16}), ""),
		}
		ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := ln.delete(newData, goRoutinesManager, modifiedHashes, trieCtx)
		assert.True(t, dirty)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Nil(t, newNode)
		assert.Equal(t, [][]byte{}, modifiedHashes.Get())
	})
	t.Run("delete not existing", func(t *testing.T) {
		t.Parallel()

		ln := newLeafNode(getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 16}), "dog"))
		trieCtx := getDefaultTrieContext()
		newData := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{2, 2, 3, 4, 16}), ""),
			getTrieDataWithDefaultVersion(string([]byte{3, 2, 3, 4, 16}), ""),
		}
		ln.commitDirty(0, 5, getTestGoroutinesManager(), hashesCollector.NewDisabledHashesCollector(), trieCtx)

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := ln.delete(newData, goRoutinesManager, modifiedHashes, trieCtx)
		assert.False(t, dirty)
		assert.Nil(t, goRoutinesManager.GetError())
		assert.Equal(t, ln, newNode)
		assert.Equal(t, [][]byte{}, modifiedHashes.Get())
	})
}
