package trie2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie2"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
)

var hasher = keccak.Keccak{}

func TestNewTrieWithNilHasher(t *testing.T) {
	tr1 := trie2.NewTrie(hasher)
	tr2 := trie2.NewTrie(nil)

	assert.Equal(t, tr1, tr2)
}

func TestPatriciaMerkleTree_Insert(t *testing.T) {
	tr := trie2.NewTrie(keccak.Keccak{})
	tr2 := trie2.NewTrie(keccak.Keccak{})

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	assert.NotEqual(t, tr, tr2)

}

func TestPatriciaMerkleTree_Root(t *testing.T) {
	tr := trie2.NewTrie(keccak.Keccak{})

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	expected := []byte{196, 44, 8, 246, 100, 243, 150, 29, 125, 145, 179, 219, 79, 44, 157,
		9, 159, 63, 183, 158, 117, 199, 97, 28, 203, 125, 149, 223, 21, 233, 225, 113}

	assert.Equal(t, tr.Root(), expected)

}
