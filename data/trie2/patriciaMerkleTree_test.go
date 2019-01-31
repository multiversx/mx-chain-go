package trie2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
)

func TestNewTrieWithNilHasher(t *testing.T) {
	tr := trie2.NewTrie()
	assert.NotNil(t, tr)
}

func TestPatriciaMerkleTree_Insert(t *testing.T) {
	tr := trie2.NewTrie()
	tr2 := trie2.NewTrie()

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	assert.NotEqual(t, tr, tr2)

}

func TestPatriciaMerkleTree_InsertWithEmptyVal(t *testing.T) {
	tr := trie2.NewTrie()

	tr.Update([]byte("doe"), []byte("reindeer"))
	root1 := tr.Root()

	tr.Update([]byte("dog"), []byte("puppy"))
	root2 := tr.Root()

	tr.Update([]byte("dog"), []byte(""))
	root3 := tr.Root()

	assert.Equal(t, root1, root3)
	assert.NotEqual(t, root1, root2)
}

func TestPatriciaMerkleTree_Delete(t *testing.T) {
	tr := trie2.NewTrie()

	tr.Update([]byte("doe"), []byte("reindeer"))
	root1 := tr.Root()

	tr.Update([]byte("dog"), []byte("puppy"))
	root2 := tr.Root()

	tr.Delete([]byte("dog"))
	root3 := tr.Root()

	assert.Equal(t, root1, root3)
	assert.NotEqual(t, root1, root2)
}

func TestPatriciaMerkleTree_Consistency(t *testing.T) {
	tr := trie2.NewTrie()

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))
	tr.Delete([]byte("dogglesworth"))
	root1 := tr.Root()

	tr = trie2.NewTrie()
	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	root2 := tr.Root()

	assert.Equal(t, root2, root1)
}

func TestPatriciaMerkleTree_Root(t *testing.T) {
	tr := trie2.NewTrie()

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	expected := []byte{196, 44, 8, 246, 100, 243, 150, 29, 125, 145, 179, 219, 79, 44, 157,
		9, 159, 63, 183, 158, 117, 199, 97, 28, 203, 125, 149, 223, 21, 233, 225, 113}

	assert.Equal(t, tr.Root(), expected)

}

func TestPatriciaMerkleTree_Copy(t *testing.T) {
	tr := trie2.NewTrie()
	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	tr2 := tr.Copy()

	assert.Equal(t, tr, tr2)
}

func TestPatriciaMerkleTree_Get(t *testing.T) {
	tr := trie2.NewTrie()
	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))

	dog, _ := tr.Get([]byte("dog"))
	doe, _ := tr.Get([]byte("doe"))
	dom, _ := tr.Get([]byte("dom"))

	assert.Equal(t, []byte("puppy"), dog)
	assert.Equal(t, []byte("reindeer"), doe)
	assert.Nil(t, dom)
}
