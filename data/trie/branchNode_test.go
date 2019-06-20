package trie

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/stretchr/testify/assert"
)

func getTestMarshAndHasher() (marshal.Marshalizer, hashing.Hasher) {
	marsh := &mock.ProtobufMarshalizerMock{}
	hasher := mock.KeccakMock{}
	return marsh, hasher
}

func getBnAndCollapsedBn() (*branchNode, *branchNode) {
	marsh, hasher := getTestMarshAndHasher()

	var children [nrOfChildren]node
	EncodedChildren := make([][]byte, nrOfChildren)

	children[2] = newLeafNode([]byte("dog"), []byte("dog"))
	children[6] = newLeafNode([]byte("doe"), []byte("doe"))
	children[13] = newLeafNode([]byte("doge"), []byte("doge"))
	bn := newBranchNode()
	bn.children = children

	EncodedChildren[2], _ = encodeNodeAndGetHash(children[2], marsh, hasher)
	EncodedChildren[6], _ = encodeNodeAndGetHash(children[6], marsh, hasher)
	EncodedChildren[13], _ = encodeNodeAndGetHash(children[13], marsh, hasher)
	collapsedBn := newBranchNode()
	collapsedBn.EncodedChildren = EncodedChildren

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
	bn := newBranchNode()

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

func TestBranchNode_setRootHash(t *testing.T) {
	db, _ := mock.NewMemDbMock()
	marsh, hsh := getTestMarshAndHasher()

	tr1, _ := NewTrie(db, marsh, hsh)
	tr2, _ := NewTrie(db, marsh, hsh)

	for i := 0; i < 100000; i++ {
		val := hsh.Compute(string(i))
		tr1.Update(val, val)
		tr2.Update(val, val)
	}

	err := tr1.root.setRootHash(marsh, hsh)
	tr2.root.setHash(marsh, hsh)
	assert.Nil(t, err)
	assert.Equal(t, tr1.root.getHash(), tr2.root.getHash())
}

func TestBranchNode_setRootHashCollapsedNode(t *testing.T) {
	t.Parallel()
	_, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedBn, marsh, hasher)

	err := collapsedBn.setRootHash(marsh, hasher)
	assert.Nil(t, err)
	assert.Equal(t, hash, collapsedBn.hash)
}

func TestBranchNode_setHashEmptyNode(t *testing.T) {
	t.Parallel()
	marsh, hasher := getTestMarshAndHasher()
	bn := newBranchNode()

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
	bn := newBranchNode()
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
	bn := newBranchNode()
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
	db, _ := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	hash, _ := encodeNodeAndGetHash(collapsedBn, marsh, hasher)
	bn.setHash(marsh, hasher)

	err := bn.commit(0, db, marsh, hasher)
	assert.Nil(t, err)

	encNode, _ := db.Get(hash)
	node, _ := decodeNode(encNode, marsh)
	h1, _ := encodeNodeAndGetHash(collapsedBn, marsh, hasher)
	h2, _ := encodeNodeAndGetHash(node, marsh, hasher)
	assert.Equal(t, h1, h2)
}

func TestBranchNode_commitEmptyNode(t *testing.T) {
	t.Parallel()
	bn := newBranchNode()
	db, _ := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()

	err := bn.commit(0, db, marsh, hasher)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_commitNilNode(t *testing.T) {
	t.Parallel()
	var bn *branchNode
	db, _ := mock.NewMemDbMock()
	marsh, hasher := getTestMarshAndHasher()

	err := bn.commit(0, db, marsh, hasher)
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
	bn := newBranchNode()
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
	db, _ := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	bn.setHash(marsh, hasher)
	bn.commit(0, db, marsh, hasher)
	resolved := newLeafNode([]byte("dog"), []byte("dog"))
	resolved.dirty = false

	err := collapsedBn.resolveCollapsed(2, db, marsh)
	assert.Nil(t, err)
	assert.Equal(t, resolved, collapsedBn.children[2])
}

func TestBranchNode_resolveCollapsedEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn := newBranchNode()
	marsh, _ := getTestMarshAndHasher()

	err := bn.resolveCollapsed(2, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
}

func TestBranchNode_resolveCollapsedENilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var bn *branchNode
	marsh, _ := getTestMarshAndHasher()

	err := bn.resolveCollapsed(2, db, marsh)
	assert.Equal(t, ErrNilNode, err)
}

func TestBranchNode_resolveCollapsedPosOutOfRange(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
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

	collapsedBn.children[2] = newLeafNode([]byte("dog"), []byte("dog"))
	assert.False(t, collapsedBn.isCollapsed())
}

func TestBranchNode_tryGet(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{2, 100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestBranchNode_tryGetEmptyKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	var key []byte
	val, err := bn.tryGet(key, db, marsh)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetChildPosOutOfRange(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetNilChild(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{3}
	val, err := bn.tryGet(key, db, marsh)
	assert.Nil(t, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()
	bn.setHash(marsh, hasher)
	bn.commit(0, db, marsh, hasher)

	key := []byte{2, 100, 111, 103}
	val, err := collapsedBn.tryGet(key, db, marsh)
	assert.Equal(t, []byte("dog"), val)
	assert.Nil(t, err)
}

func TestBranchNode_tryGetEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn := newBranchNode()
	marsh, _ := getTestMarshAndHasher()

	key := []byte{2, 100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, val)
}

func TestBranchNode_tryGetNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var bn *branchNode
	marsh, _ := getTestMarshAndHasher()

	key := []byte{2, 100, 111, 103}
	val, err := bn.tryGet(key, db, marsh)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, val)
}

func TestBranchNode_getNext(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()
	nextNode := newLeafNode([]byte("dog"), []byte("dog"))
	key := []byte{2, 100, 111, 103}

	node, key, err := bn.getNext(key, db, marsh)

	h1, _ := encodeNodeAndGetHash(nextNode, marsh, hasher)
	h2, _ := encodeNodeAndGetHash(node, marsh, hasher)

	assert.Equal(t, h1, h2)
	assert.Equal(t, []byte{100, 111, 103}, key)
	assert.Nil(t, err)
}

func TestBranchNode_getNextWrongKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
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
	db, _ := mock.NewMemDbMock()
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
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	node := newLeafNode([]byte{0, 2, 3}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	bn.children[0] = newLeafNode([]byte{2, 3}, []byte("dogs"))
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, bn, newBn)
}

func TestBranchNode_insertEmptyKey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	node := newLeafNode([]byte{}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertChildPosOutOfRange(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	node := newLeafNode([]byte{100, 111, 103}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrChildPosOutOfRange, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_insertCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn()
	node := newLeafNode([]byte{2, 100, 111, 103}, []byte("dogs"))
	marsh, hasher := getTestMarshAndHasher()
	bn.setHash(marsh, hasher)
	bn.commit(0, db, marsh, hasher)

	dirty, newBn, err := collapsedBn.insert(node, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	val, _ := newBn.tryGet([]byte{2, 100, 111, 103}, db, marsh)
	assert.Equal(t, []byte("dogs"), val)
}

func TestBranchNode_insertInNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var bn *branchNode
	node := newLeafNode([]byte{0, 2, 3}, []byte("dogs"))
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.insert(node, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_delete(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()

	var children [nrOfChildren]node
	children[6] = newLeafNode([]byte("doe"), []byte("doe"))
	children[13] = newLeafNode([]byte("doge"), []byte("doge"))
	expectedBn := newBranchNode()
	expectedBn.children = children

	dirty, newBn, err := bn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)

	expectedBn.setHash(marsh, hasher)
	newBn.setHash(marsh, hasher)
	assert.Equal(t, expectedBn.getHash(), newBn.getHash())
}

func TestBranchNode_deleteEmptyNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn := newBranchNode()
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrEmptyNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteNilNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	var bn *branchNode
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrNilNode, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteEmptykey(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, _ := getBnAndCollapsedBn()
	marsh, _ := getTestMarshAndHasher()

	dirty, newBn, err := bn.delete([]byte{}, db, marsh)
	assert.False(t, dirty)
	assert.Equal(t, ErrValueTooShort, err)
	assert.Nil(t, newBn)
}

func TestBranchNode_deleteCollapsedNode(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	bn, collapsedBn := getBnAndCollapsedBn()
	marsh, hasher := getTestMarshAndHasher()
	bn.setHash(marsh, hasher)
	bn.commit(0, db, marsh, hasher)

	dirty, newBn, err := collapsedBn.delete([]byte{2, 100, 111, 103}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)

	val, err := newBn.tryGet([]byte{2, 100, 111, 103}, db, marsh)
	assert.Nil(t, val)
	assert.Nil(t, err)
}

func TestBranchNode_deleteAndReduceBn(t *testing.T) {
	t.Parallel()
	db, _ := mock.NewMemDbMock()
	marsh, _ := getTestMarshAndHasher()

	var children [nrOfChildren]node
	children[2] = newLeafNode([]byte("dog"), []byte("dog"))
	children[6] = newLeafNode([]byte("doe"), []byte("doe"))
	bn := newBranchNode()
	bn.children = children
	ln := newLeafNode([]byte{2, 100, 111, 103}, []byte("dog"))

	dirty, newBn, err := bn.delete([]byte{6, 100, 111, 101}, db, marsh)
	assert.True(t, dirty)
	assert.Nil(t, err)
	assert.Equal(t, ln, newBn)
}

func TestBranchNode_reduceNode(t *testing.T) {
	t.Parallel()
	var children [nrOfChildren]node
	children[2] = newLeafNode([]byte("dog"), []byte("dog"))
	bn := newBranchNode()
	bn.children = children
	ln := newLeafNode([]byte{2, 100, 111, 103}, []byte("dog"))
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
	bn := newBranchNode()
	assert.Equal(t, ErrEmptyNode, bn.isEmptyOrNil())

	bn = nil
	assert.Equal(t, ErrNilNode, bn.isEmptyOrNil())
}

func newEmptyTrie() Trie {
	db, _ := mock.NewMemDbMock()
	marsh, hsh := getTestMarshAndHasher()
	tr, _ := NewTrie(db, marsh, hsh)
	return tr
}

func BenchmarkDecodeBranchNode(b *testing.B) {
	tr := newEmptyTrie()
	marsh, hsh := getTestMarshAndHasher()

	nrValuesInTrie := 100000
	values := make([][]byte, nrValuesInTrie)

	for i := 0; i < nrValuesInTrie; i++ {
		values[i] = hsh.Compute(strconv.Itoa(i))
		tr.Update(values[i], values[i])
	}

	proof, _ := tr.Prove(values[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeNode(proof[0], marsh)
	}
}

func BenchmarkMarshallNodeCapnp(b *testing.B) {
	bn, _ := getBnAndCollapsedBn()
	marsh := &marshal.CapnpMarshalizer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marsh.Marshal(bn)
	}
}

func BenchmarkMarshallNodeJson(b *testing.B) {
	bn, _ := getBnAndCollapsedBn()
	marsh := marshal.JsonMarshalizer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marsh.Marshal(bn)
	}
}
