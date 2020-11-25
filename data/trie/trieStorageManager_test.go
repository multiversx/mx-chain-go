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
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

const pruningDelay = time.Second
const snapshotDelay = time.Second

func TestNewTrieStorageManagerNilDb(t *testing.T) {
	t.Parallel()

	ts, err := NewTrieStorageManager(nil, &mock.MarshalizerMock{}, &mock.HasherMock{}, config.DBConfig{}, &mock.EvictionWaitingList{}, config.TrieStorageManagerConfig{})
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilDatabase, err)
}

func TestNewTrieStorageManagerNilMarshalizer(t *testing.T) {
	t.Parallel()

	ts, err := NewTrieStorageManager(mock.NewMemDbMock(), nil, &mock.HasherMock{}, config.DBConfig{}, &mock.EvictionWaitingList{}, config.TrieStorageManagerConfig{})
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestNewTrieStorageManagerNilHasher(t *testing.T) {
	t.Parallel()

	ts, err := NewTrieStorageManager(mock.NewMemDbMock(), &mock.MarshalizerMock{}, nil, config.DBConfig{}, &mock.EvictionWaitingList{}, config.TrieStorageManagerConfig{})
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilHasher, err)
}

func TestNewTrieStorageManagerNilEwlAndPruningEnabled(t *testing.T) {
	t.Parallel()

	ts, err := NewTrieStorageManager(mock.NewMemDbMock(), &mock.MarshalizerMock{}, &mock.HasherMock{}, config.DBConfig{}, nil, config.TrieStorageManagerConfig{})
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilEvictionWaitingList, err)
}

func TestNewTrieStorageManagerOkVals(t *testing.T) {
	t.Parallel()

	ts, err := NewTrieStorageManager(mock.NewMemDbMock(), &mock.MarshalizerMock{}, &mock.HasherMock{}, config.DBConfig{}, &mock.EvictionWaitingList{}, config.TrieStorageManagerConfig{})
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

	db := mock.NewMemDbMock()
	msh, hsh := getTestMarshAndHasher()
	size := uint(100)
	evictionWaitList, _ := mock.NewEvictionWaitingList(size, mock.NewMemDbMock(), msh)
	trieStorage, _ := NewTrieStorageManager(db, msh, hsh, cfg, evictionWaitList, generalCfg)
	maxTrieLevelInMemory := uint(5)
	tr, _ := NewTrie(trieStorage, msh, hsh, maxTrieLevelInMemory)

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	rootHash, _ := tr.Root()
	tr.TakeSnapshot(rootHash)
	time.Sleep(snapshotDelay)

	trieStorage.storageOperationMutex.Lock()
	_ = trieStorage.snapshots[0].Close()
	trieStorage.storageOperationMutex.Unlock()

	newTrieStorage, _ := NewTrieStorageManager(memorydb.New(), msh, hsh, cfg, evictionWaitList, generalCfg)
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

	db := mock.NewMemDbMock()
	msh, hsh := getTestMarshAndHasher()
	size := uint(100)
	evictionWaitList, _ := mock.NewEvictionWaitingList(size, mock.NewMemDbMock(), msh)
	trieStorage, _ := NewTrieStorageManager(db, msh, hsh, cfg, evictionWaitList, generalCfg)
	maxTrieLevelInMemory := uint(5)
	tr, _ := NewTrie(trieStorage, msh, hsh, maxTrieLevelInMemory)

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	rootHash, _ := tr.Root()
	tr.TakeSnapshot(rootHash)
	time.Sleep(snapshotDelay)

	numSnapshots := 10
	for i := 0; i < numSnapshots; i++ {
		_ = tr.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		_ = tr.Commit()
		rootHash, _ = tr.Root()
		tr.TakeSnapshot(rootHash)
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

	newTrieStorage, _ := NewTrieStorageManager(memorydb.New(), msh, hsh, cfg, evictionWaitList, generalCfg)

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

func TestTrieDatabasePruning(t *testing.T) {
	t.Parallel()

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	db := mock.NewMemDbMock()
	msh, hsh := getTestMarshAndHasher()
	size := uint(1)
	evictionWaitList, _ := mock.NewEvictionWaitingList(size, mock.NewMemDbMock(), msh)
	trieStorage, _ := NewTrieStorageManager(db, msh, hsh, config.DBConfig{}, evictionWaitList, generalCfg)

	tr := &patriciaMerkleTrie{
		trieStorage: trieStorage,
		oldHashes:   make([][]byte, 0),
		oldRoot:     make([]byte, 0),
		marshalizer: msh,
		hasher:      hsh,
	}

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.Commit()

	key := []byte{7, 6, 15, 6, 4, 6, 16}
	oldHashes := make([][]byte, 0, 4)
	n := tr.root
	rootHash, _ := tr.Root()
	oldHashes = append(oldHashes, rootHash)

	for i := 0; i < 3; i++ {
		n, key, _ = n.getNext(key, db)
		oldHashes = append(oldHashes, n.getHash())
	}

	_ = tr.Update([]byte("dog"), []byte("doee"))
	_ = tr.Commit()

	tr.CancelPrune(rootHash, data.NewRoot)
	tr.Prune(rootHash, data.OldRoot)
	time.Sleep(pruningDelay)

	for i := range oldHashes {
		encNode, err := tr.Database().Get(oldHashes[i])
		assert.Nil(t, encNode)
		assert.NotNil(t, err)
	}
}

func TestRecreateTrieFromSnapshotDb(t *testing.T) {
	t.Parallel()

	tr := initTrie()
	_ = tr.Commit()
	rootHash, _ := tr.Root()
	tr.TakeSnapshot(rootHash)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("doge"), []byte("doge"))
	_ = tr.Commit()

	tr.CancelPrune(rootHash, data.NewRoot)
	tr.Prune(rootHash, data.OldRoot)
	time.Sleep(pruningDelay)

	val, err := tr.Database().Get(rootHash)
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

	tr, trieStorage, _ := newEmptyTrie()

	for _, testVal := range testVals {
		_ = tr.Update(testVal.key, testVal.value)
		_ = tr.Commit()
		tr.TakeSnapshot(tr.root.getHash())
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

	tr, trieStorage, _ := newEmptyTrie()

	for _, testVal := range testVals {
		_ = tr.Update(testVal.key, testVal.value)
		_ = tr.Commit()
		tr.TakeSnapshot(tr.root.getHash())
	}
	time.Sleep(snapshotDelay)

	snapshots, _ := ioutil.ReadDir(trieStorage.snapshotDbCfg.FilePath)
	assert.Equal(t, 2, len(snapshots))
	assert.Equal(t, "2", snapshots[0].Name())
	assert.Equal(t, "3", snapshots[1].Name())
}

func TestPruningIsDoneAfterSnapshotIsFinished(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()
	rootHash := tr.root.getHash()
	tr.CancelPrune(rootHash, data.NewRoot)

	tr.TakeSnapshot(rootHash)
	time.Sleep(snapshotDelay)
	tr.Prune(rootHash, data.NewRoot)
	time.Sleep(pruningDelay)

	val, err := trieStorage.snapshots[0].Get(rootHash)
	assert.NotNil(t, val)
	assert.Nil(t, err)
}

func TestTrieCheckpoint(t *testing.T) {
	t.Parallel()

	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()
	tr.TakeSnapshot(tr.root.getHash())
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("doge"), []byte("reindeer"))
	_ = tr.Commit()

	val, err := tr.Get([]byte("doge"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("reindeer"), val)

	snapshotTrieStorage, _ := NewTrieStorageManager(trieStorage.snapshots[0], tr.marshalizer, tr.hasher, config.DBConfig{}, &mock.EvictionWaitingList{}, generalCfg)
	collapsedRoot, _ := tr.root.getCollapsed()
	snapshotTrie := &patriciaMerkleTrie{
		root:        collapsedRoot,
		trieStorage: snapshotTrieStorage,
		marshalizer: tr.marshalizer,
		hasher:      tr.hasher,
	}

	val, err = snapshotTrie.Get([]byte("doge"))
	assert.NotNil(t, err)
	assert.Nil(t, val)

	tr.SetCheckpoint(tr.root.getHash())
	time.Sleep(snapshotDelay)

	val, err = snapshotTrie.Get([]byte("doge"))
	assert.Nil(t, err)
	assert.Equal(t, []byte("reindeer"), val)
}

func TestTrieCheckpointWithNoSnapshotCreatesSnapshot(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	assert.Equal(t, 0, len(trieStorage.snapshots))

	_ = tr.Commit()
	tr.SetCheckpoint(tr.root.getHash())
	time.Sleep(snapshotDelay)

	trieStorage.storageOperationMutex.Lock()
	assert.Equal(t, 1, len(trieStorage.snapshots))
	trieStorage.storageOperationMutex.Unlock()
}

func TestTrieSnapshottingAndCheckpointConcurrently(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()

	tr.TakeSnapshot(tr.root.getHash())
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
			rootHash, _ := tr.Root()
			tr.TakeSnapshot(rootHash)
			mut.Unlock()
			snapshotWg.Done()
		}(i)
	}

	for i := 0; i < numCheckpoints; i++ {
		go func(j int) {
			mut.Lock()
			_ = tr.Update([]byte(strconv.Itoa(j+numSnapshots)), []byte(strconv.Itoa(j+numSnapshots)))
			_ = tr.Commit()
			rootHash, _ := tr.Root()
			tr.SetCheckpoint(rootHash)
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

func TestTriePruneAndCancelPruneWhileSnapshotInProgressAddsToPruningBuffer(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	oldRootHash, _ := tr.Root()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_ = tr.Commit()
	newRootHash, _ := tr.Root()

	tr.EnterPruningBufferingMode()
	tr.Prune(oldRootHash, data.OldRoot)
	tr.CancelPrune(newRootHash, data.NewRoot)
	tr.ExitPruningBufferingMode()

	assert.Equal(t, 2, trieStorage.pruningBuffer.len())
}

func TestTriePruneOnRollbackWhileSnapshotInProgressCancelsPrune(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	oldRootHash, _ := tr.Root()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_ = tr.Commit()
	newRootHash, _ := tr.Root()

	tr.EnterPruningBufferingMode()
	tr.CancelPrune(oldRootHash, data.OldRoot)
	tr.Prune(newRootHash, data.NewRoot)
	tr.ExitPruningBufferingMode()

	assert.Equal(t, 1, trieStorage.pruningBuffer.len())
}

func TestTriePruneAfterSnapshotIsDonePrunesBufferedHashes(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	newHashes, _ := tr.GetDirtyHashes()
	tr.SetNewHashes(newHashes)
	_ = tr.Commit()
	oldRootHash, _ := tr.Root()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	newHashes, _ = tr.GetDirtyHashes()
	tr.SetNewHashes(newHashes)
	_ = tr.Commit()
	newRootHash, _ := tr.Root()

	tr.EnterPruningBufferingMode()
	tr.Prune(oldRootHash, data.OldRoot)
	tr.CancelPrune(newRootHash, data.NewRoot)
	tr.ExitPruningBufferingMode()

	tr.Prune(oldRootHash, data.NewRoot)

	assert.Equal(t, 0, trieStorage.pruningBuffer.len())
}

func TestTrieCancelPruneAndPruningBufferNotEmptyAddsToPruningBuffer(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	oldRootHash, _ := tr.Root()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_ = tr.Commit()
	newRootHash, _ := tr.Root()

	tr.EnterPruningBufferingMode()
	tr.Prune(oldRootHash, data.OldRoot)
	tr.CancelPrune(newRootHash, data.NewRoot)
	tr.ExitPruningBufferingMode()

	tr.CancelPrune(oldRootHash, data.NewRoot)

	assert.Equal(t, 3, trieStorage.pruningBuffer.len())
}

func TestTriePruneAndCancelPruneAddedToBufferInOrder(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_ = tr.Commit()
	oldRootHash, _ := tr.Root()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_ = tr.Commit()
	newRootHash, _ := tr.Root()

	tr.EnterPruningBufferingMode()
	tr.Prune(oldRootHash, data.OldRoot)
	tr.CancelPrune(newRootHash, data.NewRoot)
	tr.ExitPruningBufferingMode()

	tr.CancelPrune(oldRootHash, data.NewRoot)

	bufferedHashes := trieStorage.pruningBuffer.removeAll()

	expectedHash := append(oldRootHash, byte(data.OldRoot))
	assert.Equal(t, append(expectedHash, byte(prune)), bufferedHashes[0])

	expectedHash = append(newRootHash, byte(data.NewRoot))
	assert.Equal(t, append(expectedHash, byte(cancelPrune)), bufferedHashes[1])

	expectedHash = append(oldRootHash, byte(data.NewRoot))
	assert.Equal(t, append(expectedHash, byte(cancelPrune)), bufferedHashes[2])
}

func TestIsPresentInLastSnapshotDbDoesNotPanicIfNoSnapshot(t *testing.T) {
	t.Parallel()

	_, trieStorage, _ := newEmptyTrie()
	assert.False(t, trieStorage.isPresentInLastSnapshotDb([]byte("rootHash")))
}

func TestIsPresentInLastSnapshotDb(t *testing.T) {
	t.Parallel()

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	_ = tr.Commit()
	rootHash1, _ := tr.Root()
	tr.TakeSnapshot(rootHash1)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.Root()
	tr.TakeSnapshot(rootHash2)
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

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	_ = tr.Commit()
	rootHash1, _ := tr.Root()
	tr.TakeSnapshot(rootHash1)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.Root()
	tr.TakeSnapshot(rootHash2)
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

	tr, trieStorage, _ := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))

	_ = tr.Commit()
	rootHash1, _ := tr.Root()
	tr.TakeSnapshot(rootHash1)
	time.Sleep(snapshotDelay)

	_ = tr.Update([]byte("dog"), []byte("puppy"))

	_ = tr.Commit()
	rootHash2, _ := tr.Root()
	tr.TakeSnapshot(rootHash2)
	time.Sleep(snapshotDelay)

	db := trieStorage.GetSnapshotThatContainsHash(rootHash1)

	_ = tr.Update([]byte("dog"), []byte("pup"))

	_ = tr.Commit()
	rootHash3, _ := tr.Root()
	tr.TakeSnapshot(rootHash3)
	time.Sleep(snapshotDelay)

	val, err := db.Get(rootHash1)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	db.DecreaseNumReferences()

	val, err = db.Get(rootHash1)
	assert.Nil(t, val)
	assert.NotNil(t, err)
}
