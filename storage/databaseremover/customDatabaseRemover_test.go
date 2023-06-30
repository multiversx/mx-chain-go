package databaseremover

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/require"
)

func TestCustomDatabaseRemover(t *testing.T) {
	t.Parallel()

	t.Run("empty pattern, should remove everything", func(t *testing.T) {
		t.Parallel()

		cdr, err := NewCustomDatabaseRemover(createCfgWithPattern(""))
		require.NoError(t, err)

		require.True(t, cdr.ShouldRemove("", 0))
	})

	t.Run("empty pattern argument, should error", func(t *testing.T) {
		t.Parallel()

		cdr, err := NewCustomDatabaseRemover(createCfgWithPattern(","))
		require.True(t, errors.Is(err, errEmptyPatternArgument))
		require.Nil(t, cdr)
	})

	t.Run("invalid pattern argument, should error", func(t *testing.T) {
		t.Parallel()

		cdr, err := NewCustomDatabaseRemover(createCfgWithPattern("50,"))
		require.True(t, errors.Is(err, errInvalidPatternArgument))
		require.Nil(t, cdr)
	})

	t.Run("invalid epoch in pattern argument, should error", func(t *testing.T) {
		t.Parallel()

		cdr, err := NewCustomDatabaseRemover(createCfgWithPattern("%ddd,"))
		require.True(t, errors.Is(err, errCannotDecodeEpochNumber))
		require.Nil(t, cdr)
	})

	t.Run("epoch 0 in pattern argument, should error", func(t *testing.T) {
		t.Parallel()

		cdr, err := NewCustomDatabaseRemover(createCfgWithPattern("%0"))
		require.True(t, errors.Is(err, errEpochCannotBeZero))
		require.Nil(t, cdr)
	})

	t.Run("single pattern argument, should work", func(t *testing.T) {
		t.Parallel()

		cdr, err := NewCustomDatabaseRemover(createCfgWithPattern("%2"))
		require.NoError(t, err)
		require.NotNil(t, cdr)

		require.False(t, cdr.ShouldRemove("", 0))
		require.True(t, cdr.ShouldRemove("", 1))
		require.False(t, cdr.ShouldRemove("", 2))
		require.True(t, cdr.ShouldRemove("", 3))
		require.False(t, cdr.ShouldRemove("", 4))
	})

	t.Run("multiple pattern arguments, should work", func(t *testing.T) {
		t.Parallel()

		cdr, err := NewCustomDatabaseRemover(createCfgWithPattern("%2,%3"))
		require.NoError(t, err)
		require.NotNil(t, cdr)

		require.False(t, cdr.ShouldRemove("", 0))
		require.True(t, cdr.ShouldRemove("", 1))
		require.False(t, cdr.ShouldRemove("", 2))
		require.False(t, cdr.ShouldRemove("", 3))
		require.False(t, cdr.ShouldRemove("", 4))
		require.True(t, cdr.ShouldRemove("", 5))
		require.False(t, cdr.ShouldRemove("", 9))
	})
}

func createCfgWithPattern(pattern string) config.StoragePruningConfig {
	return config.StoragePruningConfig{
		AccountsTrieSkipRemovalCustomPattern: pattern,
	}
}

func TestCustomDatabaseRemover_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var cdr *customDatabaseRemover
	require.True(t, cdr.IsInterfaceNil())

	cdr, _ = NewCustomDatabaseRemover(createCfgWithPattern("%2,%3"))
	require.False(t, cdr.IsInterfaceNil())
}
