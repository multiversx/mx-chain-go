package trie

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	protobuf "github.com/ElrondNetwork/elrond-go/data/trie/proto"
	"github.com/stretchr/testify/assert"
)

func getEnAndCollapsedEn() (*extensionNode, *extensionNode) {
	marsh, hasher := getTestMarshAndHasher()
	child, collapsedChild := getBnAndCollapsedBn()
	en := newExtensionNode([]byte("d"), child)

	childHash, _ := encodeNodeAndGetHash(collapsedChild, marsh, hasher)
	collapsedEn := &extensionNode{CollapsedEn: protobuf.CollapsedEn{Key: []byte("d"), EncodedChild: childHash}}
	return en, collapsedEn
}

func TestExtensionNode_newExtensionNode(t *testing.T) {
	t.Parallel()
	bn, _ := getBnAndCollapsedBn()
	expectedEn := &extensionNode{
		CollapsedEn: protobuf.CollapsedEn{
			Key:          []byte("dog"),
			EncodedChild: nil,
		},
		child: bn,
		hash:  nil,
		dirty: true,
	}
	en := newExtensionNode([]byte("dog"), bn)
	assert.Equal(t, expectedEn, en)
}

func TestExtensionNode_getHash(t *testing.T) {
	t.Parallel()
	en := &extensionNode{hash: []byte("test hash")}
	assert.Equal(t, en.hash, en.getHash())
}

func TestExtensionNode_isDirty(t *testing.T) {
	t.Parallel()
	en := &extensionNode{dirty: true}
	assert.Equal(t, true, en.isDirty())

	en = &extensionNode{dirty: false}
	assert.Equal(t, false, en.isDirty())
}

func TestExtensionNode_getCollapsed(t *testing.T) {
	t.Parallel()
	en, collapsedEn := getEnAndCollapsedEn()
	collapsedEn.dirty = true
	marsh, hasher := getTestMarshAndHasher()

	collapsed, err := en.getCollapsed(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, collapsedEn, collapsed)
}

func TestExtensionNode_getCollapsedEmptyNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	en := &extensionNode{}

	collapsed, err := en.getCollapsed(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, collapsed)
}

func TestExtensionNode_getCollapsedNilNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	var en *extensionNode

	collapsed, err := en.getCollapsed(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, collapsed)
}

func TestExtensionNode_getCollapsedCollapsedNode(t *testing.T) {
	t.Parallel()
	_, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	collapsed, err := collapsedEn.getCollapsed(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, collapsedEn, collapsed)
}

func TestExtensionNode_setHash(t *testing.T) {
	t.Parallel()
	en, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedEn, marsh, hasher)

	err := en.setHash(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, hash, en.hash)
}

func TestExtensionNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	en := &extensionNode{}

	err := en.setHash(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, en.hash)
}

func TestExtensionNode_setHashNilNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	var en *extensionNode

	err := en.setHash(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, en)
}

func TestExtensionNode_setHashCollapsedNode(t *testing.T) {
	t.Parallel()
	_, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedEn, marsh, hasher)

	err := collapsedEn.setHash(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, hash, collapsedEn.hash)
}

func TestExtensionNode_hashChildren(t *testing.T) {
	t.Parallel()
	en, _ := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	assert.Nil(t, en.child.getHash())

	err := en.hashChildren(marsh, hasher)
	assert.Nil(t, err)

	childHash, _ := encodeNodeAndGetHash(en.child, marsh, hasher)
	assert.Equal(t, childHash, en.child.getHash())
}

func TestExtensionNode_hashChildrenEmptyNode(t *testing.T) {
	t.Parallel()
	en := &extensionNode{}
	marsh, hasher := getTestMarshAndHasher()

	err := en.hashChildren(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestExtensionNode_hashChildrenNilNode(t *testing.T) {
	t.Parallel()
	var en *extensionNode
	marsh, hasher := getTestMarshAndHasher()

	err := en.hashChildren(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
}

func TestExtensionNode_hashChildrenCollapsedNode(t *testing.T) {
	t.Parallel()
	_, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	err := collapsedEn.hashChildren(marsh, hasher)
	assert.Nil(t, err)

	_, collapsedEn2 := getEnAndCollapsedEn()
	assert.Equal(t, collapsedEn2, collapsedEn)
}

func TestExtensionNode_hashNode(t *testing.T) {
	t.Parallel()
	_, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	expectedHash, _ := encodeNodeAndGetHash(collapsedEn, marsh, hasher)
	hash, err := collapsedEn.hashNode(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestExtensionNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()
	en := &extensionNode{}
	marsh, hasher := getTestMarshAndHasher()

	hash, err := en.hashNode(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, hash)
}

func TestExtensionNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()
	var en *extensionNode
	marsh, hasher := getTestMarshAndHasher()

	hash, err := en.hashNode(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, hash)
}

func TestExtensionNode_commit(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedEn, marsh, hasher)
	_ = en.setHash(marsh, hasher)

	err := en.commit(0, db, marsh, hasher)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	node, _ := decodeNode(encNode, marsh)

	h1, _ := encodeNodeAndGetHash(collapsedEn, marsh, hasher)
	h2, _ := encodeNodeAndGetHash(node, marsh, hasher)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_commitEmptyNode(t *testing.T) {
	t.Parallel()
	en := &extensionNode{}
	db, _ := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()

	err := en.commit(0, db, marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestExtensionNode_commitNilNode(t *testing.T) {
	t.Parallel()
	var en *extensionNode
	db, _ := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()

	err := en.commit(0, db, marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
}

func TestExtensionNode_commitCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	_, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedEn, marsh, hasher)
	_ = collapsedEn.setHash(marsh, hasher)

	collapsedEn.dirty = true
	err := collapsedEn.commit(0, db, marsh, hasher)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	node, _ := decodeNode(encNode, marsh)
	collapsedEn.hash = nil

	h1, _ := encodeNodeAndGetHash(collapsedEn, marsh, hasher)
	h2, _ := encodeNodeAndGetHash(node, marsh, hasher)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_getEncodedNode(t *testing.T) {
	t.Parallel()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()

	expectedEncodedNode, _ := marsh.Marshal(en)
	expectedEncodedNode = append(expectedEncodedNode, extension)

	encNode, err := en.getEncodedNode(marsh)
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestExtensionNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()
	en := &extensionNode{}
	marsh, _ := getTestMarshAndHasher()

	encNode, err := en.getEncodedNode(marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, encNode)
}

func TestExtensionNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()
	var en *extensionNode
	marsh, _ := getTestMarshAndHasher()

	encNode, err := en.getEncodedNode(marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, encNode)
}

func TestExtensionNode_resolveCollapsed(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()

	_ = en.setHash(marsh, hasher)
	_ = en.commit(0, db, marsh, hasher)
	_, resolved := getBnAndCollapsedBn()

	err := collapsedEn.resolveCollapsed(0, db, marsh)
	assert.Nil(t, err)

	h1, _ := encodeNodeAndGetHash(resolved, marsh, hasher)
	h2, _ := encodeNodeAndGetHash(collapsedEn.child, marsh, hasher)
	assert.Equal(t, h1, h2)
}

func TestExtensionNode_resolveCollapsedEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en := &extensionNode{}
	marsh, _ := getTestMarshAndHasher()

	err := en.resolveCollapsed(0, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestExtensionNode_resolveCollapsedENilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var en *extensionNode
	marsh, _ := getTestMarshAndHasher()

	err := en.resolveCollapsed(2, db, marsh)
	assert.Equal(t, ErrNilNode, err)
}

func TestExtensionNode_isCollapsed(t *testing.T) {
	t.Parallel()
	en, collapsedEn := getEnAndCollapsedEn()

	assert.True(t, collapsedEn.isCollapsed())
	assert.False(t, en.isCollapsed())

	collapsedEn.child = newLeafNode([]byte("og"), []byte("dog"))
	assert.False(t, collapsedEn.isCollapsed())
}

func TestExtensionNode_tryGet(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 2, 100, 111, 103}
	val, err := en.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestExtensionNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()

	var key []byte
	val, err := en.tryGet(key, db, marsh)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestExtensionNode_tryGetWrongKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{103, 100, 111}
	val, err := en.tryGet(key, db, marsh)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestExtensionNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()
	_ = en.setHash(marsh, hasher)
	_ = en.commit(0, db, marsh, hasher)

	key := []byte{100, 2, 100, 111, 103}
	val, err := collapsedEn.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestExtensionNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en := &extensionNode{}
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 111, 103}
	val, err := en.tryGet(key, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, val)
}

func TestExtensionNode_tryGetNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var en *extensionNode
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 111, 103}
	val, err := en.tryGet(key, db, marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, val)
}

func TestExtensionNode_getNext(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()
	nextNode, _ := getBnAndCollapsedBn()
	key := []byte{100, 2, 100, 111, 103}

	node, key, err := en.getNext(key, db, marsh)
	assert.Equal(t, nextNode, node)
	assert.Equal(t, []byte{2, 100, 111, 103}, key)
	assert.Nil(t, err)
}

func TestExtensionNode_getNextWrongKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()
	key := []byte{2, 100, 111, 103}

	node, key, err := en.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestExtensionNode_insert(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	node := newLeafNode([]byte{100, 15, 5, 6}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := en.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newNode.tryGet([]byte{100, 15, 5, 6}, db, marsh)
	assert.Equal(t, []byte("dogs"), val)
}

func TestExtensionNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	node := newLeafNode([]byte{100, 15, 5, 6}, []byte("dogs"))
	marsh, hasher := getTestMarshAndHasher()
	_ = en.setHash(marsh, hasher)
	_ = en.commit(0, db, marsh, hasher)

	dirty, newNode, err := collapsedEn.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newNode.tryGet([]byte{100, 15, 5, 6}, db, marsh)
	assert.Equal(t, []byte("dogs"), val)
}

func TestExtensionNode_insertInNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var en *extensionNode
	node := newLeafNode([]byte{0, 2, 3}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := en.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newNode)
}

func TestExtensionNode_delete(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()

	val, _ := en.tryGet([]byte{100, 2, 100, 111, 103}, db, marsh)
	assert.Equal(t, []byte("dog"), val)

	dirty, _, err := en.delete([]byte{100, 2, 100, 111, 103}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ = en.tryGet([]byte{100, 2, 100, 111, 103}, db, marsh)
	assert.Nil(t, val)
}

func TestExtendedNode_deleteEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en := &extensionNode{}
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := en.delete([]byte{100, 111, 103}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, newNode)
}

func TestExtensionNode_deleteNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var en *extensionNode
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := en.delete([]byte{100, 111, 103}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newNode)
}

func TestExtensionNode_deleteEmptykey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, _ := getEnAndCollapsedEn()
	marsh, _ := getTestMarshAndHasher()

	dirty, newNode, err := en.delete([]byte{}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newNode)
}

func TestExtensionNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	en, collapsedEn := getEnAndCollapsedEn()
	marsh, hasher := getTestMarshAndHasher()
	_ = en.setHash(marsh, hasher)
	_ = en.commit(0, db, marsh, hasher)

	val, _ := en.tryGet([]byte{100, 2, 100, 111, 103}, db, marsh)
	assert.Equal(t, []byte("dog"), val)

	dirty, newNode, err := collapsedEn.delete([]byte{100, 2, 100, 111, 103}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ = newNode.tryGet([]byte{100, 2, 100, 111, 103}, db, marsh)
	assert.Nil(t, val)
}

func TestExtensionNode_reduceNode(t *testing.T) {
	t.Parallel()
	en := &extensionNode{CollapsedEn: protobuf.CollapsedEn{Key: []byte{100, 111, 103}}}
	expected := &extensionNode{CollapsedEn: protobuf.CollapsedEn{Key: []byte{2, 100, 111, 103}}, dirty: true}
	node := en.reduceNode(2)
	assert.Equal(t, expected, node)
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
	assert.Equal(t, ErrEmptyNode, en.isEmptyOrNil())

	en = nil
	assert.Equal(t, ErrNilNode, en.isEmptyOrNil())
}

//------- deepClone

func TestExtensionNode_deepCloneNilHashShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
	en.dirty = true
	en.hash = nil
	en.EncodedChild = getRandomByteSlice()
	en.Key = getRandomByteSlice()
	en.child = &leafNode{}

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func TestExtensionNode_deepCloneNilEncodedChildShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
	en.dirty = true
	en.hash = getRandomByteSlice()
	en.EncodedChild = nil
	en.Key = getRandomByteSlice()
	en.child = &leafNode{}

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func TestExtensionNode_deepCloneNilKeyShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
	en.dirty = true
	en.hash = getRandomByteSlice()
	en.EncodedChild = getRandomByteSlice()
	en.Key = nil
	en.child = &leafNode{}

	cloned := en.deepClone().(*extensionNode)

	testSameExtensionNodeContent(t, en, cloned)
}

func TestExtensionNode_deepCloneNilChildShouldWork(t *testing.T) {
	t.Parallel()

	en := &extensionNode{}
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

	en := &extensionNode{}
	en.dirty = true
	en.hash = getRandomByteSlice()
	en.EncodedChild = getRandomByteSlice()
	en.Key = getRandomByteSlice()
	en.child = &leafNode{}

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
