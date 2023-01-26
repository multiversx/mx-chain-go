package trie_test

import (
	errorsGo "errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/hashesHolder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	hashSize = 32
)

func getNewTrieStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            &testscommon.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		GeneralConfig:          config.TrieStorageManagerConfig{SnapshotsGoroutineNum: 1},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, hashSize),
		IdleProvider:           &testscommon.ProcessStatusHandlerStub{},
	}
}

func TestNewTrieStorageManager(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := getNewTrieStorageManagerArgs()
		args.Marshalizer = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		args := getNewTrieStorageManagerArgs()
		args.Hasher = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrNilHasher, err)
	})
	t.Run("nil checkpoint hashes holder", func(t *testing.T) {
		t.Parallel()

		args := getNewTrieStorageManagerArgs()
		args.CheckpointHashesHolder = nil
		ts, err := trie.NewTrieStorageManager(args)
		assert.Nil(t, ts)
		assert.Equal(t, trie.ErrNilCheckpointHashesHolder, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getNewTrieStorageManagerArgs()
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
		ErrChan:    make(chan error, 1),
	}
	trieStorage.SetCheckpoint(rootHash, []byte{}, iteratorChannels, nil, &trieMock.MockStatistics{})
	trie.WaitForOperationToComplete(trieStorage)

	val, err = trieStorage.GetFromCheckpoint(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, 0, len(iteratorChannels.ErrChan))
}

func TestTrieStorageManager_SetCheckpointNilErrorChan(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
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

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)
	_ = ts.Close()

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	ts.SetCheckpoint(rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{})

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	assert.Equal(t, 0, len(iteratorChannels.ErrChan))
}

func TestTrieStorageManager_SetCheckpointEmptyTrieRootHash(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := make([]byte, 32)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	ts.SetCheckpoint(rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{})

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	assert.Equal(t, 0, len(iteratorChannels.ErrChan))
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
		ErrChan:    make(chan error, 1),
	}
	trieStorage.SetCheckpoint(rootHash, []byte{}, iteratorChannels, nil, &trieMock.MockStatistics{})
	trie.WaitForOperationToComplete(trieStorage)

	val, err = trieStorage.GetFromCheckpoint(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)
	assert.Equal(t, 0, len(iteratorChannels.ErrChan))
}

func TestTrieStorageManager_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	assert.True(t, ts.IsPruningEnabled())
}

func TestTrieStorageManager_IsPruningBlocked(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	assert.False(t, ts.IsPruningBlocked())

	ts.EnterPruningBufferingMode()
	assert.True(t, ts.IsPruningBlocked())
	ts.ExitPruningBufferingMode()

	assert.False(t, ts.IsPruningBlocked())
}

func TestTrieStorageManager_Remove(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.MainStorer = testscommon.NewSnapshotPruningStorerMock()
	args.CheckpointsStorer = testscommon.NewSnapshotPruningStorerMock()
	ts, _ := trie.NewTrieStorageManager(args)

	key := []byte("key")
	value := []byte("value")

	_ = args.MainStorer.Put(key, value)
	hashes := make(common.ModifiedHashes)
	hashes[string(value)] = struct{}{}
	hashes[string(key)] = struct{}{}
	_ = args.CheckpointHashesHolder.Put(key, hashes)

	val, err := args.MainStorer.Get(key)
	assert.Nil(t, err)
	assert.NotNil(t, val)
	ok := args.CheckpointHashesHolder.ShouldCommit(key)
	assert.True(t, ok)

	err = ts.Remove(key)
	assert.Nil(t, err)

	val, err = args.MainStorer.Get(key)
	assert.Nil(t, val)
	assert.NotNil(t, err)
	ok = args.CheckpointHashesHolder.ShouldCommit(key)
	assert.False(t, ok)
}

func TestTrieStorageManager_PutInEpochClosedDb(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)
	_ = ts.Close()

	key := []byte("key")
	value := []byte("value")
	err := ts.PutInEpoch(key, value, 0)
	assert.Equal(t, errors.ErrContextClosing, err)
}

func TestTrieStorageManager_PutInEpochInvalidStorer(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	key := []byte("key")
	value := []byte("value")
	err := ts.PutInEpoch(key, value, 0)
	assert.True(t, strings.Contains(err.Error(), "invalid storer type"))
}

func TestTrieStorageManager_PutInEpoch(t *testing.T) {
	t.Parallel()

	putInEpochCalled := false
	args := getNewTrieStorageManagerArgs()
	args.MainStorer = &trieMock.SnapshotPruningStorerStub{
		MemDbMock: testscommon.NewMemDbMock(),
		PutInEpochCalled: func(key []byte, data []byte, epoch uint32) error {
			putInEpochCalled = true
			return nil
		},
	}
	ts, _ := trie.NewTrieStorageManager(args)

	key := []byte("key")
	value := []byte("value")
	err := ts.PutInEpoch(key, value, 0)
	assert.Nil(t, err)
	assert.True(t, putInEpochCalled)
}

func TestTrieStorageManager_GetLatestStorageEpochInvalidStorer(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	val, err := ts.GetLatestStorageEpoch()
	assert.Equal(t, uint32(0), val)
	assert.True(t, strings.Contains(err.Error(), "invalid storer type"))
}

func TestTrieStorageManager_GetLatestStorageEpoch(t *testing.T) {
	t.Parallel()

	getLatestSorageCalled := false
	args := getNewTrieStorageManagerArgs()
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

	args := getNewTrieStorageManagerArgs()
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

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)
	_ = ts.Close()

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	ts.TakeSnapshot("", rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{}, 0)

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	assert.Equal(t, 0, len(iteratorChannels.ErrChan))
}

func TestTrieStorageManager_TakeSnapshotEmptyTrieRootHash(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := make([]byte, 32)
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	ts.TakeSnapshot("", rootHash, rootHash, iteratorChannels, nil, &trieMock.MockStatistics{}, 0)

	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)
	assert.Equal(t, 0, len(iteratorChannels.ErrChan))
}

func TestTrieStorageManager_TakeSnapshotWithGetNodeFromDBError(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.MainStorer = testscommon.NewSnapshotPruningStorerMock()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := []byte("rootHash")
	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	missingNodesChan := make(chan []byte, 2)
	ts.TakeSnapshot("", rootHash, rootHash, iteratorChannels, missingNodesChan, &trieMock.MockStatistics{}, 0)
	_, ok := <-iteratorChannels.LeavesChan
	assert.False(t, ok)

	require.Equal(t, 1, len(iteratorChannels.ErrChan))
	errRecovered := <-iteratorChannels.ErrChan
	assert.True(t, strings.Contains(errRecovered.Error(), common.GetNodeFromDBErrorString))
}

func TestTrieStorageManager_ShouldTakeSnapshotInvalidStorer(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	assert.False(t, ts.ShouldTakeSnapshot())
}

func TestNewSnapshotTrieStorageManager_GetFromCurrentEpoch(t *testing.T) {
	t.Parallel()

	getFromCurrentEpochCalled := false
	args := getNewTrieStorageManagerArgs()
	args.MainStorer = &trieMock.SnapshotPruningStorerStub{
		MemDbMock: testscommon.NewMemDbMock(),
		GetFromCurrentEpochCalled: func(_ []byte) ([]byte, error) {
			getFromCurrentEpochCalled = true
			return nil, nil
		},
	}
	ts, _ := trie.NewTrieStorageManager(args)

	_, err := ts.GetFromCurrentEpoch([]byte("key"))
	assert.Nil(t, err)
	assert.True(t, getFromCurrentEpochCalled)
}

func TestWriteInChanNonBlocking(t *testing.T) {
	t.Parallel()

	err1 := errorsGo.New("error 1")
	err2 := errorsGo.New("error 2")
	err3 := errorsGo.New("error 3")
	t.Run("unbuffered, reader has been set up, should add", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error)
		var recovered error
		wg := sync.WaitGroup{}
		wg.Add(1)

		// set up the consumer that will be blocked until writing is done
		go func() {
			recovered = <-errChan
			wg.Done()
		}()

		time.Sleep(time.Second) // allow the go routine to start

		trie.WriteInChanNonBlocking(errChan, err1)
		wg.Wait()

		assert.Equal(t, err1, recovered)
	})
	t.Run("unbuffered, no reader should skip", func(t *testing.T) {
		t.Parallel()

		chanFinish := make(chan struct{})
		go func() {
			errChan := make(chan error)
			trie.WriteInChanNonBlocking(errChan, err1)

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

		errChan := make(chan error, 1)
		trie.WriteInChanNonBlocking(errChan, err1)
		require.Equal(t, 1, len(errChan))
		recovered := <-errChan
		assert.Equal(t, err1, recovered)
	})
	t.Run("buffered (1 element), full chan should not add, but should finish", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 1)
		trie.WriteInChanNonBlocking(errChan, err1)
		trie.WriteInChanNonBlocking(errChan, err2)

		require.Equal(t, 1, len(errChan))
		recovered := <-errChan
		assert.Equal(t, err1, recovered)
	})
	t.Run("buffered (two elements), empty chan should add", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 2)
		trie.WriteInChanNonBlocking(errChan, err1)
		require.Equal(t, 1, len(errChan))
		recovered := <-errChan
		assert.Equal(t, err1, recovered)

		trie.WriteInChanNonBlocking(errChan, err1)
		trie.WriteInChanNonBlocking(errChan, err2)
		require.Equal(t, 2, len(errChan))

		recovered = <-errChan
		assert.Equal(t, err1, recovered)
		recovered = <-errChan
		assert.Equal(t, err2, recovered)
	})
	t.Run("buffered (2 elements), full chan should not add, but should finish", func(t *testing.T) {
		t.Parallel()

		errChan := make(chan error, 2)
		trie.WriteInChanNonBlocking(errChan, err1)
		trie.WriteInChanNonBlocking(errChan, err2)
		trie.WriteInChanNonBlocking(errChan, err3)

		require.Equal(t, 2, len(errChan))
		recovered := <-errChan
		assert.Equal(t, err1, recovered)
		recovered = <-errChan
		assert.Equal(t, err2, recovered)
	})
}
