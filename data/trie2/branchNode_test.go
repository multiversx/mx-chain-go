package trie2

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func getTestMarshAndHasher() (marshal.Marshalizer, hashing.Hasher) {
	marsh := marshal.JsonMarshalizer{}
	hasher := keccak.Keccak{}
	return marsh, hasher
}

func getBnAndCollapsedBn() (*branchNode, *branchNode) {
	var children [nrOfChildren]node
	var EncodedChildren [nrOfChildren][]byte
	marsh, hasher := getTestMarshAndHasher()
	children[2] = &leafNode{Key: []byte("dog"), Value: []byte("dog"), dirty: true}
	children[6] = &leafNode{Key: []byte("doe"), Value: []byte("doe"), dirty: true}
	children[13] = &leafNode{Key: []byte("doge"), Value: []byte("doge"), dirty: true}
	bn := &branchNode{children: children, dirty: true}

	EncodedChildren[2], _ = encodeNodeAndGetHash(children[2], marsh, hasher)
	EncodedChildren[6], _ = encodeNodeAndGetHash(children[6], marsh, hasher)
	EncodedChildren[13], _ = encodeNodeAndGetHash(children[13], marsh, hasher)
	collapsedBn := &branchNode{EncodedChildren: EncodedChildren}
	return bn, collapsedBn
}

func TestBranchNode_getHash(t *testing.T) {
	t.Parallel()
	bn := &branchNode{hash: []byte("test hash")}
	assert.Equal(t, bn.hash, bn.getHash())
}

func TestBranchNode_isDirty(t *testing.T) {
	t.Parallel()
	bn := &branchNode{dirty: true}
	assert.Equal(t, true, bn.isDirty())

	bn = &branchNode{dirty: false}
	assert.Equal(t, false, bn.isDirty())
}

func TestBranchNode_getCollapsed(t *testing.T) {
	t.Parallel()
	bn, collapsedBn := getBnAndCollapsedBn()
	collapsedBn.dirty = true
	marsh, hasher := getTestMarshAndHasher()

	collapsed, err := bn.getCollapsed(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, collapsedBn, collapsed)
}

func TestBranchNode_getCollapsedEmptyNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	bn := &branchNode{}

	collapsed, err := bn.getCollapsed(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, collapsed)
}

func TestBranchNode_getCollapsedNilNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	var bn *branchNode

	collapsed, err := bn.getCollapsed(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, collapsed)
}

func TestBranchNode_getCollapsedCollapsedNode(t *testing.T) {
	t.Parallel()
	_, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	collapsed, err := collapsedBn.getCollapsed(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, collapsedBn, collapsed)
}

func TestBranchNode_setHash(t *testing.T) {
	t.Parallel()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedBn, marsh, hasher)

	err := bn.setHash(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, hash, bn.hash)
}

func TestBranchNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	bn := &branchNode{}

	err := bn.setHash(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, bn.hash)
}

func TestBranchNode_setHashNilNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	var bn *branchNode

	err := bn.setHash(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, bn)
}

func TestBranchNode_setHashCollapsedNode(t *testing.T) {
	t.Parallel()
	_, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedBn, marsh, hasher)

	err := collapsedBn.setHash(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, hash, collapsedBn.hash)

}

func TestBranchNode_hashChildren(t *testing.T) {
	t.Parallel()
	bn, _ := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	for i := range bn.children {
		if bn.children[i] != nil {
			assert.Nil(t, bn.children[i].getHash())
		}
	}
	err := bn.hashChildren(marsh, hasher)
	assert.Nil(t, err)

	for i := range bn.children {
		if bn.children[i] != nil {
			childHash, _ := encodeNodeAndGetHash(bn.children[i], marsh, hasher)
			assert.Equal(t, childHash, bn.children[i].getHash())
		}
	}
}

func TestBranchNode_hashChildrenEmptyNode(t *testing.T) {
	t.Parallel()
	bn := &branchNode{}
	marsh, hasher := getTestMarshAndHasher()

	err := bn.hashChildren(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_hashChildrenNilNode(t *testing.T) {
	t.Parallel()
	var bn *branchNode
	marsh, hasher := getTestMarshAndHasher()

	err := bn.hashChildren(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
}

func TestBranchNode_hashChildrenCollapsedNode(t *testing.T) {
	t.Parallel()
	_, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	err := collapsedBn.hashChildren(marsh, hasher)
	assert.Nil(t, err)

	_, collapsedBn2 := getBnAndCollapsedBn()
	assert.Equal(t, collapsedBn2, collapsedBn)
}

func TestBranchNode_hashNode(t *testing.T) {
	t.Parallel()
	_, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	expectedHash, _ := encodeNodeAndGetHash(collapsedBn, marsh, hasher)
	hash, err := collapsedBn.hashNode(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, expectedHash, hash)
}

func TestBranchNode_hashNodeEmptyNode(t *testing.T) {
	t.Parallel()
	bn := &branchNode{}
	marsh, hasher := getTestMarshAndHasher()

	hash, err := bn.hashNode(marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, hash)
}

func TestBranchNode_hashNodeNilNode(t *testing.T) {
	t.Parallel()
	var bn *branchNode
	marsh, hasher := getTestMarshAndHasher()

	hash, err := bn.hashNode(marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, hash)
}

func TestBranchNode_commit(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedBn, marsh, hasher)
	bn.setHash(marsh, hasher)

	err := bn.commit(db, marsh, hasher)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	node, _ := decodeNode(encNode, marsh)
	assert.Equal(t, collapsedBn, node)
}

func TestBranchNode_commitEmptyNode(t *testing.T) {
	t.Parallel()
	bn := &branchNode{}
	db, _ := memorydb.New()
	marsh, hasher := getTestMarshAndHasher()

	err := bn.commit(db, marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_commitNilNode(t *testing.T) {
	t.Parallel()
	var bn *branchNode
	db, _ := memorydb.New()
	marsh, hasher := getTestMarshAndHasher()

	err := bn.commit(db, marsh, hasher)
	assert.Equal(t, ErrNilNode, err)
}

func TestBranchNode_getEncodedNode(t *testing.T) {
	t.Parallel()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	expectedEncodedNode, _ := marsh.Marshal(bn)
	expectedEncodedNode = append(expectedEncodedNode, branch)

	encNode, err := bn.getEncodedNode(marsh)
	assert.Nil(t, err)
	assert.Equal(t, expectedEncodedNode, encNode)
}

func TestBranchNode_getEncodedNodeEmpty(t *testing.T) {
	t.Parallel()
	bn := &branchNode{}
	marsh, _ := getTestMarshAndHasher()

	encNode, err := bn.getEncodedNode(marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, encNode)
}

func TestBranchNode_getEncodedNodeNil(t *testing.T) {
	t.Parallel()
	var bn *branchNode
	marsh, _ := getTestMarshAndHasher()

	encNode, err := bn.getEncodedNode(marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, encNode)
}

func TestBranchNode_resolveCollapsed(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	bn.setHash(marsh, hasher)
	bn.commit(db, marsh, hasher)
	resolved := &leafNode{Key: []byte("dog"), Value: []byte("dog")}

	err := collapsedBn.resolveCollapsed(2, db, marsh)
	assert.Nil(t, err)
	assert.Equal(t, resolved, collapsedBn.children[2])
}

func TestBranchNode_resolveCollapsedEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn := &branchNode{}
	marsh, _ := getTestMarshAndHasher()

	err := bn.resolveCollapsed(2, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_resolveCollapsedENilNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	var bn *branchNode
	marsh, _ := getTestMarshAndHasher()

	err := bn.resolveCollapsed(2, db, marsh)
	assert.Equal(t, ErrNilNode, err)
}

func TestBranchNode_resolveCollapsedPosOutOfRange(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	err := bn.resolveCollapsed(17, db, marsh)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_isCollapsed(t *testing.T) {
	t.Parallel()
	bn, collapsedBn := getBnAndCollapsedBn()

	assert.True(t, collapsedBn.isCollapsed())
	assert.False(t, bn.isCollapsed())

	collapsedBn.children[2] = &leafNode{Key: []byte("dog"), Value: []byte("dog"), dirty: true}
	assert.False(t, collapsedBn.isCollapsed())
}

func TestBranchNode_tryGet(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{2, 100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestBranchNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	var key []byte
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetChildPosOutOfRange(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetNilChild(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{3}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrNodeNotFound, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()
	bn.setHash(marsh, hasher)
	bn.commit(db, marsh, hasher)

	key := []byte{2, 100, 111, 103}
	val, err := collapsedBn.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestBranchNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn := &branchNode{}
	marsh, _ := getTestMarshAndHasher()

	key := []byte{2, 100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetNilNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	var bn *branchNode
	marsh, _ := getTestMarshAndHasher()

	key := []byte{2, 100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, val)
}

func TestBranchNode_getNext(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()
	nextNode := &leafNode{Key: []byte("dog"), Value: []byte("dog"), dirty: true}
	key := []byte{2, 100, 111, 103}

	node, key, err := bn.getNext(key, db, marsh)
	assert.Equal(t, nextNode, node)
	assert.Equal(t, []byte{100, 111, 103}, key)
	assert.Nil(t, err)
}

func TestBranchNode_getNextWrongKey(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()
	key := []byte{100, 111, 103}

	node, key, err := bn.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrChildPosOutOfRange, err)
}

func TestBranchNode_getNextNilChild(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()
	key := []byte{4, 100, 111, 103}

	node, key, err := bn.getNext(key, db, marsh)
	assert.Nil(t, node)
	assert.Nil(t, key)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestBranchNode_insert(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	node := &leafNode{Key: []byte{0, 2, 3}, Value: []byte("dogs")}
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	bn.children[0] = &leafNode{Key: []byte{2, 3}, Value: []byte("dogs")}
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_insertEmptyKey(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	node := &leafNode{Key: []byte{}, Value: []byte("dogs")}
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertChildPosOutOfRange(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	node := &leafNode{Key: []byte{100, 111, 103}, Value: []byte("dogs")}
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, collapsedBn := getBnAndCollapsedBn()
	node := &leafNode{Key: []byte{2, 100, 111, 103}, Value: []byte("dogs")}
	marsh, hasher := getTestMarshAndHasher()
	bn.setHash(marsh, hasher)
	bn.commit(db, marsh, hasher)

	dirty, newBn, err := collapsedBn.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newBn.tryGet([]byte{2, 100, 111, 103}, db, marsh)
	assert.Equal(t, []byte("dogs"), val)
}

func TestBranchNode_insertInNilNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	var bn *branchNode
	node := &leafNode{Key: []byte{0, 2, 3}, Value: []byte("dogs")}
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_delete(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	var children [nrOfChildren]node
	children[6] = &leafNode{Key: []byte("doe"), Value: []byte("doe"), dirty: true}
	children[13] = &leafNode{Key: []byte("doge"), Value: []byte("doge"), dirty: true}
	expectedBn := &branchNode{children: children, dirty: true}

	dirty, newBn, err := bn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, expectedBn, newBn)
}

func TestBranchNode_deleteEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn := &branchNode{}
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteNilNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	var bn *branchNode
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteEmptykey(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.delete([]byte{}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()
	bn.setHash(marsh, hasher)
	bn.commit(db, marsh, hasher)

	dirty, newBn, err := collapsedBn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	_, err = newBn.tryGet([]byte{2, 100, 111, 103}, db, marsh)
	assert.Equal(t, ErrNodeNotFound, err)
}

func TestBranchNode_deleteAndReduceBn(t *testing.T) {
	t.Parallel()
	db, _ := memorydb.New()
	marsh, _ := getTestMarshAndHasher()

	var children [nrOfChildren]node
	children[2] = &leafNode{Key: []byte("dog"), Value: []byte("dog"), dirty: true}
	children[6] = &leafNode{Key: []byte("doe"), Value: []byte("doe"), dirty: true}
	bn := &branchNode{children: children, dirty: true}
	ln := &leafNode{Key: []byte{2, 100, 111, 103}, Value: []byte("dog"), dirty: true}

	dirty, newBn, err := bn.delete([]byte{6, 100, 111, 101}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newBn)
}

func TestBranchNode_reduceNode(t *testing.T) {
	t.Parallel()
	var children [nrOfChildren]node
	children[2] = &leafNode{Key: []byte("dog"), Value: []byte("dog"), dirty: true}
	bn := &branchNode{children: children, dirty: true}
	ln := &leafNode{Key: []byte{2, 100, 111, 103}, Value: []byte("dog"), dirty: true}
	node := bn.reduceNode(2)
	assert.Equal(t, ln, node)
}

func TestBranchNode_getChildPosition(t *testing.T) {
	t.Parallel()
	bn, _ := getBnAndCollapsedBn()
	nr, pos := getChildPosition(bn)
	assert.Equal(t, 3, nr)
	assert.Equal(t, 13, pos)
}

func TestBranchNode_clone(t *testing.T) {
	t.Parallel()
	bn, _ := getBnAndCollapsedBn()
	clone := bn.clone()
	assert.False(t, bn == clone)
	assert.Equal(t, bn, clone)
}

func TestBranchNode_isEmptyOrNil(t *testing.T) {
	t.Parallel()
	bn := &branchNode{}
	assert.Equal(t, ErrEmptyNode, bn.isEmptyOrNil())

	bn = nil
	assert.Equal(t, ErrNilNode, bn.isEmptyOrNil())
}
