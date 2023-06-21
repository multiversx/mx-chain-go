package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmptyTrie(t *testing.T) {
	t.Parallel()

	t.Run("test nil root", func(t *testing.T) {
		t.Parallel()

		assert.True(t, IsEmptyTrie(nil))
	})
	t.Run("test empty root", func(t *testing.T) {
		t.Parallel()

		assert.True(t, IsEmptyTrie([]byte{}))
	})
	t.Run("test empty root hash", func(t *testing.T) {
		t.Parallel()

		assert.True(t, IsEmptyTrie(EmptyTrieHash))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		assert.False(t, IsEmptyTrie([]byte("hash")))
	})
}
