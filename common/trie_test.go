package common

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestTrieLeafHolder(t *testing.T) {
	t.Parallel()

	value := []byte("value")
	depth := uint32(2)
	version := core.NotSpecified

	tlh := NewTrieLeafHolder(value, depth, version)
	require.NotNil(t, tlh)

	require.Equal(t, value, tlh.Value())
	require.Equal(t, depth, tlh.Depth())
	require.Equal(t, version, tlh.Version())
}

func TestTrimSuffixFromValue(t *testing.T) {
	t.Parallel()

	val := []byte("value")
	ret, err := TrimSuffixFromValue(val, 0)
	require.Nil(t, err)
	require.Equal(t, val, ret)

	val = []byte("value")
	ret, err = TrimSuffixFromValue(val, len(val))
	require.Nil(t, err)
	require.Equal(t, []byte{}, ret)

	ret, err = TrimSuffixFromValue(val, len(val)+1)
	require.Equal(t, core.ErrSuffixNotPresentOrInIncorrectPosition, err)
	require.Nil(t, ret)

	val = []byte("value")
	suffix := []byte("_suffix")
	val = append(val, suffix...)
	ret, err = TrimSuffixFromValue(val, len(suffix))
	require.Nil(t, err)
	require.Equal(t, []byte("value"), ret)
}
