package trie

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/stretchr/testify/assert"
)

const (
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
	trieStorage.TakeSnapshot(rootHash, true, nil)
	WaitForOperationToComplete(trieStorage)

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
	foundSnapshot := newTrieStorage.GetSnapshotThatContainsHash(rootHash)
	assert.NotNil(t, foundSnapshot)
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
	trieStorage.TakeSnapshot(rootHash, true, nil)
	WaitForOperationToComplete(trieStorage)

	numSnapshots := 10
	for i := 0; i < numSnapshots; i++ {
		_ = tr.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		_ = tr.Commit()
		rootHash, _ = tr.RootHash()
		trieStorage.TakeSnapshot(rootHash, true, nil)
		WaitForOperationToComplete(trieStorage)
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
	storageManager.TakeSnapshot(rootHash, true, nil)
	WaitForOperationToComplete(storageManager)

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
		trieStorage.TakeSnapshot(tr.root.getHash(), true, nil)
		WaitForOperationToComplete(trieStorage)

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
		trieStorage.TakeSnapshot(tr.root.getHash(), true, nil)
	}
	WaitForOperationToComplete(trieStorage)

	snapshots, _ := ioutil.ReadDir(trieStorage.snapshotDbCfg.FilePath)
	assert.Equal(t, 2, len(snapshots))
	assert.Equal(t, "2", snapshots[0].Name())
	assert.Equal(t, "3", snapshots[1].Name())
}

func createSmallTestTrieAndStorageManager() (*patriciaMerkleTrie, *trieStorageManager) {
	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()
	trieStorage.TakeSnapshot(tr.root.getHash(), true, nil)
	WaitForOperationToComplete(trieStorage)

	return tr, trieStorage
}

func TestTrieCheckpoint(t *testing.T) {
	t.Parallel()

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}

	tr, trieStorage := createSmallTestTrieAndStorageManager()

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

	trieStorage.SetCheckpoint(tr.root.getHash(), nil)
	WaitForOperationToComplete(trieStorage)

	val, err = snapshotTrie.Get([]byte("doge"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("reindeer"), val)
}

func TestTrieCheckpoint_IsInLastSnapshot(t *testing.T) {
	t.Parallel()

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}

	tr, trieStorage := createSmallTestTrieAndStorageManager()

	_ = tr.Update([]byte("doge"), []byte("reindeer"))
	newHashes, _ := tr.GetDirtyHashes()
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	val, err := tr.Get([]byte("doge"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("reindeer"), val)

	args := getNewTrieStorageManagerArgs()
	args.DB = trieStorage.snapshots[0]
	_ = trieStorage.snapshots[0].Put(rootHash, rootHash)
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

	trieStorage.SetCheckpoint(tr.root.getHash(), nil)
	WaitForOperationToComplete(trieStorage)

	val, err = snapshotTrie.Get([]byte("doge"))
	assert.NotNil(t, err)
	assert.Nil(t, val)
}

func TestTrieCheckpointWithNoSnapshotCreatesSnapshot(t *testing.T) {
	t.Parallel()

	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	assert.Equal(t, 0, len(trieStorage.snapshots))

	_ = tr.Commit()
	trieStorage.SetCheckpoint(tr.root.getHash(), nil)
	WaitForOperationToComplete(trieStorage)

	trieStorage.storageOperationMutex.Lock()
	assert.Equal(t, 1, len(trieStorage.snapshots))
	trieStorage.storageOperationMutex.Unlock()
}

func TestTrieSnapshottingCheckpointAndClose(t *testing.T) {
	t.Parallel()

	tr, trieStorage := createSmallTestTrieAndStorageManager()

	numSnapshotsAndCheckpoints := 5
	totalNumSnapshot := numSnapshotsAndCheckpoints + 1

	for i := 0; i < numSnapshotsAndCheckpoints; i++ {
		hash := []byte(strconv.Itoa(i + numSnapshotsAndCheckpoints))
		doCheckpoint(tr, trieStorage, hash)

		hash = []byte(strconv.Itoa(i))
		doSnapshot(tr, trieStorage, hash)
	}

	WaitForOperationToComplete(trieStorage)

	trieStorage.storageOperationMutex.Lock()
	assert.Equal(t, totalNumSnapshot, trieStorage.snapshotId)

	lastSnapshot := len(trieStorage.snapshots) - 1
	val, err := trieStorage.snapshots[lastSnapshot].Get(tr.root.getHash())
	trieStorage.storageOperationMutex.Unlock()
	assert.NotNil(t, val)
	assert.Nil(t, err)

	err = trieStorage.Close()
	assert.Nil(t, err)
}

func TestTrieSnapshottingFromCheckpoint(t *testing.T) {
	t.Parallel()

	tr, trieStorage := createSmallTestTrieAndStorageManager()

	totalNumSnapshot := 2

	roothash, _ := tr.RootHash()
	trieStorage.TakeSnapshot(roothash, true, nil)

	hash := []byte(strconv.Itoa(0))
	doCheckpoint(tr, trieStorage, hash)

	WaitForOperationToComplete(trieStorage)

	trieStorage.TakeSnapshot(hash, true, nil)
	WaitForOperationToComplete(trieStorage)

	trieStorage.storageOperationMutex.Lock()
	assert.Equal(t, totalNumSnapshot, trieStorage.snapshotId)

	lastSnapshot := len(trieStorage.snapshots) - 1
	val, err := trieStorage.snapshots[lastSnapshot].Get(tr.root.getHash())
	trieStorage.storageOperationMutex.Unlock()
	assert.NotNil(t, val)
	assert.Nil(t, err)

	err = trieStorage.Close()
	assert.Nil(t, err)
}

func doCheckpoint(tr temporary.Trie, trieStorage temporary.StorageManager, hash []byte) {
	_ = tr.Update(hash, hash)
	newHashes, _ := tr.GetDirtyHashes()
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()

	trieStorage.AddDirtyCheckpointHashes(rootHash, newHashes)
	trieStorage.SetCheckpoint(rootHash, nil)
}

func doSnapshot(tr temporary.Trie, trieStorage temporary.StorageManager, hash []byte) {
	_ = tr.Update(hash, hash)
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash, true, nil)
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
	trieStorage.TakeSnapshot(rootHash1, true, nil)
	WaitForOperationToComplete(trieStorage)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true, nil)
	WaitForOperationToComplete(trieStorage)

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
	trieStorage.TakeSnapshot(rootHash1, true, nil)
	WaitForOperationToComplete(trieStorage)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true, nil)
	WaitForOperationToComplete(trieStorage)

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
	trieStorage.TakeSnapshot(rootHash1, true, nil)
	WaitForOperationToComplete(trieStorage)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true, nil)
	WaitForOperationToComplete(trieStorage)

	db := trieStorage.GetSnapshotThatContainsHash(rootHash1)

	_ = tr.Update([]byte("dog"), []byte("pup"))

	_ = tr.Commit()
	rootHash3, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash3, true, nil)
	WaitForOperationToComplete(trieStorage)

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
		trieStorage.TakeSnapshot(rootHash, true, nil)
		WaitForOperationToComplete(trieStorage)
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
		trieStorage.TakeSnapshot(rootHash, true, nil)
		WaitForOperationToComplete(trieStorage)
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
	trieStorage.TakeSnapshot(rootHash1, true, nil)
	WaitForOperationToComplete(trieStorage)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash2, true, nil)
	WaitForOperationToComplete(trieStorage)

	db := trieStorage.GetSnapshotThatContainsHash(rootHash1)

	_ = tr.Update([]byte("dog"), []byte("pup"))

	_ = tr.Commit()
	rootHash3, _ := tr.RootHash()
	trieStorage.TakeSnapshot(rootHash3, true, nil)
	WaitForOperationToComplete(trieStorage)

	val, err := db.Get(rootHash1)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	db.DecreaseNumReferences()

	val, err = db.Get(rootHash1)
	assert.Nil(t, val)
	assert.NotNil(t, err)
}

func TestTrieStorageManager_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := NewTrieStorageManager(args)

	assert.True(t, ts.IsPruningEnabled())
}

func TestTrieStorageManager_IsPruningBlocked(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := NewTrieStorageManager(args)

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
	ts, _ := NewTrieStorageManager(args)

	assert.Equal(t, batchDelay, ts.GetSnapshotDbBatchDelay())
}

func TestTrieStorageManager_Remove(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := NewTrieStorageManager(args)

	key := []byte("key")
	value := []byte("value")

	_ = args.DB.Put(key, value)
	hashes := make(temporary.ModifiedHashes)
	hashes[string(value)] = struct{}{}
	hashes[string(key)] = struct{}{}
	_ = args.CheckpointHashesHolder.Put(key, hashes)

	val, err := args.DB.Get(key)
	assert.Nil(t, err)
	assert.NotNil(t, val)
	ok := args.CheckpointHashesHolder.ShouldCommit(key)
	assert.True(t, ok)

	err = ts.Remove(key)
	assert.Nil(t, err)

	val, err = args.DB.Get(key)
	assert.Nil(t, val)
	assert.NotNil(t, err)
	ok = args.CheckpointHashesHolder.ShouldCommit(key)
	assert.False(t, ok)
}
