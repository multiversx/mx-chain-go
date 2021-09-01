package trie_test

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/stretchr/testify/assert"
)

const (
	checkpointHashesHolderMaxSize = 10000000
	hashSize                      = 32
)

func getNewTrieStorageManagerArgs() trie.NewTrieStorageManagerArgs {
	return trie.NewTrieStorageManagerArgs{
		DB:                     testscommon.NewMemDbMock(),
		MainStorer:             testscommon.CreateMemUnit(),
		CheckpointsStorer:      testscommon.CreateMemUnit(),
		Marshalizer:            &mock.MarshalizerMock{},
		Hasher:                 &mock.HasherMock{},
		SnapshotDbConfig:       config.DBConfig{},
		GeneralConfig:          config.TrieStorageManagerConfig{},
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10, hashSize),
	}
}

func TestNewTrieStorageManagerNilDb(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.DB = nil
	ts, err := trie.NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, trie.ErrNilDatabase, err)
}

func TestNewTrieStorageManagerNilMarshalizer(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.Marshalizer = nil
	ts, err := trie.NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieStorageManagerNilHasher(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.Hasher = nil
	ts, err := trie.NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieStorageManagerNilCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.CheckpointHashesHolder = nil
	ts, err := trie.NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, trie.ErrNilCheckpointHashesHolder, err)
}

func TestNewTrieStorageManagerOkVals(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, err := trie.NewTrieStorageManager(args)
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestNewTrieStorageManagerWithExistingSnapshot(t *testing.T) {
	t.Parallel()

	tempDir, _ := ioutil.TempDir("", "leveldb_temp")
	cfg := config.DBConfig{
		FilePath:          tempDir,
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 1,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	}
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}

	args := getNewTrieStorageManagerArgs()
	args.SnapshotDbConfig = cfg
	args.GeneralConfig = generalCfg
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	trieStorage, _ := trie.NewTrieStorageManager(args)

	snapshotDb, err := trieStorage.NewSnapshotDb()
	assert.Nil(t, err)
	assert.NotNil(t, snapshotDb)

	key := []byte("key")
	value := []byte("value")
	err = snapshotDb.Put(key, value)
	assert.Nil(t, err)
	err = snapshotDb.Close()
	assert.Nil(t, err)

	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	newTrieStorage, _ := trie.NewTrieStorageManager(args)
	foundSnapshot := newTrieStorage.GetSnapshotThatContainsHash(key)
	assert.NotNil(t, foundSnapshot)
	assert.Equal(t, 1, newTrieStorage.SnapshotId())
}

func TestNewTrieStorageManagerLoadsSnapshotsInOrder(t *testing.T) {
	t.Parallel()

	tempDir, _ := ioutil.TempDir("", "leveldb_temp")
	cfg := config.DBConfig{
		FilePath:          tempDir,
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 1,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	}
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}

	args := getNewTrieStorageManagerArgs()
	args.SnapshotDbConfig = cfg
	args.GeneralConfig = generalCfg
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	trieStorage, _ := trie.NewTrieStorageManager(args)

	numSnapshots := 10
	for i := 0; i < numSnapshots; i++ {
		snapshotDb, _ := trieStorage.NewSnapshotDb()
		err := snapshotDb.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		assert.Nil(t, err)
		_ = snapshotDb.Close()
	}

	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	newTrieStorage, _ := trie.NewTrieStorageManager(args)

	snapshots := newTrieStorage.GetSnapshots()
	for i := 0; i < numSnapshots; i++ {
		val, err := snapshots[i].Get([]byte(strconv.Itoa(i)))
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}

	assert.Equal(t, 10, newTrieStorage.SnapshotId())
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
	trieStorage.SetCheckpoint(rootHash, nil)
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

	trieStorage.SetCheckpoint(rootHash, nil)
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

func TestTrieStorageManager_GetSnapshotDbBatchDelay(t *testing.T) {
	t.Parallel()

	batchDelay := 5
	args := getNewTrieStorageManagerArgs()
	args.SnapshotDbConfig = config.DBConfig{
		BatchDelaySeconds: batchDelay,
	}
	ts, _ := trie.NewTrieStorageManager(args)

	assert.Equal(t, batchDelay, ts.GetSnapshotDbBatchDelay())
}

func TestTrieStorageManager_Remove(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
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
