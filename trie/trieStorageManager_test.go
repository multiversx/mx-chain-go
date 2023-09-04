package trie_test

import (
	errorsGo "errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	storageMx "github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	providedKey = []byte("key")
	providedVal = []byte("value")
	expectedErr = errorsGo.New("expected error")
)

// errChanWithLen extends the BufferedErrChan interface with a Len method
type errChanWithLen interface {
	common.BufferedErrChan
	Len() int
}

func TestNewTrieStorageManager(t *testing.T) {
	t.Parallel()

	t.Run("nil main storer", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.True(t, strings.Contains(err.Error(), trie.ErrNilStorer.Error()))
	})
	t.Run("nil checkpoints storer", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.CheckpointsStorer = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.True(t, strings.Contains(err.Error(), trie.ErrNilStorer.Error()))
	})
	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.Marshalizer = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.Hasher = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrNilHasher, err)
	})
	t.Run("nil checkpoint hashes holder", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.CheckpointHashesHolder = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrNilCheckpointHashesHolder, err)
	})
	t.Run("nil idle provider", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.IdleProvider = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrNilIdleNodeProvider, err)
	})
	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.GeneralConfig.SnapshotsGoroutineNum = 0
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Error(t, err)
	})
	t.Run("invalid identifier", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.Identifier = ""
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrInvalidIdentifier, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, err)
		assert.NotNil(t, ts)
	})
}

func TestTrieCheckpoint(t *testing.T) {
	t.Parallel()

	tr, trieStorage := trie.CreateSmallTestTrieAndStorageManager()
	rootHash, _ := tr.RootHash()

	val, err := trieStorage.GetFromCheckpoint(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)

	dirtyHashes := trie.GetDirtyHashes(tr)

	trieStorage.AddDirtyCheckpointHashes(rootHash, dirtyHashes)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	trieStorage.SetCheckpoint(rootHash, []byte{}, iteratorChannels, nil, &trieMock.MockStatistics{})
	trie.WaitForOperationToComplete(trieStorage)

	val, err = trieStorage.GetFromCheckpoint(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	ch, ok := iteratorChannels.ErrChan.(errChanWithLen)
	assert.True(t, ok)
	assert.Equal(t, 0, ch.Len())
}

func TestTrieStorageManager_SetCheckpointNilErrorChan(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    nil,
	}
	ts.SetCheckpoint(rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{})

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)

	_ = ts.Close()
}

func TestTrieStorageManager_SetCheckpointClosedDb(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)
	_ = ts.Close()

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	ts.SetCheckpoint(rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{})

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	ch, ok := iteratorChannels.ErrChan.(errChanWithLen)
	assert.True(t, ok)
	assert.Equal(t, 0, ch.Len())
}

func TestTrieStorageManager_SetCheckpointEmptyTrieRootHash(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := make([]byte, 32)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	ts.SetCheckpoint(rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{})

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	ch, ok := iteratorChannels.ErrChan.(errChanWithLen)
	assert.True(t, ok)
	assert.Equal(t, 0, ch.Len())
}

func TestTrieCheckpoint_DoesNotSaveToCheckpointStorageIfNotDirty(t *testing.T) {
	t.Parallel()

	tr, trieStorage := trie.CreateSmallTestTrieAndStorageManager()
	rootHash, _ := tr.RootHash()

	val, err := trieStorage.GetFromCheckpoint(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	trieStorage.SetCheckpoint(rootHash, []byte{}, iteratorChannels, nil, &trieMock.MockStatistics{})
	trie.WaitForOperationToComplete(trieStorage)

	val, err = trieStorage.GetFromCheckpoint(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)
	ch, ok := iteratorChannels.ErrChan.(errChanWithLen)
	assert.True(t, ok)
	assert.Equal(t, 0, ch.Len())
}

func TestTrieStorageManager_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)

	assert.True(t, ts.IsPruningEnabled())
}

func TestTrieStorageManager_IsPruningBlocked(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)
	ts.ExitPruningBufferingMode() // early exit

	assert.False(t, ts.IsPruningBlocked())

	ts.EnterPruningBufferingMode()
	assert.True(t, ts.IsPruningBlocked())
	ts.ExitPruningBufferingMode()

	assert.False(t, ts.IsPruningBlocked())
}

func TestTrieStorageManager_Remove(t *testing.T) {
	t.Parallel()

	t.Run("main storer not snapshotPruningStorer should call remove", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &storage.StorerStub{
			RemoveCalled: func(key []byte) error {
				wasCalled = true
				return nil
			},
		}
		ts, _ := trie.NewTrieStorageManager(args)

		err := ts.Remove(providedKey)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = testscommon.NewSnapshotPruningStorerMock()
		args.CheckpointsStorer = testscommon.NewSnapshotPruningStorerMock()
		ts, _ := trie.NewTrieStorageManager(args)

		_ = args.MainStorer.Put(providedKey, providedVal)
		hashes := make(common.ModifiedHashes)
		hashes[string(providedVal)] = struct{}{}
		hashes[string(providedKey)] = struct{}{}
		_ = args.CheckpointHashesHolder.Put(providedKey, hashes)

		val, err := args.MainStorer.Get(providedKey)
		assert.Nil(t, err)
		assert.NotNil(t, val)
		ok := args.CheckpointHashesHolder.ShouldCommit(providedKey)
		assert.True(t, ok)

		err = ts.Remove(providedKey)
		assert.Nil(t, err)

		val, err = args.MainStorer.Get(providedKey)
		assert.Nil(t, val)
		assert.NotNil(t, err)
		ok = args.CheckpointHashesHolder.ShouldCommit(providedKey)
		assert.False(t, ok)
	})
}

func TestTrieStorageManager_RemoveFromCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := trie.GetDefaultTrieStorageManagerParameters()
	args.CheckpointHashesHolder = &trieMock.CheckpointHashesHolderStub{
		RemoveCalled: func(bytes []byte) {
			wasCalled = true
		},
	}
	ts, _ := trie.NewTrieStorageManager(args)

	ts.RemoveFromCheckpointHashesHolder(providedKey)
	assert.True(t, wasCalled)
}

func TestTrieStorageManager_SetEpochForPutOperation(t *testing.T) {
	t.Parallel()

	t.Run("main storer not epochStorer should early exit", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &storage.StorerStub{}
		ts, err := trie.NewTrieStorageManager(args)
		require.Nil(t, err)

		ts.SetEpochForPutOperation(0)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedEpoch := uint32(100)
		wasCalled := false
		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &storageManager.StorageManagerStub{
			SetEpochForPutOperationCalled: func(u uint32) {
				assert.Equal(t, providedEpoch, u)
				wasCalled = true
			},
		}
		ts, err := trie.NewTrieStorageManager(args)
		require.Nil(t, err)

		ts.SetEpochForPutOperation(providedEpoch)
		assert.True(t, wasCalled)
	})
}

func TestTrieStorageManager_RemoveFromAllActiveEpochs(t *testing.T) {
	t.Parallel()

	RemoveFromAllActiveEpochsCalled := false
	removeFromCheckpointCalled := false
	args := trie.GetDefaultTrieStorageManagerParameters()
	args.MainStorer = &trieMock.SnapshotPruningStorerStub{
		MemDbMock: testscommon.NewMemDbMock(),
		RemoveFromAllActiveEpochsCalled: func(key []byte) error {
			RemoveFromAllActiveEpochsCalled = true
			return nil
		},
	}
	args.CheckpointHashesHolder = &trieMock.CheckpointHashesHolderStub{
		RemoveCalled: func(bytes []byte) {
			removeFromCheckpointCalled = true
		},
	}
	ts, _ := trie.NewTrieStorageManager(args)

	err := ts.RemoveFromAllActiveEpochs([]byte("key"))
	assert.Nil(t, err)
	assert.True(t, RemoveFromAllActiveEpochsCalled)
	assert.True(t, removeFromCheckpointCalled)
}

func TestTrieStorageManager_PutInEpochClosedDb(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)
	_ = ts.Close()

	err := ts.PutInEpoch(providedKey, providedVal, 0)
	assert.Equal(t, core.ErrContextClosing, err)
}

func TestTrieStorageManager_PutInEpochInvalidStorer(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	args.MainStorer = testscommon.CreateStorerWithStats()
	ts, _ := trie.NewTrieStorageManager(args)

	err := ts.PutInEpoch(providedKey, providedVal, 0)
	assert.True(t, strings.Contains(err.Error(), "invalid storer type"))
}

func TestTrieStorageManager_PutInEpoch(t *testing.T) {
	t.Parallel()

	putInEpochCalled := false
	args := trie.GetDefaultTrieStorageManagerParameters()
	args.MainStorer = &trieMock.SnapshotPruningStorerStub{
		MemDbMock: testscommon.NewMemDbMock(),
		PutInEpochCalled: func(key []byte, data []byte, epoch uint32) error {
			putInEpochCalled = true
			return nil
		},
	}
	ts, _ := trie.NewTrieStorageManager(args)

	err := ts.PutInEpoch(providedKey, providedVal, 0)
	assert.Nil(t, err)
	assert.True(t, putInEpochCalled)
}

func TestTrieStorageManager_GetLatestStorageEpochInvalidStorer(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	args.MainStorer = testscommon.CreateStorerWithStats()
	ts, _ := trie.NewTrieStorageManager(args)

	val, err := ts.GetLatestStorageEpoch()
	assert.Equal(t, uint32(0), val)
	assert.True(t, strings.Contains(err.Error(), "invalid storer type"))
}

func TestTrieStorageManager_GetLatestStorageEpoch(t *testing.T) {
	t.Parallel()

	getLatestSorageCalled := false
	args := trie.GetDefaultTrieStorageManagerParameters()
	args.MainStorer = &trieMock.SnapshotPruningStorerStub{
		MemDbMock: testscommon.NewMemDbMock(),
		GetLatestStorageEpochCalled: func() (uint32, error) {
			getLatestSorageCalled = true
			return 4, nil
		},
	}
	ts, _ := trie.NewTrieStorageManager(args)

	val, err := ts.GetLatestStorageEpoch()
	assert.Equal(t, uint32(4), val)
	assert.Nil(t, err)
	assert.True(t, getLatestSorageCalled)
}

func TestTrieStorageManager_TakeSnapshotNilErrorChan(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    nil,
	}
	ts.TakeSnapshot("", rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{}, 0)

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)

	_ = ts.Close()
}

func TestTrieStorageManager_TakeSnapshotClosedDb(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)
	_ = ts.Close()

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	ts.TakeSnapshot("", rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{}, 0)

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	ch, ok := iteratorChannels.ErrChan.(errChanWithLen)
	assert.True(t, ok)
	assert.Equal(t, 0, ch.Len())
}

func TestTrieStorageManager_TakeSnapshotEmptyTrieRootHash(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := make([]byte, 32)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	ts.TakeSnapshot("", rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{}, 0)

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	ch, ok := iteratorChannels.ErrChan.(errChanWithLen)
	assert.True(t, ok)
	assert.Equal(t, 0, ch.Len())
}

func TestTrieStorageManager_TakeSnapshotWithGetNodeFromDBError(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	args.MainStorer = testscommon.NewSnapshotPruningStorerMock()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	missingNodesChan := make(chan []byte, 2)
	ts.TakeSnapshot("", rootHash, rootHash, iteratorChannels, missingNodesChan, &trieMock.MockStatistics{}, 0)
	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)

	ch, ok := iteratorChannels.ErrChan.(errChanWithLen)
	assert.True(t, ok)
	assert.Equal(t, 1, ch.Len())
	errRecovered := iteratorChannels.ErrChan.ReadFromChanNonBlocking()
	assert.True(t, strings.Contains(errRecovered.Error(), core.GetNodeFromDBErrorString))
}

func TestTrieStorageManager_ShouldTakeSnapshot(t *testing.T) {
	t.Parallel()

	t.Run("invalid storer should return false", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = testscommon.CreateStorerWithStats()
		ts, err := trie.NewTrieStorageManager(args)
		require.Nil(t, err)

		assert.False(t, ts.ShouldTakeSnapshot())
	})
	t.Run("trie synced should return false", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &trieMock.SnapshotPruningStorerStub{
			GetFromCurrentEpochCalled: func(key []byte) ([]byte, error) {
				return []byte(common.TrieSyncedVal), nil
			},
			MemDbMock: testscommon.NewMemDbMock(),
		}
		ts, _ := trie.NewTrieStorageManager(args)

		assert.False(t, ts.ShouldTakeSnapshot())
	})
	t.Run("GetFromOldEpochsWithoutAddingToCacheCalled returns ActiveDBVal should return true", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &trieMock.SnapshotPruningStorerStub{
			GetFromCurrentEpochCalled: func(key []byte) ([]byte, error) {
				return nil, expectedErr // isTrieSynced returns false
			},
			GetFromOldEpochsWithoutAddingToCacheCalled: func(key []byte) ([]byte, core.OptionalUint32, error) {
				return []byte(common.ActiveDBVal), core.OptionalUint32{}, nil
			},
			MemDbMock: testscommon.NewMemDbMock(),
		}
		ts, _ := trie.NewTrieStorageManager(args)

		assert.True(t, ts.ShouldTakeSnapshot())
	})
}

func TestTrieStorageManager_Get(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		ts, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
		_ = ts.Close()

		val, err := ts.Get(providedKey)
		assert.Equal(t, core.ErrContextClosing, err)
		assert.Nil(t, val)
	})
	t.Run("main storer closing should error", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &storage.StorerStub{
			GetWithStatsCalled: func(key []byte) ([]byte, bool, error) {
				return nil, false, storageMx.ErrDBIsClosed
			},
		}
		ts, _ := trie.NewTrieStorageManager(args)

		val, err := ts.Get(providedKey)
		assert.Equal(t, storageMx.ErrDBIsClosed, err)
		assert.Nil(t, val)
	})
	t.Run("checkpoints storer closing should error", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.CheckpointsStorer = &storage.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, storageMx.ErrDBIsClosed
			},
		}
		ts, _ := trie.NewTrieStorageManager(args)

		val, err := ts.Get(providedKey)
		assert.Equal(t, storageMx.ErrDBIsClosed, err)
		assert.Nil(t, val)
	})
	t.Run("should return from main storer", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		_ = args.MainStorer.Put(providedKey, providedVal)
		ts, _ := trie.NewTrieStorageManager(args)

		val, err := ts.Get(providedKey)
		assert.Nil(t, err)
		assert.Equal(t, providedVal, val)
	})
	t.Run("should return from checkpoints storer", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		_ = args.CheckpointsStorer.Put(providedKey, providedVal)
		ts, _ := trie.NewTrieStorageManager(args)

		val, err := ts.Get(providedKey)
		assert.Nil(t, err)
		assert.Equal(t, providedVal, val)
	})
}

func TestNewSnapshotTrieStorageManager_GetFromCurrentEpoch(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		ts, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
		_ = ts.Close()

		val, err := ts.GetFromCurrentEpoch(providedKey)
		assert.Equal(t, core.ErrContextClosing, err)
		assert.Nil(t, val)
	})
	t.Run("main storer not snapshotPruningStorer should error", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &storage.StorerStub{}
		ts, _ := trie.NewTrieStorageManager(args)

		val, err := ts.GetFromCurrentEpoch(providedKey)
		assert.True(t, strings.Contains(err.Error(), "invalid storer"))
		assert.Nil(t, val)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		getFromCurrentEpochCalled := false
		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &trieMock.SnapshotPruningStorerStub{
			MemDbMock: testscommon.NewMemDbMock(),
			GetFromCurrentEpochCalled: func(_ []byte) ([]byte, error) {
				getFromCurrentEpochCalled = true
				return nil, nil
			},
		}
		ts, _ := trie.NewTrieStorageManager(args)

		_, err := ts.GetFromCurrentEpoch(providedKey)
		assert.Nil(t, err)
		assert.True(t, getFromCurrentEpochCalled)
	})
}

func TestTrieStorageManager_Put(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		ts, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
		_ = ts.Close()

		err := ts.Put(providedKey, providedVal)
		assert.Equal(t, core.ErrContextClosing, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ts, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())

		_ = ts.Put(providedKey, providedVal)
		val, err := ts.Get(providedKey)
		assert.Nil(t, err)
		assert.Equal(t, providedVal, val)
	})
}

func TestTrieStorageManager_PutInEpochWithoutCache(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		ts, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
		_ = ts.Close()

		err := ts.PutInEpochWithoutCache(providedKey, providedVal, 0)
		assert.Equal(t, core.ErrContextClosing, err)
	})
	t.Run("main storer not snapshotPruningStorer should error", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &storage.StorerStub{}
		ts, _ := trie.NewTrieStorageManager(args)

		err := ts.PutInEpochWithoutCache(providedKey, providedVal, 0)
		assert.True(t, strings.Contains(err.Error(), "invalid storer"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = testscommon.NewSnapshotPruningStorerMock()
		ts, _ := trie.NewTrieStorageManager(args)

		err := ts.PutInEpochWithoutCache(providedKey, providedVal, 0)
		assert.Nil(t, err)
	})
}

func TestTrieStorageManager_Close(t *testing.T) {
	t.Parallel()

	t.Run("error on main storer close", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.MainStorer = &storage.StorerStub{
			CloseCalled: func() error {
				return expectedErr
			},
		}
		ts, _ := trie.NewTrieStorageManager(args)

		err := ts.Close()
		assert.True(t, errorsGo.Is(err, expectedErr))
	})
	t.Run("error on checkpoints storer close", func(t *testing.T) {
		t.Parallel()

		args := trie.GetDefaultTrieStorageManagerParameters()
		args.CheckpointsStorer = &storage.StorerStub{
			CloseCalled: func() error {
				return expectedErr
			},
		}
		ts, _ := trie.NewTrieStorageManager(args)

		err := ts.Close()
		assert.True(t, errorsGo.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ts, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())

		err := ts.Close()
		assert.NoError(t, err)
	})
}

func TestWriteInChanNonBlocking(t *testing.T) {
	t.Parallel()

	err1 := errorsGo.New("error 1")
	err2 := errorsGo.New("error 2")
	err3 := errorsGo.New("error 3")
	t.Run("unbuffered, reader has been set up, should add", func(t *testing.T) {
		t.Parallel()

		errChannel := make(chan error)
		var recovered error
		wg := sync.WaitGroup{}
		wg.Add(1)

		// set up the consumer that will be blocked until writing is done
		go func() {
			recovered = <-errChannel
			wg.Done()
		}()

		time.Sleep(time.Second) // allow the go routine to start

		trie.WriteInChanNonBlocking(errChannel, err1)
		wg.Wait()

		assert.Equal(t, err1, recovered)
	})
	t.Run("unbuffered, no reader should skip", func(t *testing.T) {
		t.Parallel()

		chanFinish := make(chan struct{})
		go func() {
			errChannel := make(chan error)
			trie.WriteInChanNonBlocking(errChannel, err1)

			close(chanFinish)
		}()

		select {
		case <-chanFinish:
		case <-time.After(time.Second * 5):
			assert.Fail(t, "timeout, WriteInChanNonBlocking is blocking on an unbuffered chan")
		}
	})
	t.Run("buffered (one element), empty chan should add", func(t *testing.T) {
		t.Parallel()

		errChannel := errChan.NewErrChanWrapper()
		errChannel.WriteInChanNonBlocking(err1)

		require.Equal(t, 1, errChannel.Len())
		recovered := errChannel.ReadFromChanNonBlocking()
		assert.Equal(t, err1, recovered)
	})
	t.Run("buffered (1 element), full chan should not add, but should finish", func(t *testing.T) {
		t.Parallel()

		errChannel := errChan.NewErrChanWrapper()
		errChannel.WriteInChanNonBlocking(err1)
		errChannel.WriteInChanNonBlocking(err2)

		require.Equal(t, 1, errChannel.Len())
		recovered := errChannel.ReadFromChanNonBlocking()
		assert.Equal(t, err1, recovered)
	})
	t.Run("buffered (two elements), empty chan should add", func(t *testing.T) {
		t.Parallel()

		errChannel := make(chan error, 2)
		trie.WriteInChanNonBlocking(errChannel, err1)
		require.Equal(t, 1, len(errChannel))
		recovered := <-errChannel
		assert.Equal(t, err1, recovered)

		trie.WriteInChanNonBlocking(errChannel, err1)
		trie.WriteInChanNonBlocking(errChannel, err2)
		require.Equal(t, 2, len(errChannel))

		recovered = <-errChannel
		assert.Equal(t, err1, recovered)
		recovered = <-errChannel
		assert.Equal(t, err2, recovered)
	})
	t.Run("buffered (2 elements), full chan should not add, but should finish", func(t *testing.T) {
		t.Parallel()

		errChannel := make(chan error, 2)
		trie.WriteInChanNonBlocking(errChannel, err1)
		trie.WriteInChanNonBlocking(errChannel, err2)
		trie.WriteInChanNonBlocking(errChannel, err3)

		require.Equal(t, 2, len(errChannel))
		recovered := <-errChannel
		assert.Equal(t, err1, recovered)
		recovered = <-errChannel
		assert.Equal(t, err2, recovered)
	})
}

func TestTrieStorageManager_GetIdentifier(t *testing.T) {
	t.Parallel()

	expectedId := "testId"
	args := trie.GetDefaultTrieStorageManagerParameters()
	args.Identifier = expectedId
	ts, _ := trie.NewTrieStorageManager(args)

	id := ts.GetIdentifier()
	assert.Equal(t, expectedId, id)
}
