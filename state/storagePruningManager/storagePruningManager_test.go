package storagePruningManager

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/stretchr/testify/assert"
)

const trieDbOperationDelay = time.Second

func getDefaultTrieAndAccountsDbAndStoragePruningManager() (temporary.Trie, *state.AccountsDB, *storagePruningManager) {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	marshalizer := &mock.MarshalizerMock{}
	hsh := &hashingMocks.HasherMock{}
	args := trie.NewTrieStorageManagerArgs{
		DB:          mock.NewMemDbMock(),
		Marshalizer: marshalizer,
		Hasher:      hsh,
		SnapshotDbConfig: config.DBConfig{
			Type: "MemoryDB",
		},
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize),
	}
	trieStorage, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(trieStorage, marshalizer, hsh, 5)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, mock.NewMemDbMock(), marshalizer)
	spm, _ := NewStoragePruningManager(ewl, generalCfg.PruningBufferLen)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator(), spm)

	return tr, adb, spm
}

func TestAccountsDB_PruningIsDoneAfterSnapshotIsFinished(t *testing.T) {
	t.Parallel()

	tr, adb, spm := getDefaultTrieAndAccountsDbAndStoragePruningManager()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_, _ = adb.Commit()
	rootHash, _ := tr.RootHash()

	trieStorage := tr.GetStorageManager()
	trieStorage.TakeSnapshot(rootHash, true, nil)
	time.Sleep(trieDbOperationDelay)
	spm.PruneTrie(rootHash, temporary.NewRoot, trieStorage)
	time.Sleep(trieDbOperationDelay)

	val, err := trieStorage.Database().Get(rootHash)
	assert.Nil(t, val)
	assert.NotNil(t, err)

	snapshotDb := trieStorage.GetSnapshotThatContainsHash(rootHash)
	assert.NotNil(t, snapshotDb)

	val, err = snapshotDb.Get(rootHash)
	assert.NotNil(t, val)
	assert.Nil(t, err)
}

func TestAccountsDB_TriePruneAndCancelPruneWhileSnapshotInProgressAddsToPruningBuffer(t *testing.T) {
	t.Parallel()

	tr, adb, spm := getDefaultTrieAndAccountsDbAndStoragePruningManager()
	trieStorage := tr.GetStorageManager()

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_, _ = adb.Commit()
	oldRootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_, _ = adb.Commit()
	newRootHash, _ := tr.RootHash()

	trieStorage.EnterPruningBufferingMode()
	spm.PruneTrie(oldRootHash, temporary.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, temporary.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()

	assert.Equal(t, 2, spm.pruningBuffer.Len())
}

func TestAccountsDB_TriePruneOnRollbackWhileSnapshotInProgressCancelsPrune(t *testing.T) {
	t.Parallel()

	tr, adb, spm := getDefaultTrieAndAccountsDbAndStoragePruningManager()
	trieStorage := tr.GetStorageManager()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_, _ = adb.Commit()
	oldRootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_, _ = adb.Commit()
	newRootHash, _ := tr.RootHash()

	trieStorage.EnterPruningBufferingMode()
	spm.CancelPrune(oldRootHash, temporary.OldRoot, trieStorage)
	spm.PruneTrie(newRootHash, temporary.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()

	assert.Equal(t, 1, spm.pruningBuffer.Len())
}

func TestAccountsDB_TriePruneAfterSnapshotIsDonePrunesBufferedHashes(t *testing.T) {
	t.Parallel()

	tr, adb, spm := getDefaultTrieAndAccountsDbAndStoragePruningManager()
	trieStorage := tr.GetStorageManager()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_, _ = adb.Commit()
	oldRootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_, _ = adb.Commit()
	newRootHash, _ := tr.RootHash()

	trieStorage.EnterPruningBufferingMode()
	spm.PruneTrie(oldRootHash, temporary.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, temporary.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()
	assert.Equal(t, 2, spm.pruningBuffer.Len())

	adb.PruneTrie(oldRootHash, temporary.NewRoot)
	assert.Equal(t, 0, spm.pruningBuffer.Len())
}

func TestAccountsDB_TrieCancelPruneAndPruningBufferNotEmptyAddsToPruningBuffer(t *testing.T) {
	t.Parallel()

	tr, adb, spm := getDefaultTrieAndAccountsDbAndStoragePruningManager()
	trieStorage := tr.GetStorageManager()

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_, _ = adb.Commit()
	oldRootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_, _ = adb.Commit()
	newRootHash, _ := tr.RootHash()

	trieStorage.EnterPruningBufferingMode()
	spm.PruneTrie(oldRootHash, temporary.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, temporary.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()
	assert.Equal(t, 2, spm.pruningBuffer.Len())

	adb.CancelPrune(oldRootHash, temporary.NewRoot)
	assert.Equal(t, 3, spm.pruningBuffer.Len())
}

func TestAccountsDB_TriePruneAndCancelPruneAddedToBufferInOrder(t *testing.T) {
	t.Parallel()

	tr, adb, spm := getDefaultTrieAndAccountsDbAndStoragePruningManager()
	trieStorage := tr.GetStorageManager()

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))
	_, _ = adb.Commit()
	oldRootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dogglesworth"), []byte("catnip"))
	_, _ = adb.Commit()
	newRootHash, _ := tr.RootHash()

	trieStorage.EnterPruningBufferingMode()
	spm.PruneTrie(oldRootHash, temporary.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, temporary.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()

	spm.CancelPrune(oldRootHash, temporary.NewRoot, trieStorage)

	bufferedHashes := spm.pruningBuffer.RemoveAll()

	expectedHash := append(oldRootHash, byte(temporary.OldRoot))
	assert.Equal(t, append(expectedHash, byte(prune)), bufferedHashes[0])

	expectedHash = append(newRootHash, byte(temporary.NewRoot))
	assert.Equal(t, append(expectedHash, byte(cancelPrune)), bufferedHashes[1])

	expectedHash = append(oldRootHash, byte(temporary.NewRoot))
	assert.Equal(t, append(expectedHash, byte(cancelPrune)), bufferedHashes[2])
}

func TestAccountsDB_PruneAfterCancelPruneShouldFail(t *testing.T) {
	t.Parallel()

	tr, adb, spm := getDefaultTrieAndAccountsDbAndStoragePruningManager()
	trieStorage := tr.GetStorageManager()

	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
	_, _ = adb.Commit()
	rootHash, _ := tr.RootHash()

	_ = tr.Update([]byte("dog"), []byte("value of dog"))
	_, _ = adb.Commit()
	spm.CancelPrune(rootHash, temporary.NewRoot, trieStorage)

	spm.CancelPrune(rootHash, temporary.OldRoot, trieStorage)
	spm.PruneTrie(rootHash, temporary.OldRoot, trieStorage)

	newTr, err := tr.Recreate(rootHash)
	assert.Nil(t, err)
	assert.NotNil(t, newTr)
}

func TestStoragePruningManager_MarkForEviction_removeDuplicatedKeys(t *testing.T) {
	map1 := map[string]struct{}{
		"hash1": {},
		"hash2": {},
		"hash3": {},
		"hash4": {},
	}

	map2 := map[string]struct{}{
		"hash1": {},
		"hash4": {},
		"hash5": {},
		"hash6": {},
	}

	removeDuplicatedKeys(map1, map2)

	_, ok := map1["hash1"]
	assert.False(t, ok)
	_, ok = map1["hash4"]
	assert.False(t, ok)

	_, ok = map2["hash1"]
	assert.False(t, ok)
	_, ok = map2["hash4"]
	assert.False(t, ok)
}
