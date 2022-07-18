package trie_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

const (
	hashSize = 32
)

func getNewTrieStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            &mock.MarshalizerMock{},
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
	trieStorage.SetCheckpoint(rootHash, []byte{}, nil, nil, &trieMock.MockStatistics{})
	trie.WaitForOperationToComplete(trieStorage)

	val, err = trieStorage.GetFromCheckpoint(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, val)
}

func TestTrieCheckpoint_DoesNotSaveToCheckpointStorageIfNotDirty(t *testing.T) {
	t.Parallel()

	tr, trieStorage := trie.CreateSmallTestTrieAndStorageManager()
	rootHash, _ := tr.RootHash()

	val, err := trieStorage.GetFromCheckpoint(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)

	trieStorage.SetCheckpoint(rootHash, []byte{}, nil, nil, &trieMock.MockStatistics{})
	trie.WaitForOperationToComplete(trieStorage)

	val, err = trieStorage.GetFromCheckpoint(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)
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
		DB: memorydb.New(),
		PutInEpochWithoutCacheCalled: func(key []byte, data []byte, epoch uint32) error {
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
		DB: memorydb.New(),
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

func TestTrieStorageManager_TakeSnapshotClosedDb(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)
	_ = ts.Close()

	rootHash := []byte("rootHash")
	leavesChan := make(chan core.KeyValueHolder)
	ts.TakeSnapshot(rootHash, rootHash, leavesChan, nil, &trieMock.MockStatistics{}, 0)

	_, ok := <-leavesChan
	assert.False(t, ok)
}

func TestTrieStorageManager_TakeSnapshotEmptyTrieRootHash(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := make([]byte, 32)
	leavesChan := make(chan core.KeyValueHolder)
	ts.TakeSnapshot(rootHash, rootHash, leavesChan, nil, &trieMock.MockStatistics{}, 0)

	_, ok := <-leavesChan
	assert.False(t, ok)
}

func TestTrieStorageManager_TakeSnapshot(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManager(args)

	rootHash := []byte("rootHash")
	leavesChan := make(chan core.KeyValueHolder)
	missingNodesChan := make(chan []byte, 2)
	ts.TakeSnapshot(rootHash, rootHash, leavesChan, missingNodesChan, &trieMock.MockStatistics{}, 0)
	_, ok := <-leavesChan
	assert.False(t, ok)
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
		DB: memorydb.New(),
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
