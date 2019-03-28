package trie2_test

import (
	"testing"

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
	var err error

	ok, _ := it.Next()

	for ok {
		if it.Leaf() {
			key, err = it.LeafKey()
			break
		}
		ok, _ = it.Next()
	}

	assert.Equal(t, searchedKey, key)
	assert.Nil(t, err)
}

func TestNodeIterator_LeafBlob(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	searchedVal := []byte("reindeer")
	var val []byte
	var err error

	ok, _ := it.Next()

	for ok {
		if it.Leaf() {
			val, err = it.LeafBlob()
			break
		}
		ok, _ = it.Next()
	}

	assert.Equal(t, searchedVal, val)
	assert.Nil(t, err)
}

func TestNodeIterator_LeafProof(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	var proofs [][][]byte

	ok, _ := it.Next()

	for ok {
		if it.Leaf() {
			proof, err := it.LeafProof()
			assert.Nil(t, err)
			proofs = append(proofs, proof)
		}
		ok, _ = it.Next()
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

func TestNodeIterator_Next(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	ok, err := it.Next()
	for ok {
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = it.Next()
	}
	assert.False(t, ok)
	assert.NotNil(t, err)
}
