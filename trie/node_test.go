package trie

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	dataMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_hashChildrenAndNodeBranchNode(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	expectedNodeHash, _ := encodeNodeAndGetHash(collapsedBn)

	hash, err := hashChildrenAndNode(bn)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodeHash, hash)
}

func TestNode_hashChildrenAndNodeExtensionNode(t *testing.T) {
	t.Parallel()

	en, collapsedEn := getEnAndCollapsedEn()
	expectedNodeHash, _ := encodeNodeAndGetHash(collapsedEn)

	hash, err := hashChildrenAndNode(en)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodeHash, hash)
}

func TestNode_hashChildrenAndNodeLeafNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	expectedNodeHash, _ := encodeNodeAndGetHash(ln)

	hash, err := hashChildrenAndNode(ln)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodeHash, hash)
}

func TestNode_encodeNodeAndGetHashBranchNode(t *testing.T) {
	t.Parallel()

	bn, _ := newBranchNode(getTestMarshalizerAndHasher())
	encChildren := make([][]byte, nrOfChildren)
	encChildren[1] = []byte("dog")
	encChildren[10] = []byte("doge")
	bn.EncodedChildren = encChildren

	encNode, _ := bn.marsh.Marshal(bn)
	encNode = append(encNode, branch)
	expextedHash := bn.hasher.Compute(string(encNode))

	hash, err := encodeNodeAndGetHash(bn)
	assert.Nil(t, err)
	assert.Equal(t, expextedHash, hash)
}

func TestNode_encodeNodeAndGetHashExtensionNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	en := &extensionNode{
		CollapsedEn: CollapsedEn{
			Key:          []byte{2},
			EncodedChild: []byte("doge"),
		},
		baseNode: &baseNode{

			marsh:  marsh,
			hasher: hasher,
		},
	}

	encNode, _ := marsh.Marshal(en)
	encNode = append(encNode, extension)
	expextedHash := hasher.Compute(string(encNode))

	hash, err := encodeNodeAndGetHash(en)
	assert.Nil(t, err)
	assert.Equal(t, expextedHash, hash)
}

func TestNode_encodeNodeAndGetHashLeafNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	ln, _ := newLeafNode(getTrieDataWithDefaultVersion("dog", "dog"), marsh, hasher)

	encNode, _ := marsh.Marshal(ln)
	encNode = append(encNode, leaf)
	expextedHash := hasher.Compute(string(encNode))

	hash, err := encodeNodeAndGetHash(ln)
	assert.Nil(t, err)
	assert.Equal(t, expextedHash, hash)
}

func TestNode_encodeNodeAndCommitToDBBranchNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	encNode, _ := collapsedBn.marsh.Marshal(collapsedBn)
	encNode = append(encNode, branch)
	nodeHash := collapsedBn.hasher.Compute(string(encNode))

	_, err := encodeNodeAndCommitToDB(collapsedBn, db)
	assert.Nil(t, err)

	val, _ := db.Get(nodeHash)
	assert.Equal(t, encNode, val)
}

func TestNode_encodeNodeAndCommitToDBExtensionNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	_, collapsedEn := getEnAndCollapsedEn()
	encNode, _ := collapsedEn.marsh.Marshal(collapsedEn)
	encNode = append(encNode, extension)
	nodeHash := collapsedEn.hasher.Compute(string(encNode))

	_, err := encodeNodeAndCommitToDB(collapsedEn, db)
	assert.Nil(t, err)

	val, _ := db.Get(nodeHash)
	assert.Equal(t, encNode, val)
}

func TestNode_encodeNodeAndCommitToDBLeafNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, leaf)
	nodeHash := ln.hasher.Compute(string(encNode))

	_, err := encodeNodeAndCommitToDB(ln, db)
	assert.Nil(t, err)

	val, _ := db.Get(nodeHash)
	assert.Equal(t, encNode, val)
}

func TestNode_getNodeFromDBAndDecodeBranchNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	_ = bn.commitDirty(0, 5, db, db)

	encNode, _ := bn.marsh.Marshal(collapsedBn)
	encNode = append(encNode, branch)
	nodeHash := bn.hasher.Compute(string(encNode))

	nodeInstance, err := getNodeFromDBAndDecode(nodeHash, db, bn.marsh, bn.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedBn)
	h2, _ := encodeNodeAndGetHash(nodeInstance)
	assert.Equal(t, h1, h2)
}

func TestNode_getNodeFromDBAndDecodeExtensionNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commitDirty(0, 5, db, db)

	encNode, _ := en.marsh.Marshal(collapsedEn)
	encNode = append(encNode, extension)
	nodeHash := en.hasher.Compute(string(encNode))

	nodeInstance, err := getNodeFromDBAndDecode(nodeHash, db, en.marsh, en.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedEn)
	h2, _ := encodeNodeAndGetHash(nodeInstance)
	assert.Equal(t, h1, h2)
}

func TestNode_getNodeFromDBAndDecodeLeafNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commitDirty(0, 5, db, db)

	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, leaf)
	nodeHash := ln.hasher.Compute(string(encNode))

	nodeInstance, err := getNodeFromDBAndDecode(nodeHash, db, ln.marsh, ln.hasher)
	assert.Nil(t, err)

	ln = getLn(ln.marsh, ln.hasher)
	ln.dirty = false
	assert.Equal(t, ln, nodeInstance)
}

func TestNode_resolveIfCollapsedBranchNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	_ = bn.commitDirty(0, 5, db, db)

	err := resolveIfCollapsed(collapsedBn, childPos, db)
	assert.Nil(t, err)
	assert.False(t, collapsedBn.isCollapsed())
}

func TestNode_resolveIfCollapsedExtensionNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commitDirty(0, 5, db, db)

	err := resolveIfCollapsed(collapsedEn, 0, db)
	assert.Nil(t, err)
	assert.False(t, collapsedEn.isCollapsed())
}

func TestNode_resolveIfCollapsedLeafNode(t *testing.T) {
	t.Parallel()

	db := testscommon.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commitDirty(0, 5, db, db)

	err := resolveIfCollapsed(ln, 0, db)
	assert.Nil(t, err)
	assert.False(t, ln.isCollapsed())
}

func TestNode_resolveIfCollapsedNilNode(t *testing.T) {
	t.Parallel()

	var nodeInstance *extensionNode

	err := resolveIfCollapsed(nodeInstance, 0, nil)
	assert.Equal(t, ErrNilExtensionNode, err)
}

func TestNode_concat(t *testing.T) {
	t.Parallel()

	a := []byte{1, 2, 3}
	b := byte(4)
	ab := []byte{1, 2, 3, 4}
	assert.Equal(t, ab, concat(a, b))
}

func TestNode_hasValidHash(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	ok, err := hasValidHash(bn)
	assert.Nil(t, err)
	assert.False(t, ok)

	_ = bn.setHash()
	bn.dirty = false

	ok, err = hasValidHash(bn)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestNode_hasValidHashNilNode(t *testing.T) {
	t.Parallel()

	var nodeInstance *branchNode
	ok, err := hasValidHash(nodeInstance)
	assert.Equal(t, ErrNilBranchNode, err)
	assert.False(t, ok)
}

func TestNode_decodeNodeBranchNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	encNode, _ := collapsedBn.marsh.Marshal(collapsedBn)
	encNode = append(encNode, branch)

	nodeInstance, err := decodeNode(encNode, collapsedBn.marsh, collapsedBn.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedBn)
	h2, _ := encodeNodeAndGetHash(nodeInstance)
	assert.Equal(t, h1, h2)
}

func TestNode_decodeNodeExtensionNode(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()
	encNode, _ := collapsedEn.marsh.Marshal(collapsedEn)
	encNode = append(encNode, extension)

	nodeInstance, err := decodeNode(encNode, collapsedEn.marsh, collapsedEn.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedEn)
	h2, _ := encodeNodeAndGetHash(nodeInstance)
	assert.Equal(t, h1, h2)
}

func TestNode_decodeNodeLeafNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, leaf)

	nodeInstance, err := decodeNode(encNode, ln.marsh, ln.hasher)
	assert.Nil(t, err)
	ln.dirty = false

	h1, _ := encodeNodeAndGetHash(ln)
	h2, _ := encodeNodeAndGetHash(nodeInstance)
	assert.Equal(t, h1, h2)
}

func TestNode_decodeNodeInvalidNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	invalidNode := byte(6)

	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, invalidNode)

	nodeInstance, err := decodeNode(encNode, ln.marsh, ln.hasher)
	assert.Nil(t, nodeInstance)
	assert.Equal(t, ErrInvalidNode, err)
}

func TestNode_decodeNodeInvalidEncoding(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	var encNode []byte

	nodeInstance, err := decodeNode(encNode, marsh, hasher)
	assert.Nil(t, nodeInstance)
	assert.Equal(t, ErrInvalidEncoding, err)
}

func TestNode_getEmptyNodeOfTypeBranchNode(t *testing.T) {
	t.Parallel()

	bn, err := getEmptyNodeOfType(branch)
	assert.Nil(t, err)
	assert.IsType(t, &branchNode{}, bn)
}

func TestNode_getEmptyNodeOfTypeExtensionNode(t *testing.T) {
	t.Parallel()

	en, err := getEmptyNodeOfType(extension)
	assert.Nil(t, err)
	assert.IsType(t, &extensionNode{}, en)
}

func TestNode_getEmptyNodeOfTypeLeafNode(t *testing.T) {
	t.Parallel()

	ln, err := getEmptyNodeOfType(leaf)
	assert.Nil(t, err)
	assert.IsType(t, &leafNode{}, ln)
}

func TestNode_getEmptyNodeOfTypeWrongNode(t *testing.T) {
	t.Parallel()

	n, err := getEmptyNodeOfType(6)
	assert.Equal(t, ErrInvalidNode, err)
	assert.Nil(t, n)
}

func TestNode_childPosOutOfRange(t *testing.T) {
	t.Parallel()

	assert.True(t, childPosOutOfRange(17))
	assert.False(t, childPosOutOfRange(5))
}

func TestKeyBytesToHex(t *testing.T) {
	t.Parallel()

	reversedHexDoeKey := []byte{5, 6, 15, 6, 4, 6, 16}
	reversedHexDogKey := []byte{7, 6, 15, 6, 4, 6, 16}

	var test = []struct {
		key, hex []byte
	}{
		{[]byte("doe"), reversedHexDoeKey},
		{[]byte("dog"), reversedHexDogKey},
	}

	for i := range test {
		assert.Equal(t, test[i].hex, keyBytesToHex(test[i].key))
	}
}

func TestPrefixLen(t *testing.T) {
	t.Parallel()

	var test = []struct {
		a, b   []byte
		length int
	}{
		{[]byte("doe"), []byte("dog"), 2},
		{[]byte("dog"), []byte("dogglesworth"), 3},
		{[]byte("mouse"), []byte("mouse"), 5},
		{[]byte("caterpillar"), []byte("cats"), 3},
		{[]byte("caterpillar"), []byte(""), 0},
		{[]byte(""), []byte("caterpillar"), 0},
		{[]byte("a"), []byte("caterpillar"), 0},
	}

	for i := range test {
		assert.Equal(t, test[i].length, prefixLen(test[i].a, test[i].b))
	}
}

func TestGetOldHashesIfNodeIsCollapsed(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()

	root, _ := tr.root.(*branchNode)
	for i := 0; i < nrOfChildren; i++ {
		root.children[i] = nil
	}
	tr.root = root

	_ = tr.Update([]byte("dog"), []byte("value of dog"))

	assert.Equal(t, 4, len(tr.oldHashes))
}

func TestClearOldHashesAndOldRootOnCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	root, _ := tr.RootHash()

	_ = tr.Update([]byte("dog"), []byte("value of dog"))

	assert.Equal(t, 4, len(tr.oldHashes))
	assert.Equal(t, root, tr.oldRoot)

	_ = tr.Commit()

	assert.Equal(t, 0, len(tr.oldHashes))
	assert.Equal(t, 0, len(tr.oldRoot))
}

func TestTrieGetObsoleteHashes(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()

	_ = tr.Update([]byte("doeee"), []byte("value of doeee"))

	assert.NotEqual(t, 0, len(tr.oldHashes))
	assert.NotEqual(t, 0, len(tr.oldRoot))

	expectedHashes := tr.oldHashes
	hashes := tr.GetObsoleteHashes()
	assert.Equal(t, expectedHashes, hashes)
}

func TestNode_getDirtyHashes(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	_ = tr.root.setRootHash()
	hashes := make(map[string]struct{})
	err := tr.root.getDirtyHashes(hashes)

	assert.Nil(t, err)
	assert.NotNil(t, hashes)
	assert.Equal(t, 6, len(hashes))
}

func TestPatriciaMerkleTrie_GetAllLeavesCollapsedTrie(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()

	root, _ := tr.root.(*branchNode)
	for i := 0; i < nrOfChildren; i++ {
		root.children[i] = nil
	}
	tr.root = root

	leavesChannel := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), tr.root.getHash(), keyBuilder.NewKeyBuilder(), parsers.NewMainTrieLeafParser())
	assert.Nil(t, err)
	leaves := make(map[string][]byte)

	for l := range leavesChannel.LeavesChan {
		leaves[string(l.Key())] = l.Value()
	}

	err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
	assert.Nil(t, err)

	assert.Equal(t, 3, len(leaves))
	assert.Equal(t, []byte("reindeer"), leaves["doe"])
	assert.Equal(t, []byte("puppy"), leaves["dog"])
	assert.Equal(t, []byte("cat"), leaves["ddog"])
}

func TestPatriciaMerkleTrie_oldRootAndoldHashesAreResetAfterEveryCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()

	_ = tr.Update([]byte("doe"), []byte("deer"))
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	err := tr.Commit()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tr.oldHashes))
	assert.Equal(t, 0, len(tr.oldRoot))
}

func TestNode_NodeExtension(t *testing.T) {
	n := &branchNode{
		baseNode: &baseNode{
			hasher: &dataMock.HasherStub{
				ComputeCalled: func(s string) []byte {
					return []byte{0, 0, 0, 0}
				},
			},
		},
	}
	assert.True(t, shouldTestNode(n, make([]byte, 0)))

	n = &branchNode{
		baseNode: &baseNode{
			hasher: &dataMock.HasherStub{
				ComputeCalled: func(s string) []byte {
					return []byte{0, 0, 0, 1}
				},
			},
		},
	}
	assert.False(t, shouldTestNode(n, make([]byte, 0)))
}

func TestSnapshotGetTestPoint(t *testing.T) {
	t.Parallel()

	err := snapshotGetTestPoint([]byte("key"), 1)
	assert.True(t, strings.Contains(err.Error(), "snapshot get error"))
}

func TestShouldStopIfContextDoneBlockingIfBusy(t *testing.T) {
	t.Parallel()

	t.Run("context done", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFunc := context.WithCancel(context.Background())
		assert.False(t, shouldStopIfContextDoneBlockingIfBusy(ctx, &testscommon.ProcessStatusHandlerStub{}))
		cancelFunc()
		assert.True(t, shouldStopIfContextDoneBlockingIfBusy(ctx, &testscommon.ProcessStatusHandlerStub{}))
	})
	t.Run("wait until idle", func(t *testing.T) {
		t.Parallel()

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		flag := atomic.Flag{} // default is false so the idleProvider will say it's not idle
		idleProvider := &testscommon.ProcessStatusHandlerStub{
			IsIdleCalled: func() bool {
				return flag.IsSet()
			},
		}

		chResult := make(chan bool, 1)
		go func() {
			chResult <- shouldStopIfContextDoneBlockingIfBusy(ctx, idleProvider)
		}()

		select {
		case <-chResult:
			// we should have not received any results now since the idle provider states it is not idle
			assert.Fail(t, "should have not stop now")
		case <-time.After(time.Second):
		}

		flag.SetValue(true)

		select {
		case result := <-chResult:
			assert.False(t, result)
		case <-time.After(time.Second):
			assert.Fail(t, "timeout while waiting for the shouldStopIfContextDone call to write the result")
		}
	})
}

func TestTreatLogError(t *testing.T) {
	t.Parallel()

	t.Run("logger instance is not in Trace mode, should not call", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		err := errors.New("trie was not found")
		logInstance := &testscommon.LoggerStub{
			GetLevelCalled: func() logger.LogLevel {
				return logger.LogDebug
			},
			TraceCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called Log")
			},
		}

		treatLogError(logInstance, err, key)
		treatLogError(log, err, key) //display only
	})
	t.Run("logger instance is in Trace mode, should call", func(t *testing.T) {
		t.Parallel()

		key := []byte("key")
		wasCalled := false
		err := errors.New("error")
		logInstance := &testscommon.LoggerStub{
			GetLevelCalled: func() logger.LogLevel {
				return logger.LogTrace
			},
			TraceCalled: func(message string, args ...interface{}) {
				wasCalled = true
				require.Equal(t, core.GetNodeFromDBErrorString, message)
				require.Equal(t, 6, len(args))
				expectedFirst5Args := []interface{}{"error", err, "key", key, "stack trace"}
				require.Equal(t, expectedFirst5Args, args[:5])
			},
		}

		treatLogError(logInstance, err, key)
		assert.True(t, wasCalled)
	})
}

func TestNodesVersion_insertInLn(t *testing.T) {
	t.Parallel()

	t.Run("insert in same leaf - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.NotSpecified)
		ln, ok := tr.root.(*leafNode)
		assert.True(t, ok)
		version, _ := ln.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aab"), core.AutoBalanceEnabled)
		version, _ = ln.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("insert in leaf - create new branch node", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		tr, _ = newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		bn, ok = tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("insert in leaf - create new extension", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		tr, _ = newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok = tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

}

func TestNodesVersion_insertInEn(t *testing.T) {
	t.Parallel()

	t.Run("insert in same extension node - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		en, ok = tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.NotSpecified)
		en, ok = tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.NotSpecified, version)
	})

	t.Run("insert in extension node - create new branch - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("qqq"), []byte("qqq"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("zzz"), []byte("zzz"), core.NotSpecified)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)
	})

	t.Run("insert in extension node - create new branch - do not change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("qqq"), []byte("qqq"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("zzz"), []byte("zzz"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("insert in extension node - create new branch with following extension node - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("zzz"), []byte("zzz"), core.NotSpecified)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)
	})

	t.Run("insert in extension node - create new branch with following extension node - do not change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("zzz"), []byte("zzz"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("insert in extension node - create new extension and branch - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("bba"), []byte("bba"), core.NotSpecified)
		en, ok = tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.NotSpecified, version)
	})

	t.Run("insert in extension node - create new extension and branch - do not change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("bba"), []byte("bba"), core.AutoBalanceEnabled)
		en, ok = tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})
}

func TestNodesVersion_insertInBn(t *testing.T) {
	t.Parallel()

	t.Run("insert in branch node on nil child - same version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("ccc"), []byte("ccc"), core.AutoBalanceEnabled)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("insert in branch node on nil child - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("ccc"), []byte("ccc"), core.NotSpecified)
		version, _ = bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)
	})

	t.Run("insert in branch node on existing child - same version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aab"), core.AutoBalanceEnabled)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("insert in branch node on existing child - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aab"), core.AutoBalanceEnabled)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})
}

func TestNodesVersion_deleteFromEn(t *testing.T) {
	t.Parallel()

	t.Run("new child is leaf node - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("aaa"))
		ln, ok := tr.root.(*leafNode)
		assert.True(t, ok)
		version, _ = ln.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("new child is leaf node - same version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("baa"))
		ln, ok := tr.root.(*leafNode)
		assert.True(t, ok)
		version, _ = ln.getVersion()
		assert.Equal(t, core.NotSpecified, version)
	})

	t.Run("new child is extension node - same version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("zza"), []byte("zza"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.Delete([]byte("zza"))
		en, ok = tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("new child is extension node - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("zza"), []byte("zza"), core.NotSpecified)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("zza"))
		en, ok = tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("new child is branch node - same version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bba"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.Delete([]byte("aaa"))
		bn, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("new child is branch node - change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("baa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("baa"), []byte("baa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bba"), []byte("baa"), core.AutoBalanceEnabled)
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ := en.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("aaa"))
		bn, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})
}

func TestNodesVersion_deleteFromBn(t *testing.T) {
	t.Parallel()

	t.Run("delete leaf - branch does not reduce - bn should not change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("ccc"), []byte("ccc"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)

		_ = tr.Delete([]byte("aaa"))
		bn, ok = tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("delete leaf - branch does not reduce - bn should change version", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("ccc"), []byte("ccc"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("aaa"))
		bn, ok = tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ = bn.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("branch with branch child is reduced", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("qqq"), []byte("bbb"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("zzz"), []byte("ccc"), core.NotSpecified)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("zzz"))
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("branch with extension child is reduced", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("bba"), []byte("bbb"), core.AutoBalanceEnabled)
		_ = tr.UpdateWithVersion([]byte("zzz"), []byte("ccc"), core.NotSpecified)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("zzz"))
		en, ok := tr.root.(*extensionNode)
		assert.True(t, ok)
		version, _ = en.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})

	t.Run("branch with leaf child is reduced", func(t *testing.T) {
		t.Parallel()

		tr, _ := newEmptyTrie()
		_ = tr.UpdateWithVersion([]byte("aaa"), []byte("aaa"), core.NotSpecified)
		_ = tr.UpdateWithVersion([]byte("bbb"), []byte("bbb"), core.AutoBalanceEnabled)
		bn, ok := tr.root.(*branchNode)
		assert.True(t, ok)
		version, _ := bn.getVersion()
		assert.Equal(t, core.NotSpecified, version)

		_ = tr.Delete([]byte("aaa"))
		ln, ok := tr.root.(*leafNode)
		assert.True(t, ok)
		version, _ = ln.getVersion()
		assert.Equal(t, core.AutoBalanceEnabled, version)
	})
}

func Benchmark_ShouldStopIfContextDoneBlockingIfBusy(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = shouldStopIfContextDoneBlockingIfBusy(ctx, &testscommon.ProcessStatusHandlerStub{})
	}
}
