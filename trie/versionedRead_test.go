package trie_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/hashesCollector"
)

func TestGetWithVersion_RecreatedTrie(t *testing.T) {
	tr := emptyTrie()
	t.Cleanup(func() { require.NoError(t, tr.GetStorageManager().Close()) })
	keys := []string{"key", "key-extension", "other", "absent"}
	for i, key := range keys[:3] {
		tr.(state.DataTrie).UpdateWithVersion([]byte(key), []byte(key), core.TrieNodeVersion(i%2))
	}
	root, err := tr.RootHash()
	require.NoError(t, err)
	require.NoError(t, tr.Commit(hashesCollector.NewDisabledHashesCollector()))
	tr, err = tr.Recreate(holders.NewDefaultRootHashesHolder(root), "test trie")
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
	tr.Update(key, []byte("value"))
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
	tr.Update(key, []byte("value"))
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

func TestGetWithVersion_PendingBatchChanges(t *testing.T) {
	tr := emptyTrie()
	t.Cleanup(func() { require.NoError(t, tr.Close()) })
	key := []byte("key")
	dataTrie := tr.(state.DataTrie)
	checkRead := func(wantValue []byte, wantVersion uint32) {
		t.Helper()
		value, _, version, err := tr.GetWithVersion(key)
		require.NoError(t, err)
		require.Equal(t, wantValue, value)
		require.Equal(t, wantVersion, version)
	}
	dataTrie.UpdateWithVersion(key, []byte("first"), core.AutoBalanceEnabled)
	checkRead([]byte("first"), uint32(core.AutoBalanceEnabled))
	_, err := tr.RootHash()
	require.NoError(t, err)
	checkRead([]byte("first"), uint32(core.AutoBalanceEnabled))
	dataTrie.UpdateWithVersion(key, []byte("second"), core.NotSpecified)
	checkRead([]byte("second"), uint32(core.NotSpecified))
	tr.Delete(key)
	checkRead(nil, 0)
	require.NoError(t, tr.Commit(hashesCollector.NewDisabledHashesCollector()))
	checkRead(nil, 0)
}

func TestCommitCollectsObsoleteHashesAfterLastDeletion(t *testing.T) {
	for _, reinsert := range []bool{false, true} {
		name := "empty trie"
		if reinsert {
			name = "insert after emptying trie"
		}
		t.Run(name, func(t *testing.T) {
			tr := emptyTrie()
			t.Cleanup(func() { require.NoError(t, tr.Close()) })
			key := []byte("original")
			tr.Update(key, []byte("value"))
			require.NoError(t, tr.Commit(hashesCollector.NewDisabledHashesCollector()))
			originalRoot, err := tr.RootHash()
			require.NoError(t, err)
			tr.Delete(key)
			_, err = tr.RootHash()
			require.NoError(t, err)
			if reinsert {
				tr.Update([]byte("replacement"), []byte("new value"))
			}
			collector, err := hashesCollector.NewHashesCollector(hashesCollector.NewDataTrieHashesCollector())
			require.NoError(t, err)
			require.NoError(t, tr.Commit(collector))
			oldRoot, obsolete, _ := collector.GetCollectedData()
			require.Equal(t, originalRoot, oldRoot)
			require.Contains(t, obsolete, string(originalRoot))
			collector.Clean()
			require.NoError(t, tr.Commit(collector))
			_, obsolete, _ = collector.GetCollectedData()
			require.Empty(t, obsolete, "obsolete hashes must only be collected once")
		})
	}
}
