package trie

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie/collapseManager"
	"github.com/stretchr/testify/require"
)

func newCommitRetryTrie(t *testing.T, storage common.StorageManager, version core.TrieNodeVersion) *patriciaMerkleTrie {
	t.Helper()
	marshaller, hasher := getTestMarshalizerAndHasher()
	flags := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	if version == core.AutoBalanceEnabled {
		flags.AddActiveFlags(common.AutoBalanceDataTriesFlag)
	}
	manager, err := collapseManager.NewCollapseManager(1<<20, 100)
	require.NoError(t, err)
	tr, err := NewTrie(storage, marshaller, hasher, flags, manager)
	require.NoError(t, err)
	return tr
}

func commitRetryStoredNodes(db *testscommon.MemDbMock) map[string][]byte {
	nodes := make(map[string][]byte)
	db.RangeKeys(func(key, value []byte) bool {
		nodes[string(key)] = bytes.Clone(value)
		return true
	})
	return nodes
}

func TestTrieCommitRetry_RootStaysDirtyUntilPersistenceSucceeds(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		for _, shape := range []struct {
			name string
			kind byte
			keys []string
		}{
			{"leaf", leaf, []string{"key"}},
			{"branch", branch, []string{"a", "b"}},
			{"extension", extension, []string{"dog", "ddog"}},
		} {
			for _, failure := range []string{"write", "encoding"} {
				t.Run(fmt.Sprintf("%s/version-%d/%s", shape.name, version, failure), func(t *testing.T) {
					db, controlDB := testscommon.NewMemDbMock(), testscommon.NewMemDbMock()
					expectedErr := errors.New("commit interrupted")
					var failingRoot []byte
					failWrites := false
					putCalls := 0
					attempts, successfulWrites := make(map[string]int), make(map[string]int)
					storage := &storageManager.StorageManagerStub{GetCalled: db.Get, PutCalled: func(key, value []byte) error {
						putCalls++
						attempts[string(key)]++
						if failWrites && bytes.Equal(key, failingRoot) {
							return expectedErr
						}
						// MemDbMock.PutCalled runs after insertion, so inject the
						// failure here to model a write that did not persist.
						err := db.Put(key, value)
						if err == nil {
							successfulWrites[string(key)]++
						}
						return err
					}}
					tr := newCommitRetryTrie(t, storage, version)
					control := newCommitRetryTrie(t, &storageManager.StorageManagerStub{GetCalled: controlDB.Get, PutCalled: controlDB.Put}, version)
					for cycle := 0; cycle < 2; cycle++ {
						for _, key := range shape.keys {
							value := []byte(fmt.Sprintf("cycle-%d/%s", cycle, key))
							require.NoError(t, tr.UpdateWithVersion([]byte(key), value, version))
							require.NoError(t, control.UpdateWithVersion([]byte(key), value, version))
						}
						root, err := tr.RootHash()
						require.NoError(t, err)
						encoded, err := tr.root.getEncodedNode()
						require.NoError(t, err)
						require.Equal(t, shape.kind, encoded[len(encoded)-1])
						failingRoot = root
						originalMarshaller := tr.root.getMarshalizer()
						if failure == "write" {
							failWrites = true
						} else {
							tr.root.setMarshalizer(&marshallerMock.MarshalizerStub{MarshalCalled: func(interface{}) ([]byte, error) {
								return nil, expectedErr
							}})
						}
						for attempt := 0; attempt < 2; attempt++ {
							require.ErrorIs(t, tr.Commit(), expectedErr)
							require.True(t, tr.root.isDirty(), "a failed root write or encoding must remain retryable")
							_, readErr := db.Get(root)
							require.Error(t, readErr, "the failed root must not be present in storage")
						}
						failWrites = false
						tr.root.setMarshalizer(originalMarshaller)
						require.NoError(t, tr.Commit())
						require.False(t, tr.root.isDirty())
						require.NoError(t, control.Commit())
						controlRoot, err := control.RootHash()
						require.NoError(t, err)
						require.Equal(t, controlRoot, root)
						require.Equal(t, commitRetryStoredNodes(controlDB), commitRetryStoredNodes(db), "recovery must preserve every serialized node")
						for key, encodedNode := range commitRetryStoredNodes(db) {
							require.Equal(t, []byte(key), tr.hasher.Compute(string(encodedNode)))
						}
						expectedAttempts := 1
						if failure == "write" {
							expectedAttempts = 3
						}
						require.Equal(t, expectedAttempts, attempts[string(root)])
						for key, writes := range successfulWrites {
							require.Equal(t, 1, writes, "successfully persisted node %x must not be rewritten", key)
						}
						reader, err := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
						require.NoError(t, err)
						for _, key := range shape.keys {
							value, _, leafVersion, readErr := reader.GetWithVersion([]byte(key))
							require.NoError(t, readErr)
							require.Equal(t, []byte(fmt.Sprintf("cycle-%d/%s", cycle, key)), value)
							require.Equal(t, uint32(version), leafVersion)
						}
						beforeCleanCommit := putCalls
						require.NoError(t, tr.Commit())
						require.Equal(t, beforeCleanCommit, putCalls)
					}
				})
			}
		}
	}
}

func TestTrieCommitRetry_PartialChildFailurePreservesPendingNodes(t *testing.T) {
	for _, version := range []core.TrieNodeVersion{core.NotSpecified, core.AutoBalanceEnabled} {
		for _, collapse := range []bool{false, true} {
			t.Run(fmt.Sprintf("version-%d/collapse-%t", version, collapse), func(t *testing.T) {
				db, controlDB := testscommon.NewMemDbMock(), testscommon.NewMemDbMock()
				expectedErr := errors.New("child write interrupted")
				var failedHash []byte
				attemptsInCycle, failuresLeft := 0, 0
				successfulWrites := make(map[string]int)
				storage := &storageManager.StorageManagerStub{GetCalled: db.Get, PutCalled: func(key, value []byte) error {
					attemptsInCycle++
					if failedHash == nil && attemptsInCycle == 2 {
						failedHash = bytes.Clone(key)
					}
					if failuresLeft > 0 && bytes.Equal(key, failedHash) {
						failuresLeft--
						return expectedErr
					}
					err := db.Put(key, value)
					if err == nil {
						successfulWrites[string(key)]++
					}
					return err
				}}
				tr := newCommitRetryTrie(t, storage, version)
				control := newCommitRetryTrie(t, &storageManager.StorageManagerStub{GetCalled: controlDB.Get, PutCalled: controlDB.Put}, version)
				keys := []string{"dog", "doe", "ddog"}
				if collapse {
					keys = nil
					for i := 0; i < 64; i++ {
						keys = append(keys, fmt.Sprintf("key-%d", i))
					}
				}
				valueFor := func(cycle int, key string) []byte {
					value := []byte(fmt.Sprintf("cycle-%d/%s", cycle, key))
					if collapse {
						value = append(value, bytes.Repeat([]byte{'x'}, 32<<10)...)
					}
					return value
				}
				for cycle := 0; cycle < 2; cycle++ {
					for _, key := range keys {
						value := valueFor(cycle, key)
						require.NoError(t, tr.UpdateWithVersion([]byte(key), value, version))
						require.NoError(t, control.UpdateWithVersion([]byte(key), value, version))
					}
					root, err := tr.RootHash()
					require.NoError(t, err)
					failedHash, attemptsInCycle, failuresLeft = nil, 0, 2
					for attempt := 0; attempt < 2; attempt++ {
						require.ErrorIs(t, tr.Commit(), expectedErr)
						require.NotEmpty(t, failedHash)
						require.True(t, tr.root.isDirty())
						pending, dirtyErr := tr.GetDirtyHashes()
						require.NoError(t, dirtyErr)
						require.Contains(t, pending, string(failedHash), "the unwritten child must remain in the pending set")
						require.Contains(t, pending, string(root), "the ancestor must stay pending until its children commit")
						_, readErr := db.Get(failedHash)
						require.Error(t, readErr)
					}
					require.Zero(t, failuresLeft)
					if collapse {
						require.Greater(t, tr.SizeInMemory(), 1<<20)
					}
					require.NoError(t, tr.Commit())
					require.NoError(t, control.Commit())
					if collapse {
						require.Less(t, tr.SizeInMemory(), 1<<20, "collapse must follow successful persistence of all children")
					}
					require.Equal(t, commitRetryStoredNodes(controlDB), commitRetryStoredNodes(db))
					for key, writes := range successfulWrites {
						require.Equal(t, 1, writes, "already persisted child %x must not be rewritten", key)
					}
					reader, err := tr.Recreate(holders.NewDefaultRootHashesHolder(root))
					require.NoError(t, err)
					for _, key := range keys {
						value, _, leafVersion, readErr := reader.GetWithVersion([]byte(key))
						require.NoError(t, readErr)
						require.True(t, bytes.Equal(valueFor(cycle, key), value))
						require.Equal(t, uint32(version), leafVersion)
					}
				}
			})
		}
	}
}
