package trie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingHashes_Add(t *testing.T) {
	t.Parallel()

	t.Run("new hash should add", func(t *testing.T) {
		t.Parallel()

		missing := newMissingHashes()
		assert.Equal(t, 0, missing.len())
		assert.Equal(t, 0, len(missing.innerMap))

		missing.add("hash")
		assert.Equal(t, 1, missing.len())
		assert.Equal(t, 1, len(missing.innerMap))
	})
	t.Run("double add should not add", func(t *testing.T) {
		t.Parallel()

		missing := newMissingHashes()
		missing.add("hash")
		assert.Equal(t, 1, missing.len())
		assert.Equal(t, 1, len(missing.innerMap))

		missing.add("hash")
		assert.Equal(t, 1, missing.len())
		assert.Equal(t, 1, len(missing.innerMap))
	})
}

func TestMissingHashes_Has(t *testing.T) {
	t.Parallel()

	t.Run("new hash should return false", func(t *testing.T) {
		t.Parallel()

		missing := newMissingHashes()
		assert.False(t, missing.has("hash"))
	})
	t.Run("existing hash should return true", func(t *testing.T) {
		t.Parallel()

		missing := newMissingHashes()
		missing.add("hash")
		assert.True(t, missing.has("hash"))
	})
}

func TestMissingHashes_Remove(t *testing.T) {
	t.Parallel()

	t.Run("non existent hash should not perform remove", func(t *testing.T) {
		t.Parallel()

		missing := newMissingHashes()
		missing.remove("nonexistent hash")
		assert.Equal(t, 0, missing.len())
		assert.Equal(t, 0, len(missing.innerMap))

		missing.add("hash")
		missing.remove("nonexistent hash")
		assert.Equal(t, 1, missing.len())
		assert.Equal(t, 1, len(missing.innerMap))
	})
	t.Run("existing hash should remove", func(t *testing.T) {
		t.Parallel()

		t.Run("single hash", func(t *testing.T) {
			missing := newMissingHashes()
			missing.add("hash")
			missing.remove("hash")

			assert.Equal(t, 0, missing.len())
			assert.Equal(t, 0, len(missing.innerMap))
		})
		t.Run("last hash", func(t *testing.T) {
			missing := newMissingHashes()
			missing.add("hash1")
			missing.add("hash2")
			missing.remove("hash2")

			assert.Equal(t, 1, missing.len())
			assert.Equal(t, 1, len(missing.innerMap))
			assert.True(t, missing.has("hash1"))
			assert.False(t, missing.has("hash2"))
		})
		t.Run("first hash", func(t *testing.T) {
			missing := newMissingHashes()
			missing.add("hash1")
			missing.add("hash2")
			missing.remove("hash1")

			assert.Equal(t, 1, missing.len())
			assert.Equal(t, 1, len(missing.innerMap))
			assert.False(t, missing.has("hash1"))
			assert.True(t, missing.has("hash2"))
		})
		t.Run("middle hash", func(t *testing.T) {
			missing := newMissingHashes()
			missing.add("hash1")
			missing.add("hash2")
			missing.add("hash3")
			missing.remove("hash2")

			assert.Equal(t, 2, missing.len())
			assert.Equal(t, 2, len(missing.innerMap))
			assert.True(t, missing.has("hash1"))
			assert.False(t, missing.has("hash2"))
			assert.True(t, missing.has("hash3"))
		})
	})
}

func TestMissingHashes_GetHashesSliceWithCopy(t *testing.T) {
	t.Parallel()

	missing := newMissingHashes()
	assert.Empty(t, missing.getHashesSliceWithCopy())

	missing.add("hash1")
	assert.Equal(t, []string{"hash1"}, missing.getHashesSliceWithCopy())
	checkDifferentPointers(t, missing.order, missing.getHashesSliceWithCopy())

	missing.add("hash2")
	assert.Equal(t, []string{"hash1", "hash2"}, missing.getHashesSliceWithCopy())
	checkDifferentPointers(t, missing.order, missing.getHashesSliceWithCopy())
}

func checkDifferentPointers(t *testing.T, slice1 []string, slice2 []string) {
	address1 := fmt.Sprintf("%p", slice1)
	address2 := fmt.Sprintf("%p", slice2)

	fmt.Printf("slice 1: %s, slice 2: %s\n", address1, address2)

	assert.NotEqual(t, address1, address2)
}
