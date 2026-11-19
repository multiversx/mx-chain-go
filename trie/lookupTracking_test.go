package trie

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	"github.com/stretchr/testify/require"
)

const lookupTrackingBudget = 1 << 20

func newLookupTrackingTrie(t testing.TB, version core.TrieNodeVersion) (*patriciaMerkleTrie, []byte) {
	t.Helper()
	args := GetDefaultTrieStorageManagerParameters()
	storage, err := NewTrieStorageManager(args)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })
	manager, err := collapseManager.NewCollapseManager(lookupTrackingBudget, 100000)
	require.NoError(t, err)
	tr, err := NewTrie(storage, args.Marshalizer, args.Hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, manager)
	require.NoError(t, err)
	for _, key := range []string{"dog", "doe", "ddog"} {
		require.NoError(t, tr.UpdateWithVersion([]byte(key), []byte("value-"+key), version))
	}
	root, err := tr.RootHash()
	require.NoError(t, err)
	require.NoError(t, tr.Commit())
	recreated, err := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
	require.NoError(t, err)
	return recreated.(*patriciaMerkleTrie), root
}

func drainTrackedLookupKeys(t *testing.T, manager common.TrieCollapseManager) [][]byte {
	t.Helper()
	// Use the existing eviction API to observe the real tracking structures.
	// Temporarily cross the threshold, drain candidates without collapsing any
	// nodes, then restore the decoded-node byte count.
	manager.AddSizeInMemory(lookupTrackingBudget)
	keys, err := manager.GetCollapsibleLeaves()
	manager.AddSizeInMemory(-lookupTrackingBudget)
	require.NoError(t, err)
	return keys
}

func loadedLookupNodeSize(n node) int {
	if n == nil {
		return 0
	}
	size := n.sizeInBytes()
	switch current := n.(type) {
	case *branchNode:
		for _, child := range current.children {
			size += loadedLookupNodeSize(child)
		}
	case *extensionNode:
		size += loadedLookupNodeSize(current.child)
	}
	return size
}

func TestTrieLookupTracking_SustainedMissesDoNotRetainKeys(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		for _, versioned := range []bool{false, true} {
			t.Run(fmt.Sprintf("version-%d/versioned-read-%t", version, versioned), func(t *testing.T) {
				tr, root := newLookupTrackingTrie(t, version)
				read := func(key []byte) ([]byte, uint32, error) {
					if versioned {
						value, depth, _, err := tr.GetWithVersion(key)
						return value, depth, err
					}
					return tr.Get(key)
				}
				for cycle := 0; cycle < 2; cycle++ {
					for _, key := range []string{"dog", "doe", "ddog", "dog"} {
						value, _, err := read([]byte(key))
						require.NoError(t, err)
						require.Equal(t, []byte("value-"+key), value)
					}
					warmSize := tr.SizeInMemory()
					for batch := 0; batch < 2; batch++ {
						for i := 0; i < 10000; i++ {
							var key [32]byte
							binary.BigEndian.PutUint64(key[:8], uint64(cycle*20000+batch*10000+i))
							value, _, err := read(key[:])
							require.NoError(t, err)
							require.Empty(t, value)
						}
						require.Equal(t, warmSize, tr.SizeInMemory())
						require.NoError(t, tr.Commit(), "committing a clean trie must not be needed to bound tracking")
					}
					// Misses must neither add candidates nor disturb successful-read LRU order.
					candidates := drainTrackedLookupKeys(t, tr.collapseManager)
					require.Equal(t, 3, len(candidates))
					require.Equal(t, [][]byte{keyBytesToHex([]byte("doe")), keyBytesToHex([]byte("ddog")), keyBytesToHex([]byte("dog"))}, candidates)
					actualRoot, err := tr.RootHash()
					require.NoError(t, err)
					require.Equal(t, root, actualRoot)
				}
			})
		}
	}
}

func TestTrieLookupTracking_ColdMissAccountsForDecodedNodes(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		t.Run(fmt.Sprintf("version-%d", version), func(t *testing.T) {
			tr, root := newLookupTrackingTrie(t, version)
			initialSize := tr.SizeInMemory()
			value, _, _, err := tr.GetWithVersion([]byte("bog"))
			require.NoError(t, err)
			require.Empty(t, value)
			require.Greater(t, tr.SizeInMemory(), initialSize)
			require.Equal(t, loadedLookupNodeSize(tr.root), tr.SizeInMemory())
			require.Empty(t, drainTrackedLookupKeys(t, tr.collapseManager))
			sizeAfterMiss := tr.SizeInMemory()
			_, _, _, err = tr.GetWithVersion([]byte("bog"))
			require.NoError(t, err)
			require.Equal(t, sizeAfterMiss, tr.SizeInMemory(), "already decoded nodes must not be counted twice")
			value, _, actualVersion, err := tr.GetWithVersion([]byte("dog"))
			require.NoError(t, err)
			require.Equal(t, []byte("value-dog"), value)
			require.Equal(t, uint32(version), actualVersion)
			require.Equal(t, [][]byte{keyBytesToHex([]byte("dog"))}, drainTrackedLookupKeys(t, tr.collapseManager))
			actualRoot, err := tr.RootHash()
			require.NoError(t, err)
			require.Equal(t, root, actualRoot)
		})
	}
}

func TestTrieLookupTracking_ReadErrorPreservesDecodedNodeAccounting(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		t.Run(fmt.Sprintf("storage-read-%d", failAt), func(t *testing.T) {
			tr, root := newLookupTrackingTrie(t, core.AutoBalanceEnabled)
			storage := tr.trieStorage
			initialSize := tr.SizeInMemory()
			expectedErr := errors.New("storage temporarily unavailable")
			calls := 0
			tr.trieStorage = &storageManager.StorageManagerStub{GetCalled: func(hash []byte) ([]byte, error) {
				calls++
				if calls == failAt {
					return nil, expectedErr
				}
				return storage.Get(hash)
			}}
			value, _, version, err := tr.GetWithVersion([]byte("dog"))
			require.Error(t, err)
			// The keyed storage error preserves its cause as text, not an Unwrap chain.
			require.ErrorContains(t, err, expectedErr.Error())
			require.NotNil(t, core.UnwrapGetNodeFromDBErr(err))
			require.Equal(t, failAt, calls)
			require.Empty(t, value)
			require.Zero(t, version)
			if failAt == 1 {
				require.Equal(t, initialSize, tr.SizeInMemory())
			} else {
				require.Greater(t, tr.SizeInMemory(), initialSize)
			}
			require.Equal(t, loadedLookupNodeSize(tr.root), tr.SizeInMemory())
			require.Empty(t, drainTrackedLookupKeys(t, tr.collapseManager))
			// Retry through the same partially expanded trie after storage recovers.
			value, _, version, err = tr.GetWithVersion([]byte("dog"))
			require.NoError(t, err)
			require.Equal(t, []byte("value-dog"), value)
			require.Equal(t, uint32(core.AutoBalanceEnabled), version)
			require.Equal(t, loadedLookupNodeSize(tr.root), tr.SizeInMemory())
			require.Equal(t, [][]byte{keyBytesToHex([]byte("dog"))}, drainTrackedLookupKeys(t, tr.collapseManager))
			actualRoot, err := tr.RootHash()
			require.NoError(t, err)
			require.Equal(t, root, actualRoot)
		})
	}
}

func BenchmarkTrieLookupTracking(b *testing.B) {
	for _, workload := range []string{"hit", "distinct-miss"} {
		b.Run(workload, func(b *testing.B) {
			tr, _ := newLookupTrackingTrie(b, core.AutoBalanceEnabled)
			key := []byte("dog")
			_, _, err := tr.Get(key)
			require.NoError(b, err)
			var missingKey [32]byte
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lookupKey := key
				if workload == "distinct-miss" {
					binary.BigEndian.PutUint64(missingKey[:8], uint64(i))
					lookupKey = missingKey[:]
				}
				_, _, readErr := tr.Get(lookupKey)
				if readErr != nil {
					b.Fatal(readErr)
				}
			}
		})
	}
}
