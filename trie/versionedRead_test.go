package trie_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
)

func TestGetWithVersion_RecreatedTrie(t *testing.T) {
	tr := emptyTrie()
	t.Cleanup(func() { require.NoError(t, tr.GetStorageManager().Close()) })
	keys := []string{"key", "key-extension", "other", "absent"}
	for i, key := range keys[:3] {
		require.NoError(t, tr.(state.DataTrie).UpdateWithVersion([]byte(key), []byte(key), core.TrieNodeVersion(i%2)))
	}
	root, err := tr.RootHash()
	require.NoError(t, err)
	require.NoError(t, tr.Commit())
	tr, err = tr.Recreate(holders.NewDefaultRootHashesHolder(root))
	require.NoError(t, err)
	for i, key := range keys {
		value, depth, version, readErr := tr.GetWithVersion([]byte(key))
		require.NoError(t, readErr)
		legacyValue, legacyDepth, legacyErr := tr.Get([]byte(key))
		require.NoError(t, legacyErr)
		require.Equal(t, legacyValue, value)
		require.Equal(t, legacyDepth, depth)
		if key == "absent" {
			require.Empty(t, value)
		} else {
			require.Equal(t, uint32(i%2), version)
		}
	}
}

func TestGetWithVersion_NoAdditionalAllocation(t *testing.T) {
	tr := emptyTrie()
	t.Cleanup(func() { require.NoError(t, tr.GetStorageManager().Close()) })
	key := []byte("key")
	require.NoError(t, tr.Update(key, []byte("value")))
	ordinary := testing.AllocsPerRun(100, func() {
		_, _, err := tr.Get(key)
		if err != nil {
			panic(err)
		}
	})
	versioned := testing.AllocsPerRun(100, func() {
		_, _, _, err := tr.GetWithVersion(key)
		if err != nil {
			panic(err)
		}
	})
	require.Equal(t, ordinary, versioned)
}

func BenchmarkTrieVersionedRead(b *testing.B) {
	tr := emptyTrie()
	b.Cleanup(func() { require.NoError(b, tr.GetStorageManager().Close()) })
	key := make([]byte, 32)
	require.NoError(b, tr.Update(key, []byte("value")))
	b.Run("Get", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _, err := tr.Get(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GetWithVersion", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _, _, err := tr.GetWithVersion(key)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
