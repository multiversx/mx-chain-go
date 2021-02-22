package trie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewIterator(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	it, err := trie.NewIterator(tr)
	assert.Nil(t, err)
	assert.NotNil(t, it)
}

func TestNewIteratorNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	var tr data.Trie

	it, err := trie.NewIterator(tr)
	assert.Nil(t, it)
	assert.Equal(t, trie.ErrNilTrie, err)
}

func TestIterator_HasNext(t *testing.T) {
	t.Parallel()

	tr := emptyTrie()
	_ = tr.Update([]byte("dog"), []byte("dog"))
	it, _ := trie.NewIterator(tr)
	assert.False(t, it.HasNext())

	_ = tr.Update([]byte("doe"), []byte("doe"))
	it, _ = trie.NewIterator(tr)
	assert.True(t, it.HasNext())
}

func TestIterator_Next(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	it, _ := trie.NewIterator(tr)
	for it.HasNext() {
		err := it.Next()
		assert.Nil(t, err)
	}
}

func TestIterator_GetMarshalizedNode(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	it, _ := trie.NewIterator(tr)

	encNode, err := it.MarshalizedNode()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(encNode))

	branchNodeIdentifier := uint8(2)
	lastByte := len(encNode) - 1
	assert.Equal(t, branchNodeIdentifier, encNode[lastByte])
}

func TestIterator_GetHash(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	rootHash, _ := tr.RootHash()
	it, _ := trie.NewIterator(tr)

	hash, err := it.GetHash()
	assert.Nil(t, err)
	assert.Equal(t, rootHash, hash)
}
