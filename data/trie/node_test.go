package trie

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
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
	ln, _ := newLeafNode([]byte("dog"), []byte("dog"), marsh, hasher)

	encNode, _ := marsh.Marshal(ln)
	encNode = append(encNode, leaf)
	expextedHash := hasher.Compute(string(encNode))

	hash, err := encodeNodeAndGetHash(ln)
	assert.Nil(t, err)
	assert.Equal(t, expextedHash, hash)
}

func TestNode_encodeNodeAndCommitToDBBranchNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
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

	db := mock.NewMemDbMock()
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

	db := mock.NewMemDbMock()
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

	db := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	_ = bn.commitDirty(0, 5, db, db)

	encNode, _ := bn.marsh.Marshal(collapsedBn)
	encNode = append(encNode, branch)
	nodeHash := bn.hasher.Compute(string(encNode))

	node, err := getNodeFromDBAndDecode(nodeHash, db, bn.marsh, bn.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedBn)
	h2, _ := encodeNodeAndGetHash(node)
	assert.Equal(t, h1, h2)
}

func TestNode_getNodeFromDBAndDecodeExtensionNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commitDirty(0, 5, db, db)

	encNode, _ := en.marsh.Marshal(collapsedEn)
	encNode = append(encNode, extension)
	nodeHash := en.hasher.Compute(string(encNode))

	node, err := getNodeFromDBAndDecode(nodeHash, db, en.marsh, en.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedEn)
	h2, _ := encodeNodeAndGetHash(node)
	assert.Equal(t, h1, h2)
}

func TestNode_getNodeFromDBAndDecodeLeafNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commitDirty(0, 5, db, db)

	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, leaf)
	nodeHash := ln.hasher.Compute(string(encNode))

	node, err := getNodeFromDBAndDecode(nodeHash, db, ln.marsh, ln.hasher)
	assert.Nil(t, err)

	ln = getLn(ln.marsh, ln.hasher)
	ln.dirty = false
	assert.Equal(t, ln, node)
}

func TestNode_resolveIfCollapsedBranchNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	childPos := byte(2)
	_ = bn.commitDirty(0, 5, db, db)

	err := resolveIfCollapsed(collapsedBn, childPos, db)
	assert.Nil(t, err)
	assert.False(t, collapsedBn.isCollapsed())
}

func TestNode_resolveIfCollapsedExtensionNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commitDirty(0, 5, db, db)

	err := resolveIfCollapsed(collapsedEn, 0, db)
	assert.Nil(t, err)
	assert.False(t, collapsedEn.isCollapsed())
}

func TestNode_resolveIfCollapsedLeafNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshalizerAndHasher())
	_ = ln.commitDirty(0, 5, db, db)

	err := resolveIfCollapsed(ln, 0, db)
	assert.Nil(t, err)
	assert.False(t, ln.isCollapsed())
}

func TestNode_resolveIfCollapsedNilNode(t *testing.T) {
	t.Parallel()

	var node *extensionNode

	err := resolveIfCollapsed(node, 0, nil)
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

	var node *branchNode
	ok, err := hasValidHash(node)
	assert.Equal(t, ErrNilBranchNode, err)
	assert.False(t, ok)
}

func TestNode_decodeNodeBranchNode(t *testing.T) {
	t.Parallel()

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	encNode, _ := collapsedBn.marsh.Marshal(collapsedBn)
	encNode = append(encNode, branch)

	node, err := decodeNode(encNode, collapsedBn.marsh, collapsedBn.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedBn)
	h2, _ := encodeNodeAndGetHash(node)
	assert.Equal(t, h1, h2)
}

func TestNode_decodeNodeExtensionNode(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()
	encNode, _ := collapsedEn.marsh.Marshal(collapsedEn)
	encNode = append(encNode, extension)

	node, err := decodeNode(encNode, collapsedEn.marsh, collapsedEn.hasher)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(collapsedEn)
	h2, _ := encodeNodeAndGetHash(node)
	assert.Equal(t, h1, h2)
}

func TestNode_decodeNodeLeafNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, leaf)

	node, err := decodeNode(encNode, ln.marsh, ln.hasher)
	assert.Nil(t, err)
	ln.dirty = false

	h1, _ := encodeNodeAndGetHash(ln)
	h2, _ := encodeNodeAndGetHash(node)
	assert.Equal(t, h1, h2)
}

func TestNode_decodeNodeInvalidNode(t *testing.T) {
	t.Parallel()

	ln := getLn(getTestMarshalizerAndHasher())
	invalidNode := byte(6)

	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, invalidNode)

	node, err := decodeNode(encNode, ln.marsh, ln.hasher)
	assert.Nil(t, node)
	assert.Equal(t, ErrInvalidNode, err)
}

func TestNode_decodeNodeInvalidEncoding(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	var encNode []byte

	node, err := decodeNode(encNode, marsh, hasher)
	assert.Nil(t, node)
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

func TestHexToKeyBytes(t *testing.T) {
	t.Parallel()

	reversedHexDoeKey := []byte{5, 6, 15, 6, 4, 6, 16}
	reversedHexDogKey := []byte{7, 6, 15, 6, 4, 6, 16}

	var test = []struct {
		key, hex []byte
	}{
		{reversedHexDoeKey, []byte("doe")},
		{reversedHexDogKey, []byte("dog")},
	}

	for i := range test {
		key, err := hexToKeyBytes(test[i].key)
		assert.Nil(t, err)
		assert.Equal(t, test[i].hex, key)
	}
}

func TestHexToKeyBytesInvalidLength(t *testing.T) {
	t.Parallel()

	key, err := hexToKeyBytes([]byte{6, 4, 6, 15, 6, 5})
	assert.Nil(t, key)
	assert.Equal(t, ErrInvalidLength, err)
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

	leavesChannel, err := tr.GetAllLeavesOnChannel(tr.root.getHash())
	assert.Nil(t, err)
	leaves := make(map[string][]byte)

	for l := range leavesChannel {
		leaves[string(l.Key())] = l.Value()
	}

	assert.Equal(t, 3, len(leaves))
	assert.Equal(t, []byte("reindeer"), leaves["doe"])
	assert.Equal(t, []byte("puppy"), leaves["dog"])
	assert.Equal(t, []byte("cat"), leaves["ddog"])
}

func TestPatriciaMerkleTrie_RecreateFromSnapshotSavesStateToMainDb(t *testing.T) {
	t.Parallel()

	tr, tsm := newEmptyTrie()
	_ = tr.Update([]byte("dog"), []byte("dog"))
	_ = tr.Update([]byte("doe"), []byte("doe"))
	_ = tr.Update([]byte("ddog"), []byte("ddog"))
	_ = tr.Commit()

	rootHash, _ := tr.RootHash()
	tsm.TakeSnapshot(rootHash, true, nil)
	time.Sleep(snapshotDelay * 2)
	err := tsm.db.Remove(rootHash)
	assert.Nil(t, err)

	val, err := tsm.db.Get(rootHash)
	assert.Nil(t, val)
	assert.NotNil(t, err)

	newTr, _ := tr.Recreate(rootHash)
	newPmt, _ := newTr.(*patriciaMerkleTrie)
	newTsm, _ := newPmt.trieStorage.(*trieStorageManager)

	assert.True(t, tsm.Database() != newTsm.snapshots[0])
	val, err = tsm.Database().Get(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)
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

func TestPatriciaMerkleTrie_Close(t *testing.T) {
	numLeavesToAdd := 200
	tr, _ := newEmptyTrie()

	for i := 0; i < numLeavesToAdd; i++ {
		_ = tr.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}
	_ = tr.Commit()

	numBefore := runtime.NumGoroutine()

	rootHash, _ := tr.RootHash()
	leavesChannel1, _ := tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines := runtime.NumGoroutine()
	assert.Equal(t, 1, numGoRoutines-numBefore)

	_, _ = tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 2, numGoRoutines-numBefore)

	_ = tr.Update([]byte("god"), []byte("puppy"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	_, _ = tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 3, numGoRoutines-numBefore)

	_ = tr.Update([]byte("eggod"), []byte("cat"))
	_ = tr.Commit()

	rootHash, _ = tr.RootHash()
	leavesChannel2, _ := tr.GetAllLeavesOnChannel(rootHash)
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 4, numGoRoutines-numBefore)

	for range leavesChannel1 {
	}
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 3, numGoRoutines-numBefore)

	for range leavesChannel2 {
	}
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 2, numGoRoutines-numBefore)

	err := tr.Close()
	assert.Nil(t, err)
	time.Sleep(time.Second * 1)
	numGoRoutines = runtime.NumGoroutine()
	assert.Equal(t, 0, numGoRoutines-numBefore)
}

func TestNode_NodeExtension(t *testing.T) {
	n := &branchNode{
		baseNode: &baseNode{
			hasher: &testscommon.HasherStub{
				ComputeCalled: func(s string) []byte {
					return []byte{0, 0, 0, 0}
				},
			},
		},
	}
	assert.True(t, shouldTestNode(n, make([]byte, 0)))

	n = &branchNode{
		baseNode: &baseNode{
			hasher: &testscommon.HasherStub{
				ComputeCalled: func(s string) []byte {
					return []byte{0, 0, 0, 1}
				},
			},
		},
	}
	assert.False(t, shouldTestNode(n, make([]byte, 0)))

}
