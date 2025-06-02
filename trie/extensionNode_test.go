package trie

import (
	"bytes"
	"context"
	"errors"
	"math"
	"strconv"
	sync2 "sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state/hashesCollector"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/stretchr/testify/assert"
)

func getEnAndCollapsedEn() (*extensionNode, *extensionNode) {
	child, collapsedChild := getBnAndCollapsedBn()
	en, _ := newExtensionNode([]byte("d"), child)

	childHash, _ := encodeNodeAndGetHash(collapsedChild, getTrieContextWithCustomStorage(nil))
	collapsedEn := &extensionNode{CollapsedEn: CollapsedEn{Key: []byte("d"), ChildHash: childHash}}
	en.ChildHash = childHash
	return en, collapsedEn
}

func TestExtensionNode_newExtensionNode(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	expectedEn := &extensionNode{
		CollapsedEn: CollapsedEn{
			Key:       []byte("dog"),
			ChildHash: nil,
		},
		child: bn,
		dirty: true,
	}
	en, _ := newExtensionNode([]byte("dog"), bn)
	assert.Equal(t, expectedEn, en)
}

func TestExtensionNode_isDirty(t *testing.T) {
	t.Parallel()

	en := &extensionNode{dirty: true}
	assert.Equal(t, true, en.isDirty())

	en = &extensionNode{dirty: false}
	assert.Equal(t, false, en.isDirty())
}

func TestExtensionNode_commit(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, collapsedEn := getEnAndCollapsedEn()
	hash, _ := encodeNodeAndGetHash(collapsedEn, trieCtx)

	manager := getTestGoroutinesManager()
	en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
	assert.Nil(t, manager.GetError())

	encNode, _ := trieCtx.Get(hash)
	n, _ := decodeNode(encNode, trieCtx)

	h1, _ := encodeNodeAndGetHash(collapsedEn, trieCtx)
	h2, _ := encodeNodeAndGetHash(n, trieCtx)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_commitCollapsedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	_, collapsedEn := getEnAndCollapsedEn()
	hash, _ := encodeNodeAndGetHash(collapsedEn, trieCtx)

	collapsedEn.dirty = true
	manager := getTestGoroutinesManager()
	collapsedEn.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
	assert.Nil(t, manager.GetError())

	encNode, _ := trieCtx.Get(hash)
	n, _ := decodeNode(encNode, trieCtx)

	h1, _ := encodeNodeAndGetHash(collapsedEn, trieCtx)
	h2, _ := encodeNodeAndGetHash(n, trieCtx)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getTrieContextWithCustomStorage(nil)
	en, _ := getEnAndCollapsedEn()
	expectedEncodedNode, _ := trieCtx.Marshal(en)
	expectedEncodedNode = append(expectedEncodedNode, extension)

	encNode, err := en.getEncodedNode(trieCtx)
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestExtensionNode_resolveIfCollapsed(t *testing.T) {
	t.Parallel()

	t.Run("resolve collapsed node", func(t *testing.T) {
		t.Parallel()

		trieCtx := getDefaultTrieContext()
		en, collapsedEn := getEnAndCollapsedEn()
		_ = saveNodeToStorage(en, trieCtx)
		_, resolved := getBnAndCollapsedBn()

		child, childHash, err := collapsedEn.resolveIfCollapsed(trieCtx)
		assert.Nil(t, err)
		assert.Equal(t, en.child.(*branchNode).ChildrenHashes[2], child.(*branchNode).ChildrenHashes[2])
		assert.Equal(t, en.child.(*branchNode).ChildrenHashes[6], child.(*branchNode).ChildrenHashes[6])
		assert.Equal(t, en.child.(*branchNode).ChildrenHashes[13], child.(*branchNode).ChildrenHashes[13])
		assert.Equal(t, en.ChildHash, childHash)

		h1, _ := encodeNodeAndGetHash(resolved, trieCtx)
		h2, _ := encodeNodeAndGetHash(collapsedEn.child, trieCtx)
		assert.Equal(t, h1, h2)
	})
	t.Run("invalid node state", func(t *testing.T) {
		t.Parallel()

		en, _ := getEnAndCollapsedEn()
		en.ChildHash = nil

		child, childHash, err := en.resolveIfCollapsed(getTrieContextWithCustomStorage(nil))
		assert.Equal(t, ErrInvalidNodeState, err)
		assert.Nil(t, child)
		assert.Nil(t, childHash)
	})
	t.Run("node not collapsed", func(t *testing.T) {
		t.Parallel()

		en, _ := getEnAndCollapsedEn()

		child, childHash, err := en.resolveIfCollapsed(getTrieContextWithCustomStorage(nil))
		assert.Nil(t, err)
		assert.Equal(t, en.child, child)
		assert.Equal(t, en.ChildHash, childHash)
	})
}

func TestExtensionNode_isCollapsed(t *testing.T) {
	t.Parallel()

	en, collapsedEn := getEnAndCollapsedEn()
	assert.True(t, collapsedEn.isCollapsed())
	assert.False(t, en.isCollapsed())

	collapsedEn.child = newLeafNode(getTrieDataWithDefaultVersion("og", "dog"))
	assert.False(t, collapsedEn.isCollapsed())
}

func TestExtensionNode_tryGet(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	dogBytes := []byte("dog")

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := dogBytes
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)

	val, maxDepth, err := en.tryGet(newKeyData(key), 0, getTrieContextWithCustomStorage(nil))
	assert.Equal(t, dogBytes, val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), maxDepth)
}

func TestExtensionNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	var key []byte

	val, maxDepth, err := en.tryGet(newKeyData(key), 0, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestExtensionNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	key := []byte("gdo")

	val, maxDepth, err := en.tryGet(newKeyData(key), 0, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, uint32(0), maxDepth)
}

func TestExtensionNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = saveNodeToStorage(en, trieCtx)

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)

	val, maxDepth, err := collapsedEn.tryGet(newKeyData(key), 0, trieCtx)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
	assert.Equal(t, uint32(2), maxDepth)
}

func TestExtensionNode_getNext(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	trieCtx := getDefaultTrieContext()
	_ = saveNodeToStorage(en, trieCtx)

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)

	data, err := en.getNext(key, trieCtx)
	child, childBytes, _ := getNodeFromDBAndDecode(en.ChildHash, trieCtx)
	assert.NotNil(t, data)
	assert.Equal(t, childBytes, data.encodedNode)
	assert.Equal(t, child, data.currentNode)
	assert.Equal(t, key[1:], data.hexKey)
	assert.Nil(t, err)
}

func TestExtensionNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(bnKey, lnKey...)

	data, err := en.getNext(key, getTrieContextWithCustomStorage(nil))
	assert.Nil(t, data)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestExtensionNode_insert(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	key := []byte{100, 15, 5, 6}

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, _ := newNode.tryGet(newKeyData(key), 0, getTrieContextWithCustomStorage(nil))
	assert.Equal(t, []byte("dogs"), val)
}

func TestExtensionNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, collapsedEn := getEnAndCollapsedEn()
	key := []byte{100, 15, 5, 6}

	_ = saveNodeToStorage(en, trieCtx)
	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	newNode := collapsedEn.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), trieCtx)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())

	val, _, _ := newNode.tryGet(newKeyData(key), 0, trieCtx)
	assert.Equal(t, []byte("dogs"), val)
}

func TestExtensionNode_insertInStoredEnSameKey(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, _ := getEnAndCollapsedEn()
	enKey := []byte{100}
	key := append(enKey, []byte{11, 12}...)

	_ = saveNodeToStorage(en, trieCtx)
	bnHash := en.ChildHash
	expectedHashes := [][]byte{bnHash}

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(key), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, expectedHashes, modifiedHashes.Get())
}

func TestExtensionNode_insertInStoredEnDifferentKey(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	bn, _ := getBnAndCollapsedBn()
	enKey := []byte{1}
	en, _ := newExtensionNode(enKey, bn)
	nodeKey := []byte{11, 12}
	childHash, _ := encodeNodeAndGetHash(en.child, trieCtx)
	en.ChildHash = childHash

	_ = saveNodeToStorage(en, trieCtx)
	expectedHashes := [][]byte{childHash}

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(nodeKey), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.True(t, slicesContainSameElements(expectedHashes, modifiedHashes.Get()))
}

func TestExtensionNode_insertInDirtyEnSameKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	nodeKey := []byte{100, 11, 12}

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(nodeKey), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestExtensionNode_insertInDirtyEnDifferentKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	enKey := []byte{1}
	en, _ := newExtensionNode(enKey, bn)
	nodeKey := []byte{11, 12}
	childHash, _ := encodeNodeAndGetHash(en.child, getTrieContextWithCustomStorage(nil))
	en.ChildHash = childHash

	goRoutinesManager := getTestGoroutinesManager()
	data := []core.TrieData{getTrieDataWithDefaultVersion(string(nodeKey), "dogs")}

	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.NotNil(t, newNode)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestExtensionNode_delete(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	dogBytes := []byte("dog")

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := dogBytes
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)
	trieCtx := getTrieContextWithCustomStorage(nil)

	val, _, _ := en.tryGet(newKeyData(key), 0, trieCtx)
	assert.Equal(t, dogBytes, val)
	data := []core.TrieData{{Key: key}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, _ := en.delete(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), trieCtx)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	val, _, _ = en.tryGet(newKeyData(key), 0, trieCtx)
	assert.Nil(t, val)
}

func TestExtensionNode_deleteFromStoredEn(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, _ := getEnAndCollapsedEn()
	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)
	lnPathKey := key

	_ = saveNodeToStorage(en, trieCtx)
	bnHash := en.ChildHash
	lnHash := en.child.(*branchNode).ChildrenHashes[2]
	expectedHashes := [][]byte{lnHash, bnHash}
	data := []core.TrieData{{Key: lnPathKey}}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	dirty, _ := en.delete(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.True(t, slicesContainSameElements(expectedHashes, modifiedHashes.Get()))
}

func TestExtensionNode_deleteFromDirtyEn(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	lnKey := []byte{100, 2, 100, 111, 103}
	data := []core.TrieData{{Key: lnKey}}

	goRoutinesManager := getTestGoroutinesManager()
	modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
	dirty, _ := en.delete(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, getTrieContextWithCustomStorage(nil))
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	assert.Equal(t, [][]byte{}, modifiedHashes.Get())
}

func TestExtensionNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = saveNodeToStorage(en, trieCtx)

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)

	val, _, _ := en.tryGet(newKeyData(key), 0, trieCtx)
	assert.Equal(t, []byte("dog"), val)
	data := []core.TrieData{{Key: key}}

	goRoutinesManager := getTestGoroutinesManager()
	dirty, newNode := collapsedEn.delete(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), trieCtx)
	assert.True(t, dirty)
	assert.Nil(t, goRoutinesManager.GetError())
	val, _, _ = newNode.tryGet(newKeyData(key), 0, trieCtx)
	assert.Nil(t, val)
}

func TestExtensionNode_reduceNode(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn()
	en, _ := newExtensionNode([]byte{100, 111, 103}, bn)
	childHash, _ := encodeNodeAndGetHash(bn, getTrieContextWithCustomStorage(nil))
	en.ChildHash = childHash

	expected := &extensionNode{CollapsedEn: CollapsedEn{Key: []byte{2, 100, 111, 103}}, dirty: true}
	expected.child = en.child
	expected.ChildHash = en.ChildHash

	n, newChildPos, err := en.reduceNode(2, "mutex key", getTrieContextWithCustomStorage(nil))
	assert.Equal(t, expected, n)
	assert.Nil(t, err)
	assert.True(t, newChildPos)
}

func TestExtensionNode_reduceNodeCollapsedNode(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	rootHash, _ := tr.RootHash()
	collapsedTrie, _ := tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash), "")

	collapsedTrie.Delete([]byte("doe"))

	err := collapsedTrie.Commit(hashesCollector.NewDisabledHashesCollector())
	assert.Nil(t, err)
}

func TestExtensionNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
	assert.Equal(t, ErrEmptyExtensionNode, en.isEmptyOrNil())

	en = nil
	assert.Equal(t, ErrNilExtensionNode, en.isEmptyOrNil())
}

func TestExtensionNode_getChildren(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	trieCtx := getDefaultTrieContext()
	saveNodeToStorage(en, trieCtx)

	children, err := en.getChildren(trieCtx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(children))
}

func TestExtensionNode_getChildrenCollapsedEn(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = saveNodeToStorage(en, trieCtx)

	children, err := collapsedEn.getChildren(trieCtx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(children))
}

func TestExtensionNode_setDirty(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
	en.setDirty(true)

	assert.True(t, en.dirty)
}

func TestExtensionNode_loadChildren(t *testing.T) {
	t.Parallel()

	_, hasher := getTestMarshalizerAndHasher()
	tr, _ := newEmptyTrie()
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.Commit(hashesCollector.NewDisabledHashesCollector())
	nodes, _ := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := cache.NewLRUCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], hasher)
		nodesCacher.Put(n.hash, n, len(n.GetSerialized()))
	}

	en := getCollapsedEn(t, tr.GetRootNode())

	getNode := func(hash []byte) (node, error) {
		cacheData, _ := nodesCacher.Get(hash)
		return trieNode(cacheData, getTrieContextWithCustomStorage(nil))
	}
	_, _, err := en.loadChildren(getNode)
	assert.Nil(t, err)
	assert.NotNil(t, en.child)

	assert.Equal(t, 4, nodesCacher.Len())
}

func getCollapsedEn(t *testing.T, n node) *extensionNode {
	en, ok := n.(*extensionNode)
	assert.True(t, ok)
	en.child = nil

	return en
}

func TestExtensionNode_newExtensionNodeOkVals(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	child, _ := getBnAndCollapsedBn()
	en, err := newExtensionNode(key, child)

	assert.Nil(t, err)
	assert.Equal(t, key, en.Key)
	assert.Nil(t, en.ChildHash)
	assert.Equal(t, child, en.child)
	assert.True(t, en.dirty)
}

func TestExtensionNode_commitCollapsesTrieIfMaxTrieLevelInMemoryIsReached(t *testing.T) {
	t.Parallel()

	en, collapsedEn := getEnAndCollapsedEn()
	trieCtx := getDefaultTrieContext()
	manager := getTestGoroutinesManager()
	en.commitDirty(0, keyBuilder.NewKeyBuilder(), 1, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
	assert.Nil(t, manager.GetError())

	assert.Equal(t, collapsedEn.ChildHash, en.ChildHash)
	assert.Equal(t, collapsedEn.child, en.child)
	collapseEnHash, _ := encodeNodeAndGetHash(collapsedEn, trieCtx)
	enHash, _ := encodeNodeAndGetHash(en, trieCtx)
	assert.Equal(t, collapseEnHash, enHash)
}

func TestExtensionNode_printShouldNotPanicEvenIfNodeIsCollapsed(t *testing.T) {
	t.Parallel()

	enWriter := bytes.NewBuffer(make([]byte, 0))
	collapsedEnWriter := bytes.NewBuffer(make([]byte, 0))

	trieCtx := getDefaultTrieContext()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = saveNodeToStorage(en, trieCtx)
	_ = saveNodeToStorage(collapsedEn, trieCtx)

	en.print(enWriter, 0, trieCtx)
	collapsedEn.print(collapsedEnWriter, 0, trieCtx)

	assert.Equal(t, enWriter.Bytes(), collapsedEnWriter.Bytes())
}

func TestExtensionNode_getNextHashAndKey(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()
	proofVerified, nextHash, nextKey := collapsedEn.getNextHashAndKey([]byte("d"))

	assert.False(t, proofVerified)
	assert.Equal(t, collapsedEn.ChildHash, nextHash)
	assert.Equal(t, []byte{}, nextKey)
}

func TestExtensionNode_getNextHashAndKeyNilKey(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()
	proofVerified, nextHash, nextKey := collapsedEn.getNextHashAndKey(nil)

	assert.False(t, proofVerified)
	assert.Nil(t, nextHash)
	assert.Nil(t, nextKey)
}

func TestExtensionNode_getNextHashAndKeyNilNode(t *testing.T) {
	t.Parallel()

	var collapsedEn *extensionNode
	proofVerified, nextHash, nextKey := collapsedEn.getNextHashAndKey([]byte("d"))

	assert.False(t, proofVerified)
	assert.Nil(t, nextHash)
	assert.Nil(t, nextKey)
}

func TestExtensionNode_SizeInBytes(t *testing.T) {
	t.Parallel()

	var en *extensionNode
	assert.Equal(t, 0, en.sizeInBytes())

	collapsed := []byte("collapsed")
	key := []byte("key")
	en = &extensionNode{
		CollapsedEn: CollapsedEn{
			Key:          key,
			ChildHash:    collapsed,
			ChildVersion: 1,
		},
		child: nil,
		dirty: false,
	}
	assert.Equal(t, len(collapsed)+len(key)+1+pointerSizeInBytes+4, en.sizeInBytes())
}

func TestExtensionNode_commitSnapshotContextDone(t *testing.T) {
	t.Parallel()

	trieCtx := getDefaultTrieContext()
	en, _ := getEnAndCollapsedEn()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := en.commitSnapshot(trieCtx, nil, nil, ctx, statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, []byte("nodeBytes"), 0)
	assert.Equal(t, core.ErrContextClosing, err)
}

func TestExtensionNode_getValueReturnsEmptyByteSlice(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	assert.Equal(t, []byte{}, en.getValue())
}

func TestExtensionNode_commitSnapshotDbIsClosing(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	db.GetCalled = func(key []byte) ([]byte, error) {
		return nil, core.ErrContextClosing
	}
	args := GetDefaultTrieStorageManagerParameters()
	args.MainStorer = db
	trieStorage, _ := NewTrieStorageManager(args)
	trieCtx := getTrieContextWithCustomStorage(trieStorage)

	_, collapsedEn := getEnAndCollapsedEn()
	missingNodesChan := make(chan []byte, 10)
	nodeBytes, _ := collapsedEn.getEncodedNode(trieCtx)
	err := collapsedEn.commitSnapshot(trieCtx, nil, missingNodesChan, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, nodeBytes, 0)
	assert.True(t, core.IsClosingError(err))
	assert.Equal(t, 0, len(missingNodesChan))
}

func TestExtensionNode_getVersion(t *testing.T) {
	t.Parallel()

	t.Run("invalid node version", func(t *testing.T) {
		t.Parallel()

		en, _ := getEnAndCollapsedEn()
		en.ChildVersion = math.MaxUint8 + 1

		version, err := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Equal(t, ErrInvalidNodeVersion, err)
	})

	t.Run("NotSpecified version", func(t *testing.T) {
		t.Parallel()

		en, _ := getEnAndCollapsedEn()

		version, err := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)
		assert.Nil(t, err)
	})

	t.Run("AutoBalanceEnabled version", func(t *testing.T) {
		t.Parallel()

		en, _ := getEnAndCollapsedEn()
		en.ChildVersion = uint32(core.AutoBalanceEnabled)

		version, err := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
		assert.Nil(t, err)
	})
}

func TestExtensionNode_getNodeData(t *testing.T) {
	t.Parallel()

	t.Run("nil node", func(t *testing.T) {
		t.Parallel()

		var en *extensionNode
		nodeData, err := en.getNodeData(keyBuilder.NewDisabledKeyBuilder())
		assert.Nil(t, nodeData)
		assert.True(t, errors.Is(err, ErrNilExtensionNode))
	})
	t.Run("gets data from child", func(t *testing.T) {
		t.Parallel()

		_, en := getEnAndCollapsedEn()
		hashSize := 32
		keySize := 1

		kb := keyBuilder.NewKeyBuilder()
		nodeData, err := en.getNodeData(kb)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(nodeData))

		assert.Equal(t, uint(1), nodeData[0].GetKeyBuilder().Size())
		assert.Equal(t, en.ChildHash, nodeData[0].GetData())
		assert.Equal(t, uint64(hashSize+keySize), nodeData[0].Size())
		assert.False(t, nodeData[0].IsLeaf())
	})
}

func Test_getMinKeyMatchLen(t *testing.T) {
	t.Parallel()

	t.Run("same key", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{
				Key: []byte("dog"),
			},
		}
		keyMatchLen, index := getMinKeyMatchLen(newData, []byte("dog"))
		assert.Equal(t, 3, keyMatchLen)
		assert.Equal(t, 0, index)
	})
	t.Run("first key is min", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{
				Key: []byte("dog"),
			},
			{
				Key: []byte("doge"),
			},
		}
		keyMatchLen, index := getMinKeyMatchLen(newData, []byte("doe"))
		assert.Equal(t, 2, keyMatchLen)
		assert.Equal(t, 0, index)

	})
	t.Run("last key is min", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{
				Key: []byte("doge"),
			},
			{
				Key: []byte("dad"),
			},
		}
		keyMatchLen, index := getMinKeyMatchLen(newData, []byte("doe"))
		assert.Equal(t, 1, keyMatchLen)
		assert.Equal(t, 1, index)

	})
	t.Run("no match", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{
				Key: []byte("doge"),
			},
			{
				Key: []byte("dog"),
			},
		}
		keyMatchLen, index := getMinKeyMatchLen(newData, []byte("cat"))
		assert.Equal(t, 0, keyMatchLen)
		assert.Equal(t, 0, index)
	})
}

func Test_removeCommonPrefix(t *testing.T) {
	t.Parallel()

	t.Run("no common prefix", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{
				Key: []byte("doge"),
			},
			{
				Key: []byte("cat"),
			},
		}

		err := removeCommonPrefix(newData, 0)
		assert.Nil(t, err)
		assert.Equal(t, []byte("doge"), newData[0].Key)
		assert.Equal(t, []byte("cat"), newData[1].Key)
	})
	t.Run("remove prefix from all", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{
				Key: []byte("doge"),
			},
			{
				Key: []byte("doe"),
			},
		}

		err := removeCommonPrefix(newData, 2)
		assert.Nil(t, err)
		assert.Equal(t, []byte("ge"), newData[0].Key)
		assert.Equal(t, []byte("e"), newData[1].Key)

	})
	t.Run("one key is less than the prefix", func(t *testing.T) {
		t.Parallel()

		newData := []core.TrieData{
			{
				Key: []byte("doge"),
			},
			{
				Key: []byte("do"),
			},
		}

		err := removeCommonPrefix(newData, 3)
		assert.Equal(t, ErrValueTooShort, err)
	})
}

func getEn() (*extensionNode, []byte) {
	var children [nrOfChildren]node
	children[4] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{3, 4, 5}), "dog"))
	children[7] = newLeafNode(getTrieDataWithDefaultVersion(string([]byte{7, 8, 9}), "doe"))
	bn := newBranchNode()
	bn.children = children
	trieCtx := getTrieContextWithCustomStorage(nil)
	bn.ChildrenHashes[4], _ = encodeNodeAndGetHash(children[4], trieCtx)
	bn.ChildrenHashes[7], _ = encodeNodeAndGetHash(children[7], trieCtx)
	en, _ := newExtensionNode([]byte{1, 2}, bn)
	childHash, _ := encodeNodeAndGetHash(bn, trieCtx)
	en.ChildHash = childHash

	nodeHash, _ := encodeNodeAndGetHash(en, trieCtx)
	return en, nodeHash
}

func TestExtensionNode_insertInSameEn(t *testing.T) {
	t.Parallel()

	t.Run("insert same data", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 4, 3, 4, 5}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 7, 7, 8, 9}), "doe"),
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 0
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		assert.Nil(t, newNode)
		assert.False(t, en.dirty)
	})
	t.Run("insert different data", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 6, 7, 16}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 3, 4, 5}), "doe"),
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 1
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		en, ok := newNode.(*extensionNode)
		assert.True(t, ok)
		assert.True(t, en.dirty)
		bn, ok := en.child.(*branchNode)
		assert.True(t, ok)
		assert.False(t, bn.children[4].isDirty())
		assert.False(t, bn.children[7].isDirty())
		assert.Equal(t, []byte{4, 5}, bn.children[3].(*leafNode).Key)
		assert.Equal(t, []byte{7, 16}, bn.children[6].(*leafNode).Key)
	})
}

func TestExtensionNode_insertInNewBn(t *testing.T) {
	t.Parallel()

	t.Run("with a new en parent", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 3, 6, 7, 16}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{1, 3, 3, 4, 5}), "doe"),
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 1
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		en, ok := newNode.(*extensionNode)
		assert.True(t, ok)
		assert.True(t, en.dirty)
		assert.Equal(t, []byte{1}, en.Key)
		bn, ok := en.child.(*branchNode)
		assert.True(t, ok)
		assert.False(t, bn.children[2].isDirty())
		assert.True(t, bn.children[3].isDirty())
		_, ok = bn.children[2].(*branchNode)
		assert.True(t, ok)
		_, ok = bn.children[3].(*branchNode)
		assert.True(t, ok)
	})
	t.Run("branch at the beginning of the en", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{2, 3, 6, 7, 16}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{3, 3, 3, 4, 5}), "doe"),
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 1
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		bn, ok := newNode.(*branchNode)
		assert.True(t, ok)
		assert.True(t, bn.dirty)
		assert.Equal(t, []byte{3, 6, 7, 16}, bn.children[2].(*leafNode).Key)
		assert.Equal(t, []byte{3, 3, 4, 5}, bn.children[3].(*leafNode).Key)
		assert.Equal(t, []byte{2}, bn.children[1].(*extensionNode).Key)
	})
	t.Run("branch and insert existing value", func(t *testing.T) {
		t.Parallel()

		originalEn, nodeHash := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		originalEn.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 3, 6, 7, 16}), "dog"),   // new value
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 4, 3, 4, 5}), "dog"), // this is already in the trie
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		newNode := originalEn.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 1
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		en, ok := newNode.(*extensionNode)
		assert.True(t, ok)
		assert.True(t, en.dirty)

		bn, ok := en.child.(*branchNode)
		assert.True(t, ok)
		assert.True(t, bn.dirty)

		enHash, _ := encodeNodeAndGetHash(en, trieCtx)
		assert.False(t, bn.children[2].(*branchNode).dirty)
		assert.Equal(t, originalEn.child, bn.children[2])
		assert.NotEqual(t, nodeHash, enHash)
	})
}

func TestExtensionNode_deleteBatch(t *testing.T) {
	t.Parallel()

	t.Run("delete invalid node", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{2, 3, 6, 7, 16}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{3, 3, 3, 4, 5}), "doe"),
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := en.delete(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.False(t, dirty)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 0
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		assert.Equal(t, en, newNode)
	})
	t.Run("reduce to leaf after delete", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 4, 3, 4, 5}), "dog"),
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := en.delete(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.True(t, dirty)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 3
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		ln, ok := newNode.(*leafNode)
		assert.True(t, ok)
		assert.Equal(t, []byte{1, 2, 7, 7, 8, 9}, ln.Key)
		assert.True(t, ln.dirty)
	})
	t.Run("reduce to extension node after delete", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 4, 4, 5, 6}), "dog"),
		}
		trieCtx := getDefaultTrieContext()

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		newEn := en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, common.NewModifiedHashesSlice(initialModifiedHashesCapacity), trieCtx)
		newEn.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		dataForRemoval := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 7, 7, 8, 9}), "dog"),
		}

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := newEn.delete(keyBuilder.NewKeyBuilder(), dataForRemoval, goRoutinesManager, modifiedHashes, trieCtx)
		assert.True(t, dirty)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 2
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		en, ok := newNode.(*extensionNode)
		assert.True(t, ok)
		assert.Equal(t, []byte{1, 2, 4}, en.Key)
		assert.True(t, en.dirty)
	})
	t.Run("delete all children", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		manager := getTestGoroutinesManager()
		trieCtx := getDefaultTrieContext()
		assert.Nil(t, manager.GetError())
		en.commitDirty(0, keyBuilder.NewKeyBuilder(), 5, manager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
		assert.Nil(t, manager.GetError())

		data := []core.TrieData{
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 4, 3, 4, 5}), "dog"),
			getTrieDataWithDefaultVersion(string([]byte{1, 2, 7, 7, 8, 9}), "doe"),
		}

		th, _ := throttler.NewNumGoRoutinesThrottler(5)
		goRoutinesManager, err := NewGoroutinesManager(th, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.Nil(t, err)

		modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
		dirty, newNode := en.delete(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
		assert.True(t, dirty)
		assert.Nil(t, goRoutinesManager.GetError())
		expectedNumTrieNodesChanged := 3
		assert.Equal(t, expectedNumTrieNodesChanged, len(modifiedHashes.Get()))
		assert.Nil(t, newNode)
	})
}

func TestExtensionNode_concurrency(t *testing.T) {
	t.Parallel()

	t.Run("insert at same key and get", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		trieCtx := getDefaultTrieContext()
		_ = saveNodeToStorage(en, trieCtx)

		newVal := []core.TrieData{
			{[]byte{1, 2, 3}, []byte("value"), core.NotSpecified},
		}
		existingChildKey := []byte{1, 2, 4, 3, 4, 5}

		startExecution := atomic.Flag{}
		wg := sync2.WaitGroup{}
		wg.Add(2)
		go func() {
			for !startExecution.IsSet() {
				time.Sleep(time.Millisecond)
			}
			val, depth, err := en.tryGet(newKeyData(existingChildKey), 0, trieCtx)
			assert.Nil(t, err)
			assert.Equal(t, uint32(2), depth)
			assert.Equal(t, []byte("dog"), val)

			wg.Done()
		}()
		go func() {
			startExecution.SetValue(true)
			goRoutinesManager := getTestGoroutinesManager()
			modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)

			newNode := en.insert(keyBuilder.NewKeyBuilder(), newVal, goRoutinesManager, modifiedHashes, trieCtx)
			_, ok := newNode.(*extensionNode)
			assert.True(t, ok)
			assert.Nil(t, goRoutinesManager.GetError())

			wg.Done()
		}()

		wg.Wait()
	})
	t.Run("insert different key and get", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		trieCtx := getDefaultTrieContext()
		_ = saveNodeToStorage(en, trieCtx)

		newVal := []core.TrieData{
			{[]byte{4, 5, 6}, []byte("value"), core.NotSpecified},
		}
		existingChildKey := []byte{1, 2, 4, 3, 4, 5}

		startExecution := atomic.Flag{}
		wg := sync2.WaitGroup{}
		wg.Add(2)
		go func() {
			for !startExecution.IsSet() {
				time.Sleep(time.Millisecond)
			}
			val, depth, err := en.tryGet(newKeyData(existingChildKey), 0, trieCtx)
			assert.Nil(t, err)
			assert.Equal(t, uint32(2), depth)
			assert.Equal(t, []byte("dog"), val)

			wg.Done()
		}()
		go func() {
			startExecution.SetValue(true)
			goRoutinesManager := getTestGoroutinesManager()
			modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)

			newNode := en.insert(keyBuilder.NewKeyBuilder(), newVal, goRoutinesManager, modifiedHashes, trieCtx)
			_, ok := newNode.(*branchNode)
			assert.True(t, ok)
			assert.Nil(t, goRoutinesManager.GetError())

			wg.Done()
		}()

		wg.Wait()
	})
	t.Run("delete and get different keys", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		trieCtx := getDefaultTrieContext()
		_ = saveNodeToStorage(en, trieCtx)

		newVal := []core.TrieData{
			{[]byte{1, 2, 7, 7, 8, 9}, nil, core.NotSpecified},
		}
		existingChildKey := []byte{1, 2, 4, 3, 4, 5}

		startExecution := atomic.Flag{}
		wg := sync2.WaitGroup{}
		wg.Add(2)
		go func() {
			for !startExecution.IsSet() {
				time.Sleep(time.Millisecond)
			}
			_, depth, err := en.tryGet(newKeyData(existingChildKey), 0, trieCtx)
			assert.Nil(t, err)
			assert.Equal(t, uint32(2), depth)

			wg.Done()
		}()
		go func() {
			startExecution.SetValue(true)
			goRoutinesManager := getTestGoroutinesManager()
			modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)

			dirty, newNode := en.delete(keyBuilder.NewKeyBuilder(), newVal, goRoutinesManager, modifiedHashes, trieCtx)
			assert.True(t, dirty)
			_, ok := newNode.(*leafNode)
			assert.True(t, ok)
			assert.Nil(t, goRoutinesManager.GetError())

			wg.Done()
		}()

		wg.Wait()
	})
	t.Run("delete and get same key", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		trieCtx := getDefaultTrieContext()
		_ = saveNodeToStorage(en, trieCtx)

		existingChildKey := []byte{1, 2, 4, 3, 4, 5}
		newVal := []core.TrieData{
			{existingChildKey, nil, core.NotSpecified},
		}

		startExecution := atomic.Flag{}
		wg := sync2.WaitGroup{}
		wg.Add(2)
		go func() {
			for !startExecution.IsSet() {
				time.Sleep(time.Millisecond)
			}
			_, _, err := en.tryGet(newKeyData(existingChildKey), 0, trieCtx)
			assert.Nil(t, err)

			wg.Done()
		}()
		go func() {
			startExecution.SetValue(true)
			goRoutinesManager := getTestGoroutinesManager()
			modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)

			dirty, newNode := en.delete(keyBuilder.NewKeyBuilder(), newVal, goRoutinesManager, modifiedHashes, trieCtx)
			assert.True(t, dirty)
			_, ok := newNode.(*leafNode)
			assert.True(t, ok)
			assert.Nil(t, goRoutinesManager.GetError())

			wg.Done()
		}()

		wg.Wait()
	})
	t.Run("commit and get", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		trieCtx := getDefaultTrieContext()
		_ = saveNodeToStorage(en, trieCtx)
		en, _ = getEn()

		existingChildKey := []byte{1, 2, 4, 3, 4, 5}

		startExecution := atomic.Flag{}
		wg := sync2.WaitGroup{}
		wg.Add(2)
		go func() {
			for !startExecution.IsSet() {
				time.Sleep(time.Millisecond)
			}
			val, depth, err := en.tryGet(newKeyData(existingChildKey), 0, trieCtx)
			assert.Nil(t, err)
			assert.Equal(t, uint32(2), depth)
			assert.Equal(t, []byte("dog"), val)

			wg.Done()
		}()

		assert.True(t, en.dirty)
		go func() {
			startExecution.SetValue(true)
			goRoutinesManager := getTestGoroutinesManager()
			en.commitDirty(0, keyBuilder.NewKeyBuilder(), 1, goRoutinesManager, hashesCollector.NewDisabledHashesCollector(), trieCtx)
			assert.False(t, en.dirty)
			assert.Nil(t, goRoutinesManager.GetError())

			wg.Done()
		}()

		wg.Wait()
	})
	t.Run("get and insert, same node changes hash", func(t *testing.T) {
		t.Parallel()

		en, _ := getEn()
		existingChildKey := []byte{1, 2, 4, 3, 4, 5}
		numGet := 1000
		numInsert := 1000
		wg := sync2.WaitGroup{}
		wg.Add(2)
		trieCtx := getDefaultTrieContext()
		go func() {
			for i := 0; i < numGet; i++ {
				_, _, err := en.tryGet(newKeyData(existingChildKey), 0, trieCtx)
				assert.Nil(t, err)
			}
			wg.Done()
		}()
		go func() {
			goRoutinesManager := getTestGoroutinesManager()
			modifiedHashes := common.NewModifiedHashesSlice(initialModifiedHashesCapacity)
			for i := 0; i < numInsert; i++ {
				data := []core.TrieData{getTrieDataWithDefaultVersion(string(existingChildKey), strconv.Itoa(i))}
				_ = en.insert(keyBuilder.NewKeyBuilder(), data, goRoutinesManager, modifiedHashes, trieCtx)
				assert.Nil(t, goRoutinesManager.GetError())
			}

			wg.Done()
		}()

		wg.Wait()
	})
}
