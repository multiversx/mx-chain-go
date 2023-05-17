package trie_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewDFSIterator(t *testing.T) {
	t.Parallel()

	t.Run("nil trie should error", func(t *testing.T) {
		t.Parallel()

		it, err := trie.NewDFSIterator(nil)
		assert.Equal(t, trie.ErrNilTrie, err)
		assert.Nil(t, it)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tr := initTrie()

		it, err := trie.NewDFSIterator(tr)
		assert.Nil(t, err)
		assert.NotNil(t, it)
	})
}

func TestDFSIterator_Next(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	it, _ := trie.NewDFSIterator(tr)
	for it.HasNext() {
		err := it.Next()
		assert.Nil(t, err)
	}
}
