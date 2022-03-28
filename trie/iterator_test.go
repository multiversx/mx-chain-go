package trie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewIterator(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	it, err := trie.NewIterator(tr, common.TestPriority)
	assert.Nil(t, err)
	assert.NotNil(t, it)
}

func TestNewIteratorNilTrieShouldErr(t *testing.T) {
	t.Parallel()

	var tr common.Trie

	it, err := trie.NewIterator(tr, common.TestPriority)
	assert.Nil(t, it)
	assert.Equal(t, trie.ErrNilTrie, err)
}

func TestIterator_HasNext(t *testing.T) {
	t.Parallel()

	tr := emptyTrie(t)
	_ = tr.Update([]byte("dog"), []byte("dog"))
	it, _ := trie.NewIterator(tr, common.TestPriority)
	assert.False(t, it.HasNext())

	_ = tr.Update([]byte("doe"), []byte("doe"))
	it, _ = trie.NewIterator(tr, common.TestPriority)
	assert.True(t, it.HasNext())
}

func TestIterator_Next(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)

	it, _ := trie.NewIterator(tr, common.TestPriority)
	for it.HasNext() {
		err := it.Next()
		assert.Nil(t, err)
	}
}

func TestIterator_GetMarshalizedNode(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	it, _ := trie.NewIterator(tr, common.TestPriority)

	encNode, err := it.MarshalizedNode()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(encNode))

	branchNodeIdentifier := uint8(2)
	lastByte := len(encNode) - 1
	assert.Equal(t, branchNodeIdentifier, encNode[lastByte])
}

func TestIterator_GetHash(t *testing.T) {
	t.Parallel()

	tr := initTrie(t)
	rootHash, _ := tr.RootHash()
	it, _ := trie.NewIterator(tr, common.TestPriority)

	hash, err := it.GetHash()
	assert.Nil(t, err)
	assert.Equal(t, rootHash, hash)
}
