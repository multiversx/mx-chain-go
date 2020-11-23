package trie

import (
	"context"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/stretchr/testify/assert"
)

func TestNode_hashChildrenAndNodeBranchNode(t *testing.T) {
	t.Parallel()

	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshAndHasher())
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

	ln := getLn(getTestMarshAndHasher())
	expectedNodeHash, _ := encodeNodeAndGetHash(ln)

	hash, err := hashChildrenAndNode(ln)
	assert.Nil(t, err)
	assert.Equal(t, expectedNodeHash, hash)
}

func TestNode_encodeNodeAndGetHashBranchNode(t *testing.T) {
	t.Parallel()

	bn, _ := newBranchNode(getTestMarshAndHasher())
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

	marsh, hasher := getTestMarshAndHasher()
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

	marsh, hasher := getTestMarshAndHasher()
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
	_, collapsedBn := getBnAndCollapsedBn(getTestMarshAndHasher())
	encNode, _ := collapsedBn.marsh.Marshal(collapsedBn)
	encNode = append(encNode, branch)
	nodeHash := collapsedBn.hasher.Compute(string(encNode))

	err := encodeNodeAndCommitToDB(collapsedBn, db)
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

	err := encodeNodeAndCommitToDB(collapsedEn, db)
	assert.Nil(t, err)

	val, _ := db.Get(nodeHash)
	assert.Equal(t, encNode, val)
}

func TestNode_encodeNodeAndCommitToDBLeafNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshAndHasher())
	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, leaf)
	nodeHash := ln.hasher.Compute(string(encNode))

	err := encodeNodeAndCommitToDB(ln, db)
	assert.Nil(t, err)

	val, _ := db.Get(nodeHash)
	assert.Equal(t, encNode, val)
}

func TestNode_getNodeFromDBAndDecodeBranchNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshAndHasher())
	_ = bn.commit(false, 0, 5, db, db)

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
	_ = en.commit(false, 0, 5, db, db)

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
	ln := getLn(getTestMarshAndHasher())
	_ = ln.commit(false, 0, 5, db, db)

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
	bn, collapsedBn := getBnAndCollapsedBn(getTestMarshAndHasher())
	childPos := byte(2)
	_ = bn.commit(false, 0, 5, db, db)

	err := resolveIfCollapsed(collapsedBn, childPos, db)
	assert.Nil(t, err)
	assert.False(t, collapsedBn.isCollapsed())
}

func TestNode_resolveIfCollapsedExtensionNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commit(false, 0, 5, db, db)

	err := resolveIfCollapsed(collapsedEn, 0, db)
	assert.Nil(t, err)
	assert.False(t, collapsedEn.isCollapsed())
}

func TestNode_resolveIfCollapsedLeafNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	ln := getLn(getTestMarshAndHasher())
	_ = ln.commit(false, 0, 5, db, db)

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

	bn, _ := getBnAndCollapsedBn(getTestMarshAndHasher())
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

	_, collapsedBn := getBnAndCollapsedBn(getTestMarshAndHasher())
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

	ln := getLn(getTestMarshAndHasher())
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

	ln := getLn(getTestMarshAndHasher())
	invalidNode := byte(6)

	encNode, _ := ln.marsh.Marshal(ln)
	encNode = append(encNode, invalidNode)

	node, err := decodeNode(encNode, ln.marsh, ln.hasher)
	assert.Nil(t, node)
	assert.Equal(t, ErrInvalidNode, err)
}

func TestNode_decodeNodeInvalidEncoding(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshAndHasher()
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
	root, _ := tr.Root()

	_ = tr.Update([]byte("dog"), []byte("value of dog"))

	assert.Equal(t, 4, len(tr.oldHashes))
	assert.Equal(t, root, tr.oldRoot)

	_ = tr.Commit()

	assert.Equal(t, 0, len(tr.oldHashes))
	assert.Equal(t, 0, len(tr.oldRoot))
}

func TestTrieResetOldHashes(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()

	_ = tr.Update([]byte("doeee"), []byte("value of doeee"))

	assert.NotEqual(t, 0, len(tr.oldHashes))
	assert.NotEqual(t, 0, len(tr.oldRoot))

	expectedHashes := tr.oldHashes
	hashes := tr.ResetOldHashes()
	assert.Equal(t, expectedHashes, hashes)
	assert.Equal(t, 0, len(tr.oldHashes))
	assert.Equal(t, 0, len(tr.oldRoot))
}

func TestTrieAddHashesToOldHashes(t *testing.T) {
	t.Parallel()

	hashes := [][]byte{[]byte("one"), []byte("two"), []byte("three")}
	tr := initTrie()
	_ = tr.Commit()

	_ = tr.Update([]byte("doeee"), []byte("value of doeee"))

	expectedHLength := len(tr.oldHashes) + len(hashes)
	tr.AppendToOldHashes(hashes)
	assert.Equal(t, expectedHLength, len(tr.oldHashes))
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

func TestPruningAndPruningCancellingOnTrieRollback(t *testing.T) {
	t.Parallel()

	testVals := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("doe"), []byte("reindeer")},
		{[]byte("dog"), []byte("puppy")},
		{[]byte("dogglesworth"), []byte("cat")},
		{[]byte("horse"), []byte("stallion")},
	}

	tr, _, _ := newEmptyTrie()

	rootHashes := make([][]byte, 0)
	for _, testVal := range testVals {
		_ = tr.Update(testVal.key, testVal.value)

		newHashes, _ := tr.GetDirtyHashes()
		tr.SetNewHashes(newHashes)
		_ = tr.Commit()
		rootHashes = append(rootHashes, tr.root.getHash())
	}

	for i := 0; i < len(rootHashes); i++ {
		_, err := tr.Recreate(rootHashes[i])
		assert.Nil(t, err)
	}

	tr.CancelPrune(rootHashes[0], data.NewRoot)
	finalizeTrieState(t, 1, tr, rootHashes)
	finalizeTrieState(t, 2, tr, rootHashes)
	rollbackTrieState(t, 3, tr, rootHashes)

	_, err := tr.Recreate(rootHashes[2])
	assert.Nil(t, err)
}

func finalizeTrieState(t *testing.T, index int, tr data.Trie, rootHashes [][]byte) {
	tr.Prune(rootHashes[index-1], data.OldRoot)
	tr.CancelPrune(rootHashes[index], data.NewRoot)
	time.Sleep(pruningDelay)

	_, err := tr.Recreate(rootHashes[index-1])
	assert.NotNil(t, err)
}

func rollbackTrieState(t *testing.T, index int, tr data.Trie, rootHashes [][]byte) {
	tr.Prune(rootHashes[index], data.NewRoot)
	tr.CancelPrune(rootHashes[index-1], data.OldRoot)
	time.Sleep(pruningDelay)

	_, err := tr.Recreate(rootHashes[index])
	assert.NotNil(t, err)
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

	leavesChannel, err := tr.GetAllLeavesOnChannel(tr.root.getHash(), context.Background())
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

func TestPatriciaMerkleTrie_removeDuplicatedKeys(t *testing.T) {
	map1 := map[string]struct{}{
		"hash1": {},
		"hash2": {},
		"hash3": {},
		"hash4": {},
	}

	map2 := map[string]struct{}{
		"hash1": {},
		"hash4": {},
		"hash5": {},
		"hash6": {},
	}

	removeDuplicatedKeys(map1, map2)

	_, ok := map1["hash1"]
	assert.False(t, ok)
	_, ok = map1["hash4"]
	assert.False(t, ok)

	_, ok = map2["hash1"]
	assert.False(t, ok)
	_, ok = map2["hash4"]
	assert.False(t, ok)
}

func TestPatriciaMerkleTrie_RecreateFromSnapshotSavesStateToMainDb(t *testing.T) {
	t.Parallel()

	tr, tsm, _ := newEmptyTrie()
	_ = tr.Update([]byte("dog"), []byte("dog"))
	_ = tr.Update([]byte("doe"), []byte("doe"))
	_ = tr.Update([]byte("ddog"), []byte("ddog"))
	newHashes, _ := tr.GetDirtyHashes()
	tr.SetNewHashes(newHashes)
	_ = tr.Commit()

	rootHash, _ := tr.Root()
	tr.TakeSnapshot(rootHash)
	time.Sleep(snapshotDelay * 2)
	tr.Prune(rootHash, data.NewRoot)

	val, err := tsm.db.Get(rootHash)
	assert.Nil(t, val)
	assert.NotNil(t, err)

	newTr, _ := tr.Recreate(rootHash)
	newPmt, _ := newTr.(*patriciaMerkleTrie)
	newTsm, _ := newPmt.trieStorage.(*trieStorageManager)

	assert.True(t, newTr.Database() != newTsm.snapshots[0])
	val, err = newTr.Database().Get(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)
}

func TestPatriciaMerkleTrie_newHashesAndOldHashesAreResetAfterEveryCommit(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()

	_ = tr.Update([]byte("doe"), []byte("deer"))
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	newHashes, _ := tr.GetDirtyHashes()
	tr.SetNewHashes(newHashes)

	err := tr.Commit()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tr.oldHashes))
	assert.Equal(t, 0, len(tr.newHashes))
}
