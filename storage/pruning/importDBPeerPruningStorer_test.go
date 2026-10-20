package pruning

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/stretchr/testify/require"
)

func TestRotationRoutesAllWritesAndDestroysOnlySource(t *testing.T) {
	s := NewImportDBPeerPruningStorer(7)
	require.NoError(t, s.Put([]byte("source"), []byte("old")))
	require.NoError(t, s.Put([]byte(common.TrieSyncedKey), []byte(common.TrieSyncedVal)))
	source := s.current
	s.Rotate()
	replacement := s.current
	_, err := s.GetFromCurrentEpoch([]byte(common.TrieSyncedKey))
	require.Error(t, err, "old completion markers must not skip the new snapshot")
	value, err := s.Get([]byte("source"))
	require.NoError(t, err)
	require.Equal(t, "old", string(value))

	require.NoError(t, s.Put([]byte("source"), []byte("new")))
	require.NoError(t, s.PutInEpoch([]byte("repair"), []byte("repaired"), 6))
	require.NoError(t, s.PutInEpochWithoutCache([]byte("snapshot"), []byte("copied"), 7))
	s.SetEpochForPutOperation(8)
	s.Rotate()
	require.Same(t, replacement, s.current, "retargeting must not allocate a third generation")
	require.Same(t, source, s.old)
	value, _, err = s.GetWithoutAddingToCache([]byte("source"), 0)
	require.NoError(t, err)
	require.Equal(t, "new", string(value))
	s.SnapshotFinished()
	require.Nil(t, s.old)
	_, err = source.Get([]byte("source"))
	require.Error(t, err, "Close alone would leave the source map allocated")
	for _, key := range []string{"source", "repair", "snapshot"} {
		_, err = s.Get([]byte(key))
		require.NoError(t, err)
	}
	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
	_, err = s.Get([]byte("source"))
	require.ErrorIs(t, err, storage.ErrDBIsClosed)
}

func TestRotationConcurrentReplacementWrites(t *testing.T) {
	s := NewImportDBPeerPruningStorer(0)
	s.Rotate()
	var writers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		writers.Add(1)
		go func(worker int) {
			defer writers.Done()
			for i := 0; i < 300; i++ {
				key := []byte(fmt.Sprintf("%d-%d", worker, i))
				if err := s.Put(key, key); err != nil {
					t.Error(err)
					return
				}
				if _, err := s.Get(key); err != nil {
					t.Error(err)
					return
				}
			}
		}(worker)
	}
	s.SnapshotFinished()
	writers.Wait()
	for worker := 0; worker < 4; worker++ {
		_, err := s.Get([]byte(fmt.Sprintf("%d-299", worker)))
		require.NoError(t, err)
	}
}

func TestCloseDestroysBothMemoryGenerations(t *testing.T) {
	s := NewImportDBPeerPruningStorer(0)
	key := []byte("node")
	require.NoError(t, s.Put(key, []byte("old")))
	old := s.current
	s.Rotate()
	require.NoError(t, s.Put(key, []byte("new")))
	current := s.current
	require.NoError(t, s.Close())
	require.Nil(t, s.old)
	for _, db := range []*database.MemDB{old, current} {
		_, err := db.Get(key)
		require.Error(t, err, "closing must release both maps, not just mark the storer closed")
	}
	require.NoError(t, s.Close())
}
