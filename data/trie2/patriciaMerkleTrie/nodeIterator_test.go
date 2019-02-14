package patriciaMerkleTrie_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/patriciaMerkleTrie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

func testTrie() *patriciaMerkleTrie.PatriciaMerkleTree {
	tr := patriciaMerkleTrie.New(keccak.Keccak{}, marshal.JsonMarshalizer{})

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	return tr
}

func TestNodeIterator_newIterator(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

	assert.NotNil(t, it)
}

func TestNodeIterator_Error(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

	assert.Nil(t, it.Error())
}

func TestNodeIterator_Hash(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

	assert.Equal(t, []byte{}, it.Hash())

	it.Next()

	assert.NotEqual(t, []byte{}, it.Hash())
}

func TestNodeIterator_Parent(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

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
	it := tr.NodeIterator()

	assert.Nil(t, it.Path())

	it.Next()
	it.Next()

	assert.Equal(t, []byte{6, 4, 6, 15, 6}, it.Path())
}

func TestNodeIterator_Leaf(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

	it.Next()
	it.Next()
	assert.False(t, it.Leaf())

	it.Next()
	it.Next()
	assert.True(t, it.Leaf())
}

func TestNodeIterator_LeafKey(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

	searchedKey := []byte("doe")
	var key []byte

	for it.Next() {
		if it.Leaf() {
			key = it.LeafKey()
			break
		}
	}

	assert.Equal(t, searchedKey, key)
}

func TestNodeIterator_LeafBlob(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

	searchedVal := []byte("reindeer")
	var val []byte

	for it.Next() {
		if it.Leaf() {
			val = it.LeafBlob()
			break
		}
	}

	assert.Equal(t, searchedVal, val)
}

func TestNodeIterator_LeafProof(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()
	var proofs [][][]byte
	rootHash := tr.Root()

	for it.Next() {
		if it.Leaf() {
			proofs = append(proofs, it.LeafProof())
		}
	}

	assert.True(t, tr.VerifyProof(rootHash, proofs[0], []byte("doe")))
	assert.True(t, tr.VerifyProof(rootHash, proofs[1], []byte("dogglesworth")))
	assert.True(t, tr.VerifyProof(rootHash, proofs[2], []byte("dog")))
	assert.False(t, tr.VerifyProof(rootHash, proofs[2], []byte("doge")))
}
