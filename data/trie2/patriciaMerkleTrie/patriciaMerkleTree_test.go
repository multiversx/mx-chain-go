package patriciaMerkleTrie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2/patriciaMerkleTrie"

	"github.com/stretchr/testify/assert"
)

func testTrie2(nr int) (*patriciaMerkleTrie.PatriciaMerkleTree, [][]byte) {
	tr := patriciaMerkleTrie.New(keccak.Keccak{}, marshal.JsonMarshalizer{})

	var values [][]byte
	hsh := keccak.Keccak{}

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(string(i)))
		tr.Update(values[i], values[i])
	}

	return tr, values

}

func TestNewTrieWithNilParameters(t *testing.T) {
	tr := patriciaMerkleTrie.New(nil, nil)

	assert.NotNil(t, tr)
}

func TestPatriciaMerkleTree_Get(t *testing.T) {
	tr, val := testTrie2(50)

	for i := range val {
		v, _ := tr.Get(val[i])
		assert.Equal(t, val[i], v)
	}
}

func TestPatriciaMerkleTree_Update(t *testing.T) {
	tr := testTrie()

	newVal := []byte("doge")
	tr.Update([]byte("dog"), newVal)

	val, _ := tr.Get([]byte("dog"))
	assert.Equal(t, newVal, val)
}

func TestPatriciaMerkleTree_UpdateEmptyVal(t *testing.T) {
	tr := testTrie()
	var empty []byte

	tr.Update([]byte("doe"), []byte{})

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_UpdateNotExisting(t *testing.T) {
	tr := testTrie()

	tr.Update([]byte("does"), []byte("this"))

	v, _ := tr.Get([]byte("does"))
	assert.Equal(t, []byte("this"), v)
}

func TestPatriciaMerkleTree_Delete(t *testing.T) {
	tr := testTrie()
	var empty []byte

	tr.Delete([]byte("doe"))

	v, _ := tr.Get([]byte("doe"))
	assert.Equal(t, empty, v)
}

func TestPatriciaMerkleTree_Root(t *testing.T) {
	tr := testTrie()

	assert.NotNil(t, tr.Root())
}

func TestPatriciaMerkleTree_Copy(t *testing.T) {
	tr := testTrie()
	tr2 := tr.Copy()

	assert.Equal(t, tr, tr2)
}

func TestPatriciaMerkleTree_Prove(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()
	var proof1 [][]byte

	for it.Next() {
		if it.Leaf() {
			proof1 = it.LeafProof()
			break
		}
	}

	proof2 := tr.Prove([]byte("doe"))

	assert.Equal(t, proof1, proof2)

}

func TestPatriciaMerkleTree_VerifyProof(t *testing.T) {
	tr, val := testTrie2(50)
	rootHash := tr.Root()

	for i := range val {
		proof := tr.Prove(val[i])
		assert.True(t, tr.VerifyProof(rootHash, proof, val[i]))
		assert.False(t, tr.VerifyProof(rootHash, proof, []byte("dog"+string(i))))
	}

}

func TestPatriciaMerkleTree_NodeIterator(t *testing.T) {
	tr := testTrie()
	it := tr.NodeIterator()

	assert.NotNil(t, it)
}

func TestPatriciaMerkleTree_Consistency(t *testing.T) {
	tr := testTrie()
	root1 := tr.Root()

	tr.Update([]byte("dodge"), []byte("viper"))
	root2 := tr.Root()

	tr.Delete([]byte("dodge"))
	root3 := tr.Root()

	assert.Equal(t, root1, root3)
	assert.NotEqual(t, root1, root2)
}
