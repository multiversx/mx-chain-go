package pruning

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestImportDBPeerBootstrapSourceIsReadOnlyAndClosedWithoutDestroy(t *testing.T) {
	closed := 0
	source := &storageMocks.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) { return []byte("disk-value"), nil },
		PutCalled:         func(_, _ []byte) error { t.Fatal("must not write to disk source"); return nil },
		RemoveCalled:      func(_ []byte) error { t.Fatal("must not remove from disk source"); return nil },
		DestroyUnitCalled: func() error { t.Fatal("must not destroy disk source"); return nil },
		CloseCalled:       func() error { closed++; return nil },
	}
	s := NewImportDBPeerPruningStorer(7, source)
	require.True(t, s.HasBootstrapSource())
	value, err := s.Get([]byte("source-node"))
	require.NoError(t, err)
	require.Equal(t, "disk-value", string(value))
	_, err = s.GetFromCurrentEpoch([]byte(common.TrieSyncedKey))
	require.Error(t, err, "disk markers must not skip copying to memory")
	replacement := s.current
	s.Rotate()
	require.Same(t, replacement, s.current, "do not rotate before the startup snapshot")
	require.Nil(t, s.old)
	require.NoError(t, s.Put([]byte("source-node"), []byte("memory-value")))
	value, err = s.Get([]byte("source-node"))
	require.NoError(t, err)
	require.Equal(t, "memory-value", string(value))
	require.NoError(t, s.Remove([]byte("other-key")))
	require.Zero(t, closed)
	s.SnapshotFinished()
	require.False(t, s.HasBootstrapSource())
	require.Equal(t, 1, closed)
	s.Rotate()
	require.Same(t, replacement, s.old, "normal memory rotation resumes after initialization")
	require.NoError(t, s.Close())
	require.Equal(t, 1, closed)
}

func TestImportDBPeerCloseBeforeStartupSnapshotClosesSource(t *testing.T) {
	closed := 0
	s := NewImportDBPeerPruningStorer(7, &storageMocks.StorerStub{
		CloseCalled: func() error { closed++; return nil },
	})
	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
	require.Equal(t, 1, closed)
	require.False(t, s.HasBootstrapSource())
}

func TestRotationRoutesAllWritesAndDestroysOnlySource(t *testing.T) {
	s := NewImportDBPeerPruningStorer(7, nil)
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
	s := NewImportDBPeerPruningStorer(0, nil)
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
	s := NewImportDBPeerPruningStorer(0, nil)
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
