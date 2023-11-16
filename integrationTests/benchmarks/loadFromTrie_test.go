package benchmarks

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	testStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/hashesHolder/disabled"
	"github.com/stretchr/testify/require"
)

func TestTrieLoadTime(t *testing.T) {
	t.Skip()

	numTrieLevels := 50
	numTries := 100000
	numChildrenPerBranch := 8
	for i := 1; i <= numTrieLevels; i++ {
		testTrieLoadTime(t, numChildrenPerBranch, numTries, i)
	}
}

func TestTrieLoadTimeForOneLevel(t *testing.T) {
	numTrieLevels := 1
	numTries := 10000
	numChildrenPerBranch := 8
	for i := 1; i <= numTrieLevels; i++ {
		testTrieLoadTime(t, numChildrenPerBranch, numTries, i)
	}
}

func testTrieLoadTime(t *testing.T, numChildrenPerBranch int, numTries int, maxTrieLevel int) {
	store := getNewTrieStorage()
	defer func() {
		_ = store.DestroyUnit()
	}()
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := blake2b.NewBlake2b()

	tsm := getTrieStorageManager(store, marshaller, hasher)
	tries := generateTriesWithMaxDepth(t, numTries, maxTrieLevel, numChildrenPerBranch, tsm, marshaller, hasher)
	store.ClearCache()

	if maxTrieLevel == 1 {
		timeTrieRecreate(tries, maxTrieLevel)
		return
	}

	timeTrieLoad(t, tries, maxTrieLevel)
}

func timeTrieRecreate(tries []*keyForTrie, depth int) {
	startTime := time.Now()
	for j := range tries {
		_, _ = tries[j].tr.Recreate(tries[j].key)
	}
	duration := time.Since(startTime)
	fmt.Printf("trie with depth %d, duration %d \n", depth, duration.Nanoseconds()/int64(len(tries)))
}

func timeTrieLoad(t *testing.T, tries []*keyForTrie, depth int) {
	startTime := time.Now()
	for j := range tries {
		_, err := tries[j].tr.Get(tries[j].key)
		require.Nil(t, err)
		tries[j] = nil
	}
	duration := time.Since(startTime)
	fmt.Printf("trie with depth %d, duration %d \n", depth, duration.Nanoseconds()/int64(len(tries)))
}

type keyForTrie struct {
	key []byte
	tr  common.Trie
}

func generateTriesWithMaxDepth(
	t *testing.T,
	numTries int,
	numTrieLevels int,
	numChildrenPerBranch int,
	storage common.StorageManager,
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
) []*keyForTrie {
	tries := make([]*keyForTrie, numTries)
	for i := 0; i < numTries; i++ {
		tr, _ := trie.NewTrie(storage, marshaller, hasher, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, 2)
		key := insertKeysIntoTrie(t, tr, numTrieLevels, numChildrenPerBranch)

		rootHash, _ := tr.RootHash()
		collapsedTrie, _ := tr.Recreate(rootHash)

		if numTrieLevels == 1 {
			key = rootHash
		}

		tries[i] = &keyForTrie{
			key: key,
			tr:  collapsedTrie,
		}
	}

	return tries
}

func insertKeysIntoTrie(t *testing.T, tr common.Trie, numTrieLevels int, numChildrenPerBranch int) []byte {
	key := insertInTrie(tr, numTrieLevels, numChildrenPerBranch)

	tld, _ := tr.Get(key)
	require.Equal(t, uint32(numTrieLevels), tld.Depth()+1)
	_ = tr.Commit()
	return key
}

func insertInTrie(tr common.Trie, numTrieLevels int, numChildrenPerBranch int) []byte {
	keys := integrationTests.GenerateTrieKeysForMaxLevel(numTrieLevels, numChildrenPerBranch)
	for _, key := range keys {
		_ = tr.Update(key, key)
	}

	lastKeyIndex := len(keys) - 1
	return keys[lastKeyIndex]
}

func getTrieStorageManager(store storage.Storer, marshaller marshal.Marshalizer, hasher hashing.Hasher) common.StorageManager {
	args := testStorage.GetStorageManagerArgs()
	args.MainStorer = store
	args.Marshalizer = marshaller
	args.Hasher = hasher
	args.CheckpointHashesHolder = disabled.NewDisabledCheckpointHashesHolder()

	trieStorageManager, _ := trie.NewTrieStorageManager(args)

	return trieStorageManager
}

func getNewTrieStorage() storage.Storer {
	batchDelaySeconds := 1
	maxBatchSize := 40000
	maxNumOpenedFiles := 10

	db, _ := database.NewSerialDB("AccountsTrie", batchDelaySeconds, maxBatchSize, maxNumOpenedFiles)
	cacher, _ := storageunit.NewCache(storageunit.CacheConfig{
		Type:        storageunit.SizeLRUCache,
		Capacity:    1,
		Shards:      1,
		SizeInBytes: 1024,
	})

	store, _ := storageunit.NewStorageUnit(
		cacher,
		db,
	)

	return store
}
