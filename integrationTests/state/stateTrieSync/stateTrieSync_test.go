package stateTrieSync

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"testing"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/syncer"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationtests/state/statetriesync")

func createTestProcessorNodeAndTrieStorage(
	t *testing.T,
	numOfShards uint32,
	shardID uint32,
	txSignPrivKeyShardId uint32,
) (*integrationTests.TestProcessorNode, storage.Storer) {

	cacheConfig := storageUnit.CacheConfig{
		Name:        "trie",
		Type:        "SizeLRU",
		SizeInBytes: 314572800, //300MB
		Capacity:    500000,
	}
	trieCache, err := storageUnit.NewCache(cacheConfig)
	require.Nil(t, err)

	dbConfig := config.DBConfig{
		FilePath:          "trie",
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 2,
		MaxBatchSize:      45000,
		MaxOpenFiles:      10,
	}
	persisterFactory := storageFactory.NewPersisterFactory(dbConfig)
	tempDir, err := ioutil.TempDir("", "integrationTest")
	require.Nil(t, err)

	triePersister, err := persisterFactory.Create(tempDir)
	require.Nil(t, err)

	trieStorage, err := storageUnit.NewStorageUnit(trieCache, triePersister)
	require.Nil(t, err)

	node := integrationTests.NewTestProcessorNodeWithStorageTrieAndGasModel(numOfShards, shardID, txSignPrivKeyShardId, trieStorage, createTestGasMap())
	_ = node.Messenger.CreateTopic(common.ConsensusTopic+node.ShardCoordinator.CommunicationIdentifier(node.ShardCoordinator.SelfId()), true)

	return node, trieStorage
}

func TestNode_RequestInterceptTrieNodesWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester, trieStorageRequester := createTestProcessorNodeAndTrieStorage(t, nrOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver, trieStorageResolver := createTestProcessorNodeAndTrieStorage(t, nrOfShards, shardID, txSignPrivKeyShardId)

	defer func() {
		_ = trieStorageRequester.DestroyUnit()
		_ = trieStorageResolver.DestroyUnit()

		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.ConnectTo(nResolver)
	require.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	resolverTrie := nResolver.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	//we have tested even with the 1000000 value and found out that it worked in a reasonable amount of time ~3.5 minutes
	numTrieLeaves := 10000
	for i := 0; i < numTrieLeaves; i++ {
		_ = resolverTrie.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	nodes := resolverTrie.GetNumNodes()
	log.Info("trie nodes",
		"total", nodes.Branches+nodes.Extensions+nodes.Leaves,
		"branches", nodes.Branches,
		"extensions", nodes.Extensions,
		"leaves", nodes.Leaves,
		"max level", nodes.MaxLevel,
	)

	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.RootHash()

	leavesChannel, _ := resolverTrie.GetAllLeavesOnChannel(rootHash)
	numLeaves := 0
	for range leavesChannel {
		numLeaves++
	}
	assert.Equal(t, numTrieLeaves, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	nilRootHash, _ := requesterTrie.RootHash()

	timeout := 10 * time.Second
	tss := statistics.NewTrieSyncStatistics()
	arg := trie.ArgTrieSyncer{
		RequestHandler:            nRequester.RequestHandler,
		InterceptedNodes:          nRequester.DataPool.TrieNodes(),
		DB:                        requesterTrie.GetStorageManager().Database(),
		Marshalizer:               integrationTests.TestMarshalizer,
		Hasher:                    integrationTests.TestHasher,
		ShardId:                   shardID,
		Topic:                     factory.AccountTrieNodesTopic,
		TrieSyncStatistics:        tss,
		ReceivedNodesTimeout:      timeout,
		MaxHardCapForMissingNodes: 10000,
	}
	trieSyncer, _ := trie.NewDoubleListTrieSyncer(arg)

	ctxPrint, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctxPrint.Done():
				fmt.Printf("Sync done: received: %d, large: %d, missing: %d\n", tss.NumReceived(), tss.NumLarge(), tss.NumMissing())
				return
			case <-time.After(time.Millisecond * 100):
				fmt.Printf("Sync in progress: received: %d, large: %d, missing: %d\n", tss.NumReceived(), tss.NumLarge(), tss.NumMissing())
			}
		}
	}()

	err = trieSyncer.StartSyncing(rootHash, context.Background())
	assert.Nil(t, err)
	cancel()

	requesterTrie, err = requesterTrie.Recreate(rootHash)
	require.Nil(t, err)

	newRootHash, _ := requesterTrie.RootHash()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	leavesChannel, _ = requesterTrie.GetAllLeavesOnChannel(newRootHash)
	numLeaves = 0
	for range leavesChannel {
		numLeaves++
	}
	assert.Equal(t, numTrieLeaves, numLeaves)
}

func TestNode_RequestInterceptTrieNodesWithMessengerNotSyncingShouldErr(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester, trieStorageRequester := createTestProcessorNodeAndTrieStorage(t, nrOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver, trieStorageResolver := createTestProcessorNodeAndTrieStorage(t, nrOfShards, shardID, txSignPrivKeyShardId)

	defer func() {
		_ = trieStorageRequester.DestroyUnit()
		_ = trieStorageResolver.DestroyUnit()

		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.ConnectTo(nResolver)
	require.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	resolverTrie := nResolver.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	//we have tested even with the 1000000 value and found out that it worked in a reasonable amount of time ~3.5 minutes
	numTrieLeaves := 10000
	for i := 0; i < numTrieLeaves; i++ {
		_ = resolverTrie.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	nodes := resolverTrie.GetNumNodes()
	log.Info("trie nodes",
		"total", nodes.Branches+nodes.Extensions+nodes.Leaves,
		"branches", nodes.Branches,
		"extensions", nodes.Extensions,
		"leaves", nodes.Leaves,
		"max level", nodes.MaxLevel,
	)

	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.RootHash()

	leavesChannel, _ := resolverTrie.GetAllLeavesOnChannel(rootHash)
	numLeaves := 0
	for range leavesChannel {
		numLeaves++
	}
	assert.Equal(t, numTrieLeaves, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))

	timeout := 10 * time.Second
	tss := statistics.NewTrieSyncStatistics()
	arg := trie.ArgTrieSyncer{
		RequestHandler:            nRequester.RequestHandler,
		InterceptedNodes:          nRequester.DataPool.TrieNodes(),
		DB:                        requesterTrie.GetStorageManager().Database(),
		Marshalizer:               integrationTests.TestMarshalizer,
		Hasher:                    integrationTests.TestHasher,
		ShardId:                   shardID,
		Topic:                     factory.AccountTrieNodesTopic,
		TrieSyncStatistics:        tss,
		ReceivedNodesTimeout:      timeout,
		MaxHardCapForMissingNodes: 10000,
	}
	trieSyncer, _ := trie.NewDoubleListTrieSyncer(arg)

	ctxPrint, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctxPrint.Done():
				fmt.Printf("Sync done: received: %d, large: %d, missing: %d\n", tss.NumReceived(), tss.NumLarge(), tss.NumMissing())
				return
			case <-time.After(time.Millisecond * 100):
				fmt.Printf("Sync in progress: received: %d, large: %d, missing: %d\n", tss.NumReceived(), tss.NumLarge(), tss.NumMissing())
			}
		}
	}()

	go func() {
		//sudden close of the resolver node after just 2 seconds
		time.Sleep(time.Second * 2)
		_ = nResolver.Messenger.Close()
		log.Info("resolver node closed, the requester should soon fail in error")
	}()

	err = trieSyncer.StartSyncing(rootHash, context.Background())
	assert.Equal(t, trie.ErrTrieSyncTimeout, err)
	cancel()
}

func createTestGasMap() map[string]map[string]uint64 {
	gasSchedule := arwenConfig.MakeGasMapForTests()
	gasSchedule = defaults.FillGasMapInternal(gasSchedule, 1)

	return gasSchedule
}

func TestMultipleDataTriesSyncSmallValues(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testMultipleDataTriesSync(t, 1000, 50, 32)
}

func TestMultipleDataTriesSyncLargeValues(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testMultipleDataTriesSync(t, 3, 3, 1<<21)
}

func testMultipleDataTriesSync(t *testing.T, numAccounts int, numDataTrieLeaves int, valSize int) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester, trieStorageRequester := createTestProcessorNodeAndTrieStorage(t, nrOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver, trieStorageResolver := createTestProcessorNodeAndTrieStorage(t, nrOfShards, shardID, txSignPrivKeyShardId)

	defer func() {
		_ = trieStorageRequester.DestroyUnit()
		_ = trieStorageResolver.DestroyUnit()

		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.ConnectTo(nResolver)
	require.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	accState := nResolver.AccntState
	dataTrieRootHashes := make([][]byte, numAccounts)

	for i := 0; i < numAccounts; i++ {
		address := integrationTests.CreateAccount(accState, 1, big.NewInt(100))
		account, _ := accState.LoadAccount(address)
		userAcc, ok := account.(state.UserAccountHandler)
		assert.True(t, ok)

		rootHash := addValuesToDataTrie(t, accState, userAcc, numDataTrieLeaves, valSize)
		dataTrieRootHashes[i] = rootHash
	}

	rootHash, _ := accState.RootHash()
	leavesChannel, err := accState.GetAllLeaves(rootHash)
	for range leavesChannel {
	}
	require.Nil(t, err)

	requesterTrie := nRequester.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	nilRootHash, _ := requesterTrie.RootHash()

	thr, _ := throttler.NewNumGoRoutinesThrottler(50)
	syncerArgs := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                    integrationTests.TestHasher,
			Marshalizer:               integrationTests.TestMarshalizer,
			TrieStorageManager:        nRequester.TrieStorageManagers[trieFactory.UserAccountTrie],
			RequestHandler:            nRequester.RequestHandler,
			Timeout:                   common.TimeoutGettingTrieNodes,
			Cacher:                    nRequester.DataPool.TrieNodes(),
			MaxTrieLevelInMemory:      200,
			MaxHardCapForMissingNodes: 5000,
			TrieSyncerVersion:         2,
		},
		ShardId:   shardID,
		Throttler: thr,
	}

	userAccSyncer, err := syncer.NewUserAccountsSyncer(syncerArgs)
	assert.Nil(t, err)

	err = userAccSyncer.SyncAccounts(rootHash)
	assert.Nil(t, err)

	_ = nRequester.AccntState.RecreateTrie(rootHash)

	newRootHash, _ := nRequester.AccntState.RootHash()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	leavesChannel, err = nRequester.AccntState.GetAllLeaves(rootHash)
	assert.Nil(t, err)
	numLeaves := 0
	for range leavesChannel {
		numLeaves++
	}
	assert.Equal(t, numAccounts, numLeaves)
	checkAllDataTriesAreSynced(t, numDataTrieLeaves, nRequester.AccntState, dataTrieRootHashes)
}

func checkAllDataTriesAreSynced(t *testing.T, numDataTrieLeaves int, adb state.AccountsAdapter, dataTriesRootHashes [][]byte) {
	for i := range dataTriesRootHashes {
		dataTrieLeaves, err := adb.GetAllLeaves(dataTriesRootHashes[i])
		assert.Nil(t, err)
		numLeaves := 0
		for range dataTrieLeaves {
			numLeaves++
		}
		assert.Equal(t, numDataTrieLeaves, numLeaves)
	}
}

func addValuesToDataTrie(t *testing.T, adb state.AccountsAdapter, acc state.UserAccountHandler, numVals int, valSize int) []byte {
	for i := 0; i < numVals; i++ {
		keyRandBytes := integrationTests.CreateRandomBytes(32)
		valRandBytes := integrationTests.CreateRandomBytes(valSize)
		_ = acc.DataTrieTracker().SaveKeyValue(keyRandBytes, valRandBytes)
	}

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	return acc.GetRootHash()
}
