package trie2_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"

	"github.com/stretchr/testify/assert"
)

func TestNodeIterator_newIterator(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	assert.NotNil(t, it)
	assert.Equal(t, []byte{}, it.Hash())
	assert.Nil(t, it.Path())
}

func TestNodeIterator_Hash(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	it.Next()

	assert.NotEqual(t, []byte{}, it.Hash())
}

func TestNodeIterator_Parent(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	rootParent := it.Hash()

	it.Next()
	parent := it.Hash()

	it.Next()
	actual := it.Parent()

	assert.Equal(t, parent, actual)
	assert.Equal(t, []byte{}, rootParent)
}

func TestNodeIterator_Path(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	it.Next()
	it.Next()

	assert.Equal(t, []byte{6, 4, 6, 15, 6}, it.Path())
}

func TestNodeIterator_Leaf(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	it.Next()
	it.Next()
	assert.False(t, it.Leaf())

	it.Next()
	assert.True(t, it.Leaf())
}

func TestNodeIterator_LeafKey(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	searchedKey := []byte("doe")
	var key []byte

	err := it.Next()
	for err == nil {
		if it.Leaf() {
			key, err = it.LeafKey()
			break
		}
		err = it.Next()
	}

	assert.Equal(t, searchedKey, key)
	assert.Nil(t, err)
}

func TestNodeIterator_LeafKeyNotAtLeaf(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	key, err := it.LeafKey()
	assert.Nil(t, key)
	assert.Equal(t, trie2.ErrNotAtLeaf, err)
}

func TestNodeIterator_LeafBlob(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	searchedVal := []byte("reindeer")
	var val []byte

	err := it.Next()
	for err == nil {
		if it.Leaf() {
			val, err = it.LeafBlob()
			break
		}
		err = it.Next()
	}

	assert.Equal(t, searchedVal, val)
	assert.Nil(t, err)
}

func TestNodeIterator_LeafBlobNotAtLeaf(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	val, err := it.LeafBlob()
	assert.Nil(t, val)
	assert.Equal(t, trie2.ErrNotAtLeaf, err)
}

func TestNodeIterator_LeafProof(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	var proofs [][][]byte

	err := it.Next()
	for err == nil {
		if it.Leaf() {
			proof, err := it.LeafProof()
			assert.Nil(t, err)
			proofs = append(proofs, proof)
		}
		err = it.Next()
	}

	assert.NotNil(t, proofs)
	if proofs != nil {
		ok, err := tr.VerifyProof(proofs[0], []byte("doe"))
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = tr.VerifyProof(proofs[1], []byte("dogglesworth"))
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = tr.VerifyProof(proofs[2], []byte("dog"))
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = tr.VerifyProof(proofs[2], []byte("doge"))
		assert.Nil(t, err)
		assert.False(t, ok)
	}

}

func TestNodeIterator_LeafProofNilRoot(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	it := tr.NewNodeIterator()

	proof, err := it.LeafProof()
	assert.Nil(t, proof)
	assert.Equal(t, trie2.ErrNilNode, err)
}

func TestNodeIterator_LeafProofNotAtLeaf(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	proof, err := it.LeafProof()
	assert.Nil(t, proof)
	assert.Equal(t, trie2.ErrNotAtLeaf, err)
}

func TestNodeIterator_Next(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	err := it.Next()
	for err == nil {
		assert.Nil(t, err)
		err = it.Next()
	}
	assert.NotNil(t, err)
}

func TestNodeIterator_NextEmptyTrie(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	it := tr.NewNodeIterator()

	err := it.Next()
	assert.Equal(t, trie2.ErrNilNode, err)
}
