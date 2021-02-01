package trie

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/stretchr/testify/assert"
)

func getEnAndCollapsedEn() (*extensionNode, *extensionNode) {
	child, collapsedChild := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	en, _ := newExtensionNode([]byte("d"), child, child.marsh, child.hasher)

	childHash, _ := encodeNodeAndGetHash(collapsedChild)
	collapsedEn := &extensionNode{CollapsedEn: CollapsedEn{Key: []byte("d"), EncodedChild: childHash}, baseNode: &baseNode{}}
	collapsedEn.marsh = child.marsh
	collapsedEn.hasher = child.hasher
	return en, collapsedEn
}

func TestExtensionNode_newExtensionNode(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	expectedEn := &extensionNode{
		CollapsedEn: CollapsedEn{
			Key:          []byte("dog"),
			EncodedChild: nil,
		},
		child: bn,
		baseNode: &baseNode{
			dirty:  true,
			marsh:  bn.marsh,
			hasher: bn.hasher,
		},
	}
	en, _ := newExtensionNode([]byte("dog"), bn, bn.marsh, bn.hasher)
	assert.Equal(t, expectedEn, en)
}

func TestExtensionNode_getHash(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{hash: []byte("test hash")}}
	assert.Equal(t, en.hash, en.getHash())
}

func TestExtensionNode_isDirty(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{dirty: true}}
	assert.Equal(t, true, en.isDirty())

	en = &extensionNode{baseNode: &baseNode{dirty: false}}
	assert.Equal(t, false, en.isDirty())
}

func TestExtensionNode_getCollapsed(t *testing.T) {
	t.Parallel()

	en, collapsedEn := getEnAndCollapsedEn()
	collapsedEn.dirty = true

	collapsed, err := en.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, collapsedEn, collapsed)
}

func TestExtensionNode_getCollapsedEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}

	collapsed, err := en.getCollapsed()
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
	assert.Nil(t, collapsed)
}

func TestExtensionNode_getCollapsedNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	collapsed, err := en.getCollapsed()
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
	assert.Nil(t, collapsed)
}

func TestExtensionNode_getCollapsedCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()

	collapsed, err := collapsedEn.getCollapsed()
	assert.Nil(t, err)
	assert.Equal(t, collapsedEn, collapsed)
}

func TestExtensionNode_setHash(t *testing.T) {
	t.Parallel()

	en, collapsedEn := getEnAndCollapsedEn()
	hash, _ := encodeNodeAndGetHash(collapsedEn)

	err := en.setHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, en.hash)
}

func TestExtensionNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}

	err := en.setHash()
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
	assert.Nil(t, en.hash)
}

func TestExtensionNode_setHashNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	err := en.setHash()
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
	assert.Nil(t, en)
}

func TestExtensionNode_setHashCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()
	hash, _ := encodeNodeAndGetHash(collapsedEn)

	err := collapsedEn.setHash()
	assert.Nil(t, err)
	assert.Equal(t, hash, collapsedEn.hash)
}

func TestExtensionNode_setGivenHash(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}
	expectedHash := []byte("node hash")

	en.setGivenHash(expectedHash)
	assert.Equal(t, expectedHash, en.hash)
}

func TestExtensionNode_hashChildren(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	assert.Nil(t, en.child.getHash())

	err := en.hashChildren()
	assert.Nil(t, err)

	childHash, _ := encodeNodeAndGetHash(en.child)
	assert.Equal(t, childHash, en.child.getHash())
}

func TestExtensionNode_hashChildrenEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}

	err := en.hashChildren()
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
}

func TestExtensionNode_hashChildrenNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	err := en.hashChildren()
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
}

func TestExtensionNode_hashChildrenCollapsedNode(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()

	err := collapsedEn.hashChildren()
	assert.Nil(t, err)

	_, collapsedEn2 := getEnAndCollapsedEn()
	assert.Equal(t, collapsedEn2, collapsedEn)
}

func TestExtensionNode_hashNode(t *testing.T) {
	t.Parallel()

	_, collapsedEn := getEnAndCollapsedEn()
	expectedHash, _ := encodeNodeAndGetHash(collapsedEn)

	hash, err := collapsedEn.hashNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestExtensionNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}

	hash, err := en.hashNode()
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
	assert.Nil(t, hash)
}

func TestExtensionNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	hash, err := en.hashNode()
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
	assert.Nil(t, hash)
}

func TestExtensionNode_commit(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	hash, _ := encodeNodeAndGetHash(collapsedEn)
	_ = en.setHash()

	err := en.commit(false, 0, 5, db, db)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	n, _ := decodeNode(encNode, en.marsh, en.hasher)

	h1, _ := encodeNodeAndGetHash(collapsedEn)
	h2, _ := encodeNodeAndGetHash(n)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_commitEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}

	err := en.commit(false, 0, 5, nil, nil)
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
}

func TestExtensionNode_commitNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	err := en.commit(false, 0, 5, nil, nil)
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
}

func TestExtensionNode_commitCollapsedNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	_, collapsedEn := getEnAndCollapsedEn()
	hash, _ := encodeNodeAndGetHash(collapsedEn)
	_ = collapsedEn.setHash()

	collapsedEn.dirty = true
	err := collapsedEn.commit(false, 0, 5, db, db)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	n, _ := decodeNode(encNode, collapsedEn.marsh, collapsedEn.hasher)
	collapsedEn.hash = nil

	h1, _ := encodeNodeAndGetHash(collapsedEn)
	h2, _ := encodeNodeAndGetHash(n)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_getEncodedNode(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	expectedEncodedNode, _ := en.marsh.Marshal(en)
	expectedEncodedNode = append(expectedEncodedNode, extension)

	encNode, err := en.getEncodedNode()
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestExtensionNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}

	encNode, err := en.getEncodedNode()
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
	assert.Nil(t, encNode)
}

func TestExtensionNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	encNode, err := en.getEncodedNode()
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
	assert.Nil(t, encNode)
}

func TestExtensionNode_resolveCollapsed(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.setHash()
	_ = en.commit(false, 0, 5, db, db)
	_, resolved := getBnAndCollapsedBn(en.marsh, en.hasher)

	err := collapsedEn.resolveCollapsed(0, db)
	assert.Nil(t, err)
	assert.Equal(t, en.child.getHash(), collapsedEn.child.getHash())

	h1, _ := encodeNodeAndGetHash(resolved)
	h2, _ := encodeNodeAndGetHash(collapsedEn.child)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_resolveCollapsedEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}

	err := en.resolveCollapsed(0, nil)
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
}

func TestExtensionNode_resolveCollapsedNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	err := en.resolveCollapsed(2, nil)
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
}

func TestExtensionNode_isCollapsed(t *testing.T) {
	t.Parallel()

	en, collapsedEn := getEnAndCollapsedEn()
	assert.True(t, collapsedEn.isCollapsed())
	assert.False(t, en.isCollapsed())

	collapsedEn.child, _ = newLeafNode([]byte("og"), []byte("dog"), en.marsh, en.hasher)
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

	val, err := en.tryGet(key, nil)
	assert.Equal(t, dogBytes, val)
	assert.Nil(t, err)
}

func TestExtensionNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	var key []byte

	val, err := en.tryGet(key, nil)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestExtensionNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	key := []byte("gdo")

	val, err := en.tryGet(key, nil)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestExtensionNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.setHash()
	_ = en.commit(false, 0, 5, db, db)

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)

	val, err := collapsedEn.tryGet(key, db)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestExtensionNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
	key := []byte("dog")

	val, err := en.tryGet(key, nil)
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
	assert.Nil(t, val)
}

func TestExtensionNode_tryGetNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode
	key := []byte("dog")

	val, err := en.tryGet(key, nil)
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
	assert.Nil(t, val)
}

func TestExtensionNode_getNext(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	nextNode, _ := getBnAndCollapsedBn(en.marsh, en.hasher)

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)

	n, newKey, err := en.getNext(key, nil)
	assert.Equal(t, nextNode, n)
	assert.Equal(t, key[1:], newKey)
	assert.Nil(t, err)
}

func TestExtensionNode_getNextWrongKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(bnKey, lnKey...)

	n, key, err := en.getNext(key, nil)
	assert.Nil(t, n)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestExtensionNode_insert(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	key := []byte{100, 15, 5, 6}
	n, _ := newLeafNode(key, []byte("dogs"), en.marsh, en.hasher)

	dirty, newNode, _, err := en.insert(n, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, _ := newNode.tryGet(key, nil)
	assert.Equal(t, []byte("dogs"), val)
}

func TestExtensionNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	key := []byte{100, 15, 5, 6}
	n, _ := newLeafNode(key, []byte("dogs"), en.marsh, en.hasher)

	_ = en.setHash()
	_ = en.commit(false, 0, 5, db, db)

	dirty, newNode, _, err := collapsedEn.insert(n, db)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, _ := newNode.tryGet(key, db)
	assert.Equal(t, []byte("dogs"), val)
}

func TestExtensionNode_insertInStoredEnSameKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	enKey := []byte{100}
	key := append(enKey, []byte{11, 12}...)
	n, _ := newLeafNode(key, []byte("dogs"), en.marsh, en.hasher)

	_ = en.commit(false, 0, 5, db, db)
	enHash := en.getHash()
	bn, _, _ := en.getNext(enKey, db)
	bnHash := bn.getHash()
	expectedHashes := [][]byte{bnHash, enHash}

	dirty, _, oldHashes, err := en.insert(n, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestExtensionNode_insertInStoredEnDifferentKey(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	enKey := []byte{1}
	en, _ := newExtensionNode(enKey, bn, bn.marsh, bn.hasher)
	nodeKey := []byte{11, 12}
	n, _ := newLeafNode(nodeKey, []byte("dogs"), bn.marsh, bn.hasher)

	_ = en.commit(false, 0, 5, db, db)
	expectedHashes := [][]byte{en.getHash()}

	dirty, _, oldHashes, err := en.insert(n, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestExtensionNode_insertInDirtyEnSameKey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	nodeKey := []byte{100, 11, 12}
	n, _ := newLeafNode(nodeKey, []byte("dogs"), en.marsh, en.hasher)

	dirty, _, oldHashes, err := en.insert(n, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestExtensionNode_insertInDirtyEnDifferentKey(t *testing.T) {
	t.Parallel()

	bn, _ := getBnAndCollapsedBn(getTestMarshalizerAndHasher())
	enKey := []byte{1}
	en, _ := newExtensionNode(enKey, bn, bn.marsh, bn.hasher)
	nodeKey := []byte{11, 12}
	n, _ := newLeafNode(nodeKey, []byte("dogs"), bn.marsh, bn.hasher)

	dirty, _, oldHashes, err := en.insert(n, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestExtensionNode_insertInNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	dirty, newNode, _, err := en.insert(&leafNode{}, nil)
	assert.False(t, dirty)
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
	assert.Nil(t, newNode)
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

	val, _ := en.tryGet(key, nil)
	assert.Equal(t, dogBytes, val)

	dirty, _, _, err := en.delete(key, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ = en.tryGet(key, nil)
	assert.Nil(t, val)
}

func TestExtensionNode_deleteFromStoredEn(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)
	lnPathKey := key

	_ = en.commit(false, 0, 5, db, db)
	bn, key, _ := en.getNext(key, db)
	ln, _, _ := bn.getNext(key, db)
	expectedHashes := [][]byte{ln.getHash(), bn.getHash(), en.getHash()}

	dirty, _, oldHashes, err := en.delete(lnPathKey, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedHashes, oldHashes)
}

func TestExtensionNode_deleteFromDirtyEn(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	lnKey := []byte{100, 2, 100, 111, 103}

	dirty, _, oldHashes, err := en.delete(lnKey, nil)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, [][]byte{}, oldHashes)
}

func TestExtendedNode_deleteEmptyNode(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}

	dirty, newNode, _, err := en.delete([]byte("dog"), nil)
	assert.False(t, dirty)
	assert.True(t, errors.Is(err, ErrEmptyExtensionNode))
	assert.Nil(t, newNode)
}

func TestExtensionNode_deleteNilNode(t *testing.T) {
	t.Parallel()

	var en *extensionNode

	dirty, newNode, _, err := en.delete([]byte("dog"), nil)
	assert.False(t, dirty)
	assert.True(t, errors.Is(err, ErrNilExtensionNode))
	assert.Nil(t, newNode)
}

func TestExtensionNode_deleteEmptykey(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()

	dirty, newNode, _, err := en.delete([]byte{}, nil)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newNode)
}

func TestExtensionNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.setHash()
	_ = en.commit(false, 0, 5, db, db)

	enKey := []byte{100}
	bnKey := []byte{2}
	lnKey := []byte("dog")
	key := append(enKey, bnKey...)
	key = append(key, lnKey...)

	val, _ := en.tryGet(key, db)
	assert.Equal(t, []byte("dog"), val)

	dirty, newNode, _, err := collapsedEn.delete(key, db)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ = newNode.tryGet(key, db)
	assert.Nil(t, val)
}

func TestExtensionNode_reduceNode(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	en, _ := newExtensionNode([]byte{100, 111, 103}, nil, marsh, hasher)

	expected := &extensionNode{CollapsedEn: CollapsedEn{Key: []byte{2, 100, 111, 103}}, baseNode: &baseNode{dirty: true}}
	expected.marsh = en.marsh
	expected.hasher = en.hasher

	n, newChildPos, err := en.reduceNode(2)
	assert.Equal(t, expected, n)
	assert.Nil(t, err)
	assert.True(t, newChildPos)
}

func TestExtensionNode_reduceNodeCollapsedNode(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()
	collapsedTrie, _ := tr.Recreate(rootHash)

	err := collapsedTrie.Delete([]byte("doe"))
	assert.Nil(t, err)

	err = collapsedTrie.Commit()
	assert.Nil(t, err)
}

func TestExtensionNode_clone(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	clone := en.clone()
	assert.False(t, en == clone)
	assert.Equal(t, en, clone)
}

func TestExtensionNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
	assert.Equal(t, ErrEmptyExtensionNode, en.isEmptyOrNil())

	en = nil
	assert.Equal(t, ErrNilExtensionNode, en.isEmptyOrNil())
}

//------- deepClone

func TestExtensionNode_deepCloneNilHashShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}
	en.dirty = true
	en.hash = nil
	en.EncodedChild = getRandomByteSlice()
	en.Key = getRandomByteSlice()
	en.child = &leafNode{baseNode: &baseNode{}}

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func TestExtensionNode_deepCloneNilEncodedChildShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}
	en.dirty = true
	en.hash = getRandomByteSlice()
	en.EncodedChild = nil
	en.Key = getRandomByteSlice()
	en.child = &leafNode{baseNode: &baseNode{}}

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func TestExtensionNode_deepCloneNilKeyShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}
	en.dirty = true
	en.hash = getRandomByteSlice()
	en.EncodedChild = getRandomByteSlice()
	en.Key = nil
	en.child = &leafNode{baseNode: &baseNode{}}

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func TestExtensionNode_getChildren(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()

	children, err := en.getChildren(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(children))
}

func TestExtensionNode_getChildrenCollapsedEn(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commit(true, 0, 5, db, db)

	children, err := collapsedEn.getChildren(db)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(children))
}

func TestExtensionNode_isValid(t *testing.T) {
	t.Parallel()

	en, _ := getEnAndCollapsedEn()
	assert.True(t, en.isValid())

	en.child = nil
	assert.False(t, en.isValid())
}

func TestExtensionNode_setDirty(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}
	en.setDirty(true)

	assert.True(t, en.dirty)
}

func TestExtensionNode_loadChildren(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	tr, _, _ := newEmptyTrie()
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.root.setRootHash()
	nodes, _ := getEncodedTrieNodesAndHashes(tr)
	nodesCacher, _ := lrucache.NewCache(100)
	for i := range nodes {
		n, _ := NewInterceptedTrieNode(nodes[i], marsh, hasher)
		nodesCacher.Put(n.hash, n, len(n.EncodedNode()))
	}

	en := getCollapsedEn(t, tr.root)

	getNode := func(hash []byte) (node, error) {
		cacheData, _ := nodesCacher.Get(hash)
		return trieNode(cacheData)
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

func TestExtensionNode_deepCloneNilChildShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}
	en.dirty = true
	en.hash = getRandomByteSlice()
	en.EncodedChild = getRandomByteSlice()
	en.Key = getRandomByteSlice()
	en.child = nil

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func TestExtensionNode_deepCloneShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{baseNode: &baseNode{}}
	en.dirty = true
	en.hash = getRandomByteSlice()
	en.EncodedChild = getRandomByteSlice()
	en.Key = getRandomByteSlice()
	en.child = &leafNode{baseNode: &baseNode{}}

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func testSameExtensionNodeContent(t *testing.T, expected *extensionNode, actual *extensionNode) {
	if !reflect.DeepEqual(expected, actual) {
		assert.Fail(t, "not equal content")
		fmt.Printf(
			"expected:\n %s, got: \n%s",
			getExtensionNodeContents(expected),
			getExtensionNodeContents(actual),
		)
	}
	assert.False(t, expected == actual)
}

func getExtensionNodeContents(en *extensionNode) string {
	str := fmt.Sprintf(`extension node:
   		key: %s
   		encoded child: %s
		hash: %s
   		child: %p,	
   		dirty: %v
`,
		hex.EncodeToString(en.Key),
		hex.EncodeToString(en.EncodedChild),
		hex.EncodeToString(en.hash),
		en.child,
		en.dirty)

	return str
}

func TestExtensionNode_newExtensionNodeNilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	en, err := newExtensionNode([]byte("key"), &branchNode{}, nil, mock.HasherMock{})
	assert.Nil(t, en)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestExtensionNode_newExtensionNodeNilHasherShouldErr(t *testing.T) {
	t.Parallel()

	en, err := newExtensionNode([]byte("key"), &branchNode{}, &mock.MarshalizerMock{}, nil)
	assert.Nil(t, en)
	assert.Equal(t, ErrNilHasher, err)
}

func TestExtensionNode_newExtensionNodeOkVals(t *testing.T) {
	t.Parallel()

	marsh, hasher := getTestMarshalizerAndHasher()
	key := []byte("key")
	child := &branchNode{}
	en, err := newExtensionNode(key, child, marsh, hasher)

	assert.Nil(t, err)
	assert.Equal(t, key, en.Key)
	assert.Nil(t, en.EncodedChild)
	assert.Equal(t, child, en.child)
	assert.Equal(t, hasher, en.hasher)
	assert.Equal(t, marsh, en.marsh)
	assert.True(t, en.dirty)
}

func TestExtensionNode_getMarshalizer(t *testing.T) {
	t.Parallel()

	marsh, _ := getTestMarshalizerAndHasher()
	en := &extensionNode{
		baseNode: &baseNode{
			marsh: marsh,
		},
	}

	assert.Equal(t, marsh, en.getMarshalizer())
}

func TestExtensionNode_commitCollapsesTrieIfMaxTrieLevelInMemoryIsReached(t *testing.T) {
	t.Parallel()

	en, collapsedEn := getEnAndCollapsedEn()
	_ = collapsedEn.setRootHash()

	err := en.commit(true, 0, 1, mock.NewMemDbMock(), mock.NewMemDbMock())
	assert.Nil(t, err)

	assert.Equal(t, collapsedEn.EncodedChild, en.EncodedChild)
	assert.Equal(t, collapsedEn.child, en.child)
	assert.Equal(t, collapsedEn.hash, en.hash)
}

func TestExtensionNode_printShouldNotPanicEvenIfNodeIsCollapsed(t *testing.T) {
	t.Parallel()

	enWriter := bytes.NewBuffer(make([]byte, 0))
	collapsedEnWriter := bytes.NewBuffer(make([]byte, 0))

	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commit(true, 0, 5, db, db)
	_ = collapsedEn.commit(true, 0, 5, db, db)

	en.print(enWriter, 0, db)
	collapsedEn.print(collapsedEnWriter, 0, db)

	assert.Equal(t, enWriter.Bytes(), collapsedEnWriter.Bytes())
}

func TestExtensionNode_getDirtyHashesFromCleanNode(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	_ = en.commit(true, 0, 5, db, db)
	dirtyHashes := make(data.ModifiedHashes)

	err := en.getDirtyHashes(dirtyHashes)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(dirtyHashes))
}

func TestExtensionNode_getAllHashes(t *testing.T) {
	t.Parallel()

	trieNodes := 5
	en, _ := getEnAndCollapsedEn()

	hashes, err := en.getAllHashes(mock.NewMemDbMock())
	assert.Nil(t, err)
	assert.Equal(t, trieNodes, len(hashes))
}

func TestExtensionNode_getAllHashesResolvesCollapsed(t *testing.T) {
	t.Parallel()

	trieNodes := 5
	db := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	_ = en.commit(true, 0, 5, db, db)

	hashes, err := collapsedEn.getAllHashes(db)
	assert.Nil(t, err)
	assert.Equal(t, trieNodes, len(hashes))
}
