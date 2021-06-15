package trie

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie/hashesHolder"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

const (
	snapshotDelay                 = time.Second
	checkpointHashesHolderMaxSize = 10000000
	hashSize                      = 32
)

func getNewTrieStorageManagerArgs() NewTrieStorageManagerArgs {
	return NewTrieStorageManagerArgs{
		DB:                     mock.NewMemDbMock(),
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
	ts, err := NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilDatabase, err)
}

func TestNewTrieStorageManagerNilMarshalizer(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.Marshalizer = nil
	ts, err := NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestNewTrieStorageManagerNilHasher(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.Hasher = nil
	ts, err := NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilHasher, err)
}

func TestNewTrieStorageManagerNilCheckpointHashesHolder(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	args.CheckpointHashesHolder = nil
	ts, err := NewTrieStorageManager(args)
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilCheckpointHashesHolder, err)
}

func TestNewTrieStorageManagerOkVals(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, err := NewTrieStorageManager(args)
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

	msh, hsh := getTestMarshalizerAndHasher()
	args := getNewTrieStorageManagerArgs()
	args.Marshalizer = msh
	args.Hasher = hsh
	args.SnapshotDbConfig = cfg
	args.GeneralConfig = generalCfg
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	trieStorage, _ := NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(5)
	tr, _ := NewTrie(trieStorage, msh, hsh, maxTrieLevelInMemory)

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash, true)
	time.Sleep(snapshotDelay)

	trieStorage.storageOperationMutex.Lock()
	_ = trieStorage.snapshots[0].Close()
	trieStorage.storageOperationMutex.Unlock()

	args = getNewTrieStorageManagerArgs()
	args.Marshalizer = msh
	args.Hasher = hsh
	args.SnapshotDbConfig = cfg
	args.GeneralConfig = generalCfg
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	newTrieStorage, _ := NewTrieStorageManager(args)
	snapshot := newTrieStorage.GetSnapshotThatContainsHash(rootHash)
	assert.NotNil(t, snapshot)
	assert.Equal(t, 1, newTrieStorage.snapshotId)
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

	msh, hsh := getTestMarshalizerAndHasher()
	args := getNewTrieStorageManagerArgs()
	args.Marshalizer = msh
	args.Hasher = hsh
	args.SnapshotDbConfig = cfg
	args.GeneralConfig = generalCfg
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)

	trieStorage, _ := NewTrieStorageManager(args)
	maxTrieLevelInMemory := uint(5)
	tr, _ := NewTrie(trieStorage, msh, hsh, maxTrieLevelInMemory)

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash, true)
	time.Sleep(snapshotDelay)

	numSnapshots := 10
	for i := 0; i < numSnapshots; i++ {
		_ = tr.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		_ = tr.Commit()
		rootHash, _ = tr.RootHash()
		trieStorage.TakeSnapshot(rootHash, true)
		time.Sleep(snapshotDelay)
	}

	trieStorage.storageOperationMutex.Lock()

	val, err := trieStorage.snapshots[0].Get(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)
	val, err = trieStorage.snapshots[1].Get(rootHash)
	assert.NotNil(t, val)
	assert.Nil(t, err)

	_ = trieStorage.snapshots[0].Close()
	_ = trieStorage.snapshots[1].Close()
	trieStorage.storageOperationMutex.Unlock()

	args = getNewTrieStorageManagerArgs()
	args.Marshalizer = msh
	args.Hasher = hsh
	args.SnapshotDbConfig = cfg
	args.GeneralConfig = generalCfg
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	newTrieStorage, _ := NewTrieStorageManager(args)

	newTrieStorage.storageOperationMutex.Lock()
	val, err = newTrieStorage.snapshots[0].Get(rootHash)
	assert.NotNil(t, err)
	assert.Nil(t, val)
	val, err = newTrieStorage.snapshots[1].Get(rootHash)
	assert.NotNil(t, val)
	assert.Nil(t, err)

	assert.Equal(t, 11, newTrieStorage.snapshotId)
	newTrieStorage.storageOperationMutex.Unlock()

}

func TestRecreateTrieFromSnapshotDb(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	storageManager := tr.GetStorageManager()
	rootHash, _ := tr.RootHash()
	storageManager.TakeSnapshot(rootHash, true)
	time.Sleep(snapshotDelay)

	err := storageManager.Database().Remove(rootHash)
	assert.Nil(t, err)

	val, err := storageManager.Database().Get(rootHash)
	assert.Nil(t, val)
	assert.NotNil(t, err)

	newTrie, err := tr.Recreate(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, newTrie)
}

func TestEachSnapshotCreatesOwnDatabase(t *testing.T) {
	t.Parallel()

	testVals := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("doe"), []byte("reindeer")},
		{[]byte("dog"), []byte("puppy")},
		{[]byte("dogglesworth"), []byte("cat")},
	}

	tr, trieStorage := newEmptyTrie()

	for _, testVal := range testVals {
		_ = tr.Update(testVal.key, testVal.value)
		_ = tr.Commit()
		trieStorage.TakeSnapshot(tr.root.getHash(), true)
		time.Sleep(snapshotDelay)

		trieStorage.storageOperationMutex.Lock()
		snapshotId := strconv.Itoa(trieStorage.snapshotId - 1)
		snapshotPath := path.Join(trieStorage.snapshotDbCfg.FilePath, snapshotId)
		trieStorage.storageOperationMutex.Unlock()
		f, _ := os.Stat(snapshotPath)
		assert.True(t, f.IsDir())
	}

	trieStorage.storageOperationMutex.Lock()
	assert.Equal(t, len(testVals), trieStorage.snapshotId)
	trieStorage.storageOperationMutex.Unlock()
}

func TestDeleteOldSnapshots(t *testing.T) {
	t.Parallel()

	testVals := []struct {
		key   []byte
		value []byte
	}{
		{[]byte("doe"), []byte("reindeer")},
		{[]byte("dog"), []byte("puppy")},
		{[]byte("dogglesworth"), []byte("cat")},
		{[]byte("horse"), []byte("mustang")},
	}

	tr, trieStorage := newEmptyTrie()

	for _, testVal := range testVals {
		_ = tr.Update(testVal.key, testVal.value)
		_ = tr.Commit()
		trieStorage.TakeSnapshot(tr.root.getHash(), true)
	}
	time.Sleep(snapshotDelay)

	snapshots, _ := ioutil.ReadDir(trieStorage.snapshotDbCfg.FilePath)
	assert.Equal(t, 2, len(snapshots))
	assert.Equal(t, "2", snapshots[0].Name())
	assert.Equal(t, "3", snapshots[1].Name())
}

func TestTrieCheckpoint(t *testing.T) {
	t.Parallel()

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}

	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()
	trieStorage.TakeSnapshot(tr.root.getHash(), true)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("doge"), []byte("reindeer"))
	newHashes, _ := tr.GetDirtyHashes()
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	val, err := tr.Get([]byte("doge"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("reindeer"), val)

	args := getNewTrieStorageManagerArgs()
	args.DB = trieStorage.snapshots[0]
	args.GeneralConfig = generalCfg
	args.CheckpointHashesHolder = hashesHolder.NewCheckpointHashesHolder(checkpointHashesHolderMaxSize, hashSize)
	snapshotTrieStorage, _ := NewTrieStorageManager(args)
	collapsedRoot, _ := tr.root.getCollapsed()
	snapshotTrie := &patriciaMerkleTrie{
		root:        collapsedRoot,
		trieStorage: snapshotTrieStorage,
		marshalizer: tr.marshalizer,
		hasher:      tr.hasher,
	}
	trieStorage.AddDirtyCheckpointHashes(rootHash, newHashes)

	val, err = snapshotTrie.Get([]byte("doge"))
	assert.NotNil(t, err)
	assert.Nil(t, val)

	trieStorage.SetCheckpoint(tr.root.getHash())
	time.Sleep(snapshotDelay)

	val, err = snapshotTrie.Get([]byte("doge"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("reindeer"), val)
}

func TestTrieCheckpointWithNoSnapshotCreatesSnapshot(t *testing.T) {
	t.Parallel()

	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	assert.Equal(t, 0, len(trieStorage.snapshots))

	_ = tr.Commit()
	trieStorage.SetCheckpoint(tr.root.getHash())
	time.Sleep(snapshotDelay)

	trieStorage.storageOperationMutex.Lock()
	assert.Equal(t, 1, len(trieStorage.snapshots))
	trieStorage.storageOperationMutex.Unlock()
}

func TestTrieSnapshottingAndCheckpointConcurrently(t *testing.T) {
	t.Parallel()

	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()

	trieStorage.TakeSnapshot(tr.root.getHash(), true)
	time.Sleep(snapshotDelay)

	numSnapshots := 5
	numCheckpoints := 5
	totalNumSnapshot := numSnapshots + 1

	var snapshotWg sync.WaitGroup
	var checkpointWg sync.WaitGroup
	mut := sync.Mutex{}
	snapshotWg.Add(numSnapshots)
	checkpointWg.Add(numCheckpoints)

	for i := 0; i < numSnapshots; i++ {
		go func(j int) {
			mut.Lock()
			_ = tr.Update([]byte(strconv.Itoa(j)), []byte(strconv.Itoa(j)))
			_ = tr.Commit()
			rootHash, _ := tr.RootHash()
			trieStorage.TakeSnapshot(rootHash, true)
			mut.Unlock()
			snapshotWg.Done()
		}(i)
	}

	for i := 0; i < numCheckpoints; i++ {
		go func(j int) {
			mut.Lock()
			_ = tr.Update([]byte(strconv.Itoa(j+numSnapshots)), []byte(strconv.Itoa(j+numSnapshots)))
			newHashes, _ := tr.GetDirtyHashes()
			_ = tr.Commit()
			rootHash, _ := tr.RootHash()

			trieStorage.AddDirtyCheckpointHashes(rootHash, newHashes)
			trieStorage.SetCheckpoint(rootHash)
			mut.Unlock()
			checkpointWg.Done()
		}(i)
	}

	snapshotWg.Wait()
	checkpointWg.Wait()
	time.Sleep(snapshotDelay * 3)

	trieStorage.storageOperationMutex.Lock()
	assert.Equal(t, totalNumSnapshot, trieStorage.snapshotId)

	lastSnapshot := len(trieStorage.snapshots) - 1
	val, err := trieStorage.snapshots[lastSnapshot].Get(tr.root.getHash())
	trieStorage.storageOperationMutex.Unlock()
	assert.NotNil(t, val)
	assert.Nil(t, err)
}

func TestIsPresentInLastSnapshotDbDoesNotPanicIfNoSnapshot(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	assert.False(t, trieStorage.isPresentInLastSnapshotDb([]byte("rootHash")))
}

func TestIsPresentInLastSnapshotDb(t *testing.T) {
	t.Parallel()

	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	_ = tr.Commit()
	rootHash1, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash1, true)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true)
	time.Sleep(snapshotDelay)

	trieStorage.storageOperationMutex.Lock()
	val, err := trieStorage.snapshots[0].Get(rootHash2)
	assert.Nil(t, val)
	assert.NotNil(t, err)

	val, err = trieStorage.snapshots[1].Get(rootHash2)
	assert.Nil(t, err)
	assert.NotNil(t, val)
	trieStorage.storageOperationMutex.Unlock()

	assert.True(t, trieStorage.isPresentInLastSnapshotDb(rootHash2))
}

func TestTrieSnapshotChecksOnlyLastSnapshotDbForTheHash(t *testing.T) {
	t.Parallel()

	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	_ = tr.Commit()
	rootHash1, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash1, true)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true)
	time.Sleep(snapshotDelay)

	trieStorage.storageOperationMutex.Lock()
	val, err := trieStorage.snapshots[0].Get(rootHash1)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	val, err = trieStorage.snapshots[1].Get(rootHash1)
	assert.Nil(t, val)
	assert.NotNil(t, err)
	trieStorage.storageOperationMutex.Unlock()

	assert.False(t, trieStorage.isPresentInLastSnapshotDb(rootHash1))
}

func TestShouldNotRemoveSnapshotDbIfItIsStillInUse(t *testing.T) {
	t.Parallel()

	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	_ = tr.Commit()
	rootHash1, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash1, true)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true)
	time.Sleep(snapshotDelay)

	db := trieStorage.GetSnapshotThatContainsHash(rootHash1)

	_ = tr.Update([]byte("dog"), []byte("pup"))

	_ = tr.Commit()
	rootHash3, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash3, true)
	time.Sleep(snapshotDelay)

	val, err := db.Get(rootHash1)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	db.DecreaseNumReferences()

	val, err = db.Get(rootHash1)
	assert.Nil(t, val)
	assert.NotNil(t, err)
}

func TestShouldNotRemoveSnapshotDbsIfKeepSnapshotsTrue(t *testing.T) {
	t.Parallel()
	nrOfSnapshots := 5
	tr, trieStorage := newEmptyTrie()
	trieStorage.keepSnapshots = true

	for i := 0; i < nrOfSnapshots; i++ {
		key := strconv.Itoa(i) + "doe"
		value := strconv.Itoa(i) + "reindeer"
		_ = tr.Update([]byte(key), []byte(value))
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()
		trieStorage.TakeSnapshot(rootHash, true)
		time.Sleep(snapshotDelay)
	}

	for i := 0; i < nrOfSnapshots; i++ {
		snapshotPath := path.Join(trieStorage.snapshotDbCfg.FilePath, strconv.Itoa(i))
		folderInfo, err := os.Stat(snapshotPath)

		assert.NotNil(t, folderInfo)
		assert.Nil(t, err)
	}

	err := os.RemoveAll(trieStorage.snapshotDbCfg.FilePath)
	assert.Nil(t, err)
}

func TestShouldRemoveSnapshotDbsIfKeepSnapshotsFalse(t *testing.T) {
	t.Parallel()
	nrOfSnapshots := 5
	tr, trieStorage := newEmptyTrie()
	trieStorage.keepSnapshots = false

	for i := 0; i < nrOfSnapshots; i++ {
		key := strconv.Itoa(i) + "doe"
		value := strconv.Itoa(i) + "reindeer"
		_ = tr.Update([]byte(key), []byte(value))
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()
		trieStorage.TakeSnapshot(rootHash, true)
		time.Sleep(snapshotDelay)
	}

	for i := 0; i < nrOfSnapshots-int(trieStorage.maxSnapshots); i++ {
		snapshotPath := path.Join(trieStorage.snapshotDbCfg.FilePath, strconv.Itoa(i))
		folderInfo, err := os.Stat(snapshotPath)
		assert.Nil(t, folderInfo)
		assert.NotNil(t, err)
	}
	for i := nrOfSnapshots - int(trieStorage.maxSnapshots); i < nrOfSnapshots; i++ {
		snapshotPath := path.Join(trieStorage.snapshotDbCfg.FilePath, strconv.Itoa(i))
		folderInfo, err := os.Stat(snapshotPath)
		assert.NotNil(t, folderInfo)
		assert.Nil(t, err)

	}

	err := os.RemoveAll(trieStorage.snapshotDbCfg.FilePath)
	assert.Nil(t, err)
}

func TestShouldNotDisconnectSnapshotDbIfItIsStillInUse(t *testing.T) {
	t.Parallel()

	tr, trieStorage := newEmptyTrie()
	trieStorage.keepSnapshots = true
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	_ = tr.Commit()
	rootHash1, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash1, true)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true)
	time.Sleep(snapshotDelay)

	db := trieStorage.GetSnapshotThatContainsHash(rootHash1)

	_ = tr.Update([]byte("dog"), []byte("pup"))

	_ = tr.Commit()
	rootHash3, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash3, true)
	time.Sleep(snapshotDelay)

	val, err := db.Get(rootHash1)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	db.DecreaseNumReferences()

	val, err = db.Get(rootHash1)
	assert.Nil(t, val)
	assert.NotNil(t, err)
}
