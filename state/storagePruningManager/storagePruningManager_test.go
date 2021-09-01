package storagePruningManager

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/factory"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/stretchr/testify/assert"
)

func getDefaultTrieAndAccountsDbAndStoragePruningManager() (common.Trie, *state.AccountsDB, *storagePruningManager) {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	marshalizer := &testscommon.MarshalizerMock{}
	hsh := testscommon.HasherMock{}
	args := trie.NewTrieStorageManagerArgs{
		DB:                testscommon.NewMemDbMock(),
		MainStorer:        testscommon.CreateMemUnit(),
		CheckpointsStorer: testscommon.CreateMemUnit(),
		Marshalizer:       marshalizer,
		Hasher:            hsh,
		SnapshotDbConfig: config.DBConfig{
			Type: "MemoryDB",
		},
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, testscommon.HashSize),
	}
	trieStorage, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(trieStorage, marshalizer, hsh, 5)
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, testscommon.NewMemDbMock(), marshalizer)
	spm, _ := NewStoragePruningManager(ewl, generalCfg.PruningBufferLen)
	adb, _ := state.NewAccountsDB(tr, hsh, marshalizer, factory.NewAccountCreator(), spm)

	return tr, adb, spm
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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, state.NewRoot, trieStorage)
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
	spm.CancelPrune(oldRootHash, state.OldRoot, trieStorage)
	spm.PruneTrie(newRootHash, state.NewRoot, trieStorage)
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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, state.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()
	assert.Equal(t, 2, spm.pruningBuffer.Len())

	adb.PruneTrie(oldRootHash, state.NewRoot)
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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, state.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()
	assert.Equal(t, 2, spm.pruningBuffer.Len())

	adb.CancelPrune(oldRootHash, state.NewRoot)
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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage)
	spm.CancelPrune(newRootHash, state.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()

	spm.CancelPrune(oldRootHash, state.NewRoot, trieStorage)

	bufferedHashes := spm.pruningBuffer.RemoveAll()

	expectedHash := append(oldRootHash, byte(state.OldRoot))
	assert.Equal(t, append(expectedHash, byte(prune)), bufferedHashes[0])

	expectedHash = append(newRootHash, byte(state.NewRoot))
	assert.Equal(t, append(expectedHash, byte(cancelPrune)), bufferedHashes[1])

	expectedHash = append(oldRootHash, byte(state.NewRoot))
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
	spm.CancelPrune(rootHash, state.NewRoot, trieStorage)

	spm.CancelPrune(rootHash, state.OldRoot, trieStorage)
	spm.PruneTrie(rootHash, state.OldRoot, trieStorage)

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
