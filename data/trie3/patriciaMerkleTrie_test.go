package trie3_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie3"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/stretchr/testify/assert"
)

func testTrie2(nr int) (trie3.Trie, [][]byte) {
	db, _ := memorydb.New()
	tr, _ := trie3.NewTrie(keccak.Keccak{}, marshal.JsonMarshalizer{}, db)

	var values [][]byte
	hsh := keccak.Keccak{}

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(string(i)))
		tr.Update(values[i], values[i])
	}

	return tr, values

}

func testTrie() trie3.Trie {
	db, _ := memorydb.New()
	tr, _ := trie3.NewTrie(keccak.Keccak{}, marshal.JsonMarshalizer{}, db)

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	return tr
}

func TestNewTrieWithNilParameters(t *testing.T) {
	tr, err := trie3.NewTrie(nil, nil, nil)

	assert.Nil(t, tr)
	assert.NotNil(t, err)
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

	root, err := tr.Root()
	assert.NotNil(t, root)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_Prove(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()
	var proof1 [][]byte

	ok, _ := it.Next()
	for ok {
		if it.Leaf() {
			proof1, _ = it.LeafProof()
			break
		}
		ok, _ = it.Next()
	}

	proof2, err := tr.Prove([]byte("doe"))

	assert.Equal(t, proof1, proof2)
	assert.Nil(t, err)

}

func TestPatriciaMerkleTree_VerifyProof(t *testing.T) {
	tr, val := testTrie2(50)

	for i := range val {
		proof, _ := tr.Prove(val[i])

		ok, err := tr.VerifyProof(proof, val[i])
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = tr.VerifyProof(proof, []byte("dog"+string(i)))
		assert.Nil(t, err)
		assert.False(t, ok)
	}

}

func TestPatriciaMerkleTree_NodeIterator(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()

	assert.NotNil(t, it)
}

func TestPatriciaMerkleTree_Consistency(t *testing.T) {
	tr := testTrie()
	root1, _ := tr.Root()

	tr.Update([]byte("dodge"), []byte("viper"))
	root2, _ := tr.Root()

	tr.Delete([]byte("dodge"))
	root3, _ := tr.Root()

	assert.Equal(t, root1, root3)
	assert.NotEqual(t, root1, root2)
}

func TestPatriciaMerkleTree_Commit(t *testing.T) {
	db, err := memorydb.New()
	assert.Nil(t, err)
	assert.NotNil(t, db)
	tr := testTrie()

	err = tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_GetAfterCommit(t *testing.T) {
	tr := testTrie()

	err := tr.Commit()
	assert.Nil(t, err)

	val, err := tr.Get([]byte("dog"))
	assert.Equal(t, []byte("puppy"), val)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_InsertAfterCommit(t *testing.T) {
	tr1 := testTrie()
	tr2 := testTrie()

	err := tr1.Commit()
	assert.Nil(t, err)

	tr1.Update([]byte("doge"), []byte("coin"))
	tr2.Update([]byte("doge"), []byte("coin"))

	root1, _ := tr1.Root()
	root2, _ := tr2.Root()

	assert.Equal(t, root2, root1)

}

func TestPatriciaMerkleTree_DeleteAfterCommit(t *testing.T) {
	tr1 := testTrie()
	tr2 := testTrie()

	err := tr1.Commit()
	assert.Nil(t, err)

	tr1.Delete([]byte("dogglesworth"))
	tr2.Delete([]byte("dogglesworth"))

	root1, _ := tr1.Root()
	root2, _ := tr2.Root()

	assert.Equal(t, root2, root1)
}
