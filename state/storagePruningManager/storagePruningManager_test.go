package storagePruningManager

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/iteratorChannelsProvider"
	"github.com/multiversx/mx-chain-go/state/lastSnapshotMarker"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
)

func getDefaultTrieAndAccountsDbAndStoragePruningManager() (common.Trie, *state.AccountsDB, *storagePruningManager) {
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		SnapshotsGoroutineNum: 1,
	}
	marshaller := &marshallerMock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	args := storage.GetStorageManagerArgs()
	trieStorage, _ := trie.NewTrieStorageManager(args)
	tr, _ := trie.NewTrie(trieStorage, marshaller, hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, 5)
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := NewStoragePruningManager(ewl, generalCfg.PruningBufferLen)

	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:                 hasher,
		Marshaller:             marshaller,
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		StateAccessesCollector: &stateMock.StateAccessesCollectorStub{},
	}
	accCreator, _ := factory.NewAccountCreator(argsAccCreator)

	snapshotsManager, _ := state.NewSnapshotsManager(state.ArgsNewSnapshotsManager{
		ProcessingMode:       common.Normal,
		Marshaller:           marshaller,
		AddressConverter:     &testscommon.PubkeyConverterMock{},
		ProcessStatusHandler: &testscommon.ProcessStatusHandlerStub{},
		StateMetrics:         &stateMock.StateMetricsStub{},
		AccountFactory:       accCreator,
		ChannelsProvider:     iteratorChannelsProvider.NewUserStateIteratorChannelsProvider(),
		LastSnapshotMarker:   lastSnapshotMarker.NewLastSnapshotMarker(),
		StateStatsHandler:    statistics.NewStateStatistics(),
	})

	argsAccountsDB := state.ArgsAccountsDB{
		Trie:                   tr,
		Hasher:                 hasher,
		Marshaller:             marshaller,
		AccountFactory:         accCreator,
		StoragePruningManager:  spm,
		AddressConverter:       &testscommon.PubkeyConverterMock{},
		SnapshotsManager:       snapshotsManager,
		StateAccessesCollector: &stateMock.StateAccessesCollectorStub{},
	}
	adb, _ := state.NewAccountsDB(argsAccountsDB)

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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))
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
	spm.PruneTrie(newRootHash, state.NewRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))
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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))
	spm.CancelPrune(newRootHash, state.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()
	assert.Equal(t, 2, spm.pruningBuffer.Len())

	adb.PruneTrie(oldRootHash, state.NewRoot, state.NewPruningHandler(state.EnableDataRemoval))
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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))
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
	spm.PruneTrie(oldRootHash, state.OldRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))
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
	spm.PruneTrie(rootHash, state.OldRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))

	newTr, err := tr.Recreate(holders.NewDefaultRootHashesHolder(rootHash))
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

func TestStoragePruningManager_Reset(t *testing.T) {
	t.Parallel()

	args := storage.GetStorageManagerArgs()
	trieStorage, _ := trie.NewTrieStorageManager(args)
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := NewStoragePruningManager(ewl, 1000)

	err := spm.MarkForEviction([]byte("rootHash"), []byte("newRootHash"), map[string]struct{}{"hash1": {}, "hash2": {}}, map[string]struct{}{"hash3": {}, "hash4": {}})
	assert.Nil(t, err)
	err = spm.markForEviction([]byte("rootHash2"), map[string]struct{}{"hash5": {}, "hash6": {}}, state.NewRoot)
	assert.Nil(t, err)

	trieStorage.EnterPruningBufferingMode()
	spm.PruneTrie([]byte("rootHash"), state.OldRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))
	spm.CancelPrune([]byte("newRootHash"), state.NewRoot, trieStorage)
	trieStorage.ExitPruningBufferingMode()

	assert.Equal(t, 2, spm.pruningBuffer.Len())

	spm.Reset()
	assert.Equal(t, 0, spm.pruningBuffer.Len())

	// rootHash2 should not be added to the pruning buffer because ewl was also reset when spm.Reset() was called
	trieStorage.EnterPruningBufferingMode()
	spm.PruneTrie([]byte("rootHash2"), state.NewRoot, trieStorage, state.NewPruningHandler(state.EnableDataRemoval))
	trieStorage.ExitPruningBufferingMode()
	assert.Equal(t, 0, spm.pruningBuffer.Len())
}

func TestStoragePruningManager_EvictBeforePut(t *testing.T) {
	t.Parallel()

	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := NewStoragePruningManager(ewl, 1000)

	// Simulate dismissed block: MarkForEviction from oldRoot=R0 to newRoot=R1
	dismissedOldHashes := map[string]struct{}{"old_dismissed_1": {}, "old_dismissed_2": {}}
	dismissedNewHashes := map[string]struct{}{"new_dismissed_1": {}}
	err := spm.MarkForEviction([]byte("R0"), []byte("R1"), dismissedOldHashes, dismissedNewHashes)
	assert.Nil(t, err)

	// Verify EWL has 2 entries (R0|OldRoot and R1|NewRoot)
	assert.Equal(t, 2, spm.EvictionWaitingListCacheLen())

	// Simulate replacement block with SAME oldRoot: MarkForEviction from oldRoot=R0 to newRoot=R2
	// The evict-before-put should clear the stale R0|OldRoot entry before writing the new one
	replacementOldHashes := map[string]struct{}{"old_replacement_1": {}, "old_replacement_3": {}}
	replacementNewHashes := map[string]struct{}{"new_replacement_1": {}}
	err = spm.MarkForEviction([]byte("R0"), []byte("R2"), replacementOldHashes, replacementNewHashes)
	assert.Nil(t, err)

	// EWL should have 3 entries: R1|NewRoot (from dismissed), R0|OldRoot (replacement), R2|NewRoot (replacement)
	// The stale R0|OldRoot entry from the dismissed block was evicted before the replacement's Put
	assert.Equal(t, 3, spm.EvictionWaitingListCacheLen())

	// Now simulate pruning the replacement block's old state:
	// CancelPrune(R0, NewRoot) - cancel the "new" marking from the previous block
	// The R0|NewRoot was set by the DISMISSED block's MarkForEviction (removeDuplicatedKeys already ran)
	// PruneTrie(R0, OldRoot) - prune old state
	// The key R0|OldRoot should return the REPLACEMENT's hashes, not the dismissed block's
	evictedOld, errEvict := ewl.Evict(append([]byte("R0"), byte(state.OldRoot)))
	assert.Nil(t, errEvict)

	// The evicted hashes must be from the replacement block, not the dismissed block
	_, hasReplacementHash := evictedOld["old_replacement_1"]
	assert.True(t, hasReplacementHash, "should contain replacement hashes")
	_, hasDismissedHash := evictedOld["old_dismissed_1"]
	assert.False(t, hasDismissedHash, "should NOT contain dismissed hashes")

	// The dismissed block's NewRoot entry (R1|NewRoot) should still be in EWL
	evictedDismissedNew, errEvict2 := ewl.Evict(append([]byte("R1"), byte(state.NewRoot)))
	assert.Nil(t, errEvict2)
	_, hasDismissedNewHash := evictedDismissedNew["new_dismissed_1"]
	assert.True(t, hasDismissedNewHash, "dismissed NewRoot entry should still exist")

	// After evicting both R0|OldRoot and R1|NewRoot, only R2|NewRoot should remain
	assert.Equal(t, 1, spm.EvictionWaitingListCacheLen())
}

func TestStoragePruningManager_EvictionWaitingListCacheLen(t *testing.T) {
	t.Parallel()

	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := NewStoragePruningManager(ewl, 1000)

	assert.Equal(t, 0, spm.EvictionWaitingListCacheLen())

	err := spm.MarkForEviction([]byte("old"), []byte("new"),
		map[string]struct{}{"h1": {}},
		map[string]struct{}{"h2": {}})
	assert.Nil(t, err)
	assert.Equal(t, 2, spm.EvictionWaitingListCacheLen())
}
