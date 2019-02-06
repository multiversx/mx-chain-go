package trie2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testTrie() *PatriciaMerkleTree {
	tr := NewTrie()

	tr.Update([]byte("doe"), []byte("reindeer"))
	tr.Update([]byte("dog"), []byte("puppy"))
	tr.Update([]byte("dogglesworth"), []byte("cat"))

	return tr
}

func TestNewNodeIterator(t *testing.T) {
	tr := testTrie()

	it := newNodeIterator(tr)

	assert.NotNil(t, it)
}
