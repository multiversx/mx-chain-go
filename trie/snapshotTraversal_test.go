package trie

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/trie/statistics"
)

type epochSnapshotDB struct {
	get func([]byte, uint32) ([]byte, uint32, error)
	put func([]byte, []byte) error
}

func (db *epochSnapshotDB) GetWithoutAddingToCache(key []byte, epoch uint32) ([]byte, uint32, error) {
	return db.get(key, epoch)
}
func (db *epochSnapshotDB) GetFromLastEpoch(key []byte) ([]byte, error) {
	value, _, err := db.get(key, 0)
	return value, err
}
func (db *epochSnapshotDB) PutInEpochWithoutCache(key, data []byte) error { return db.put(key, data) }
func (db *epochSnapshotDB) GetIdentifier() string                         { return "snapshot test" }

func TestSnapshotTraversalUsesParentEpochAndSharedCodec(t *testing.T) {
	trieCtx := getDefaultTrieContext()
	t.Cleanup(func() { require.NoError(t, trieCtx.GetStorage().Close()) })
	leaf := getLn()
	leafBytes, err := leaf.getEncodedNode(trieCtx)
	require.NoError(t, err)
	leafHash := trieCtx.Compute(string(leafBytes))
	extension := &extensionNode{CollapsedEn: CollapsedEn{Key: []byte{1}, ChildHash: leafHash}}
	extensionBytes, err := extension.getEncodedNode(trieCtx)
	require.NoError(t, err)
	extensionHash := trieCtx.Compute(string(extensionBytes))
	root := newBranchNode()
	root.ChildrenHashes[2] = extensionHash
	rootBytes, err := root.getEncodedNode(trieCtx)
	require.NoError(t, err)
	rootHash := trieCtx.Compute(string(rootBytes))
	encoded := map[string][]byte{string(rootHash): rootBytes, string(extensionHash): extensionBytes, string(leafHash): leafBytes}
	epochs := map[string]uint32{string(rootHash): 6, string(extensionHash): 4, string(leafHash): 3}
	expectedMaxEpoch := map[string]uint32{string(rootHash): 8, string(extensionHash): 6, string(leafHash): 4}
	var reads, writes [][]byte
	db := &epochSnapshotDB{
		get: func(key []byte, maxEpoch uint32) ([]byte, uint32, error) {
			require.Equal(t, expectedMaxEpoch[string(key)], maxEpoch)
			reads = append(reads, key)
			return encoded[string(key)], epochs[string(key)], nil
		},
		put: func(key, data []byte) error {
			require.Equal(t, encoded[string(key)], data)
			writes = append(writes, key)
			return nil
		},
	}
	missing := make(chan []byte, 3)
	snapshotCtx := &snapshotTrieContext{TrieContext: trieCtx, storage: db}
	snapshot, encodedRoot, foundInEpoch, err := newSnapshotNode(snapshotCtx, rootHash, 8, missing)
	require.NoError(t, err)
	leaves := make(chan core.KeyValueHolder, 1)
	err = snapshot.commitSnapshot(snapshotCtx, foundInEpoch, leaves, missing, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, encodedRoot, 0)
	require.NoError(t, err)
	require.Equal(t, [][]byte{rootHash, extensionHash, leafHash}, reads)
	require.Equal(t, [][]byte{rootHash, leafHash, extensionHash}, writes)
	require.Empty(t, missing)
	require.Len(t, leaves, 1)
}
