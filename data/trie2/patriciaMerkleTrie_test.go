package trie2_test

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func testTrie2(nr int) (trie2.Trie, [][]byte) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	var values [][]byte
	hsh := keccak.Keccak{}

	for i := 0; i < nr; i++ {
		values = append(values, hsh.Compute(string(i)))
		tr.Update(values[i], values[i])
	}

	return tr, values

}

func testTrie() trie2.Trie {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	return tr
}

func TestNewTrieWithNilDB(t *testing.T) {
	tr, err := trie2.NewTrie(nil, marshal.JsonMarshalizer{}, keccak.Keccak{})

	assert.Nil(t, tr)
	assert.NotNil(t, err)
}

func TestNewTrieWithNilMarshalizer(t *testing.T) {
	db, _ := memorydb.New()
	tr, err := trie2.NewTrie(db, nil, keccak.Keccak{})

	assert.Nil(t, tr)
	assert.NotNil(t, err)
}

func TestNewTrieWithNilHasher(t *testing.T) {
	db, _ := memorydb.New()
	tr, err := trie2.NewTrie(db, marshal.JsonMarshalizer{}, nil)

	assert.Nil(t, tr)
	assert.NotNil(t, err)
}

func TestPatriciaMerkleTree_Get(t *testing.T) {
	tr, val := testTrie2(10000)

	for i := range val {
		v, _ := tr.Get(val[i])
		assert.Equal(t, val[i], v)
	}
}

func TestPatriciaMerkleTree_GetEmptyTrie(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	val, err := tr.Get([]byte("dog"))
	assert.Equal(t, trie2.ErrNilNode, err)
	assert.Nil(t, val)
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

func TestPatriciaMerkleTree_DeleteEmptyTrie(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	err := tr.Delete([]byte("dog"))
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_Root(t *testing.T) {
	tr := testTrie()

	root, err := tr.Root()
	assert.NotNil(t, root)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_NilRoot(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	root, err := tr.Root()
	assert.Equal(t, trie2.ErrNilNode, err)
	assert.Nil(t, root)
}

func TestPatriciaMerkleTree_Prove(t *testing.T) {
	tr := testTrie()
	it := tr.NewNodeIterator()
	var proof1 [][]byte

	err := it.Next()
	for err == nil {
		if it.Leaf() {
			proof1, _ = it.LeafProof()
			break
		}
		err = it.Next()
	}

	proof2, err := tr.Prove([]byte("doe"))

	assert.Equal(t, proof1, proof2)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_ProveOnEmptyTrie(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	proof, err := tr.Prove([]byte("dog"))
	assert.Nil(t, proof)
	assert.Equal(t, trie2.ErrNilNode, err)
}

func TestPatriciaMerkleTree_VerifyProof(t *testing.T) {
	tr, val := testTrie2(50)

	for i := range val {
		proof, _ := tr.Prove(val[i])

		ok, err := tr.VerifyProof(proof, val[i])
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = tr.VerifyProof(proof, []byte("dog"+strconv.Itoa(i)))
		assert.Nil(t, err)
		assert.False(t, ok)
	}

}

func TestPatriciaMerkleTree_VerifyProofNilProofs(t *testing.T) {
	tr := testTrie()

	ok, err := tr.VerifyProof(nil, []byte("dog"))
	assert.False(t, ok)
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_VerifyProofEmptyProofs(t *testing.T) {
	tr := testTrie()

	ok, err := tr.VerifyProof([][]byte{}, []byte("dog"))
	assert.False(t, ok)
	assert.Nil(t, err)
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
	tr := testTrie()

	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitAfterCommit(t *testing.T) {
	tr := testTrie()

	tr.Commit()
	err := tr.Commit()
	assert.Nil(t, err)
}

func TestPatriciaMerkleTree_CommitEmptyRoot(t *testing.T) {
	db, _ := memorydb.New()
	tr, _ := trie2.NewTrie(db, marshal.JsonMarshalizer{}, keccak.Keccak{})

	err := tr.Commit()
	assert.Equal(t, trie2.ErrNilNode, err)
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
