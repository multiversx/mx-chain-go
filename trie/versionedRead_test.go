package trie_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
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
			require.Equal(t, []byte(key), value)
			require.Equal(t, uint32(i%2), version)
		}
	}
}

func TestGetWithVersion_CollapseAndReload(t *testing.T) {
	for _, keys := range [][]string{{"key", "key-extension"}, {"key", "other"}} {
		t.Run(keys[1], func(t *testing.T) {
			const maxSize = 1024 * 1024
			storage, marshaller, hasher, flags, _ := getDefaultTrieParameters()
			t.Cleanup(func() { require.NoError(t, storage.Close()) })
			manager, err := collapseManager.NewCollapseManager(maxSize, 100)
			require.NoError(t, err)
			tr, err := trie.NewTrie(storage, marshaller, hasher, flags, manager)
			require.NoError(t, err)

			for cycle := 1; cycle <= 2; cycle++ {
				values := make([][]byte, len(keys))
				for i, key := range keys {
					values[i] = bytes.Repeat([]byte{byte(cycle + i)}, maxSize/2)
					require.NoError(t, tr.UpdateWithVersion([]byte(key), values[i], core.TrieNodeVersion(i)))
				}
				root, rootErr := tr.RootHash()
				require.NoError(t, rootErr)
				require.Greater(t, tr.SizeInMemory(), maxSize)
				require.NoError(t, tr.Commit())
				require.Less(t, tr.SizeInMemory(), maxSize)

				for _, recreate := range []bool{false, true} {
					var reader common.Trie = tr
					if recreate {
						reader, err = tr.Recreate(holders.NewDefaultRootHashesHolder(root))
						require.NoError(t, err)
					}
					for i, key := range keys {
						value, depth, version, readErr := reader.GetWithVersion([]byte(key))
						require.NoError(t, readErr)
						require.Equal(t, values[i], value)
						require.Equal(t, uint32(i), version)
						require.Greater(t, depth, uint32(0))
					}
				}
			}
		})
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
