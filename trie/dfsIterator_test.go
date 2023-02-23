package trie_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestDFSIterator_Next(t *testing.T) {
	t.Parallel()

	tr := initTrie()

	it, _ := trie.NewDFSIterator(tr)
	for it.HasNext() {
		err := it.Next()
		assert.Nil(t, err)
	}
}
