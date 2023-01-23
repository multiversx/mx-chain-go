package holders

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockInfo(t *testing.T) {
	t.Parallel()

	t.Run("nil and default values", func(t *testing.T) {
		t.Parallel()

		bi := NewBlockInfo(nil, 0, nil)

		assert.False(t, check.IfNil(bi))
		assert.Nil(t, bi.GetHash())
		assert.Nil(t, bi.GetRootHash())
		assert.Zero(t, bi.GetNonce())
	})
	t.Run("some values", func(t *testing.T) {
		t.Parallel()

		hash := []byte("hash")
		rootHash := []byte("root hash")
		nonce := uint64(6658)
		bi := NewBlockInfo(hash, nonce, rootHash)

		assert.False(t, check.IfNil(bi))
		assert.Equal(t, hash, bi.GetHash())
		assert.Equal(t, rootHash, bi.GetRootHash())
		assert.Equal(t, nonce, bi.GetNonce())
	})
}

func TestBlockInfo_Equal(t *testing.T) {
	t.Parallel()

	t.Run("with nil", func(t *testing.T) {
		t.Parallel()

		bi := NewBlockInfo(nil, 0, nil)

		assert.False(t, bi.Equal(nil))
	})
	t.Run("different hash", func(t *testing.T) {
		t.Parallel()

		bi1 := NewBlockInfo([]byte("hash1"), 1, []byte("root hash"))
		bi2 := NewBlockInfo([]byte("hash2"), 1, []byte("root hash"))

		assert.False(t, bi1.Equal(bi2))
	})
	t.Run("different nonce", func(t *testing.T) {
		t.Parallel()

		bi1 := NewBlockInfo([]byte("hash"), 0, []byte("root hash"))
		bi2 := NewBlockInfo([]byte("hash"), 1, []byte("root hash"))

		assert.False(t, bi1.Equal(bi2))
	})
	t.Run("different root hash", func(t *testing.T) {
		t.Parallel()

		bi1 := NewBlockInfo([]byte("hash"), 1, []byte("root hash1"))
		bi2 := NewBlockInfo([]byte("hash"), 1, []byte("root hash2"))

		assert.False(t, bi1.Equal(bi2))
	})
	t.Run("equal with default values", func(t *testing.T) {
		t.Parallel()

		bi1 := NewBlockInfo(nil, 0, nil)
		bi2 := NewBlockInfo(nil, 0, nil)

		assert.True(t, bi1.Equal(bi2))
	})
	t.Run("equal values", func(t *testing.T) {
		t.Parallel()

		bi1 := NewBlockInfo([]byte("hash"), 1, []byte("root hash"))
		bi2 := NewBlockInfo([]byte("hash"), 1, []byte("root hash"))

		assert.True(t, bi1.Equal(bi2))
	})
}
