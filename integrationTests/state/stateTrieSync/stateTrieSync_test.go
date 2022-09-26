package stateTrieSync

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/throttler"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/syncer"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	testStorage "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/trie"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
	"github.com/ElrondNetwork/elrond-go/trie/keyBuilder"
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
	"github.com/ElrondNetwork/elrond-go/trie/storageMarker"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationtests/state/statetriesync")

const timeout = 30 * time.Second

func createTestProcessorNodeAndTrieStorage(
	t *testing.T,
	numOfShards uint32,
	shardID uint32,
	txSignPrivKeyShardId uint32,
) (*integrationTests.TestProcessorNode, storage.Storer) {
	mainStorer, _, err := testStorage.CreateTestingTriePruningStorer(&testscommon.ShardsCoordinatorMock{}, notifier.NewEpochStartSubscriptionHandler())
	assert.Nil(t, err)

	node := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            numOfShards,
		NodeShardId:          shardID,
		TxSignPrivKeyShardId: txSignPrivKeyShardId,
		TrieStore:            mainStorer,
		GasScheduleMap:       createTestGasMap(),
	})
	_ = node.Messenger.CreateTopic(common.ConsensusTopic+node.ShardCoordinator.CommunicationIdentifier(node.ShardCoordinator.SelfId()), true)

	return node, mainStorer
}

func TestNode_RequestInterceptTrieNodesWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester, trieStorageRequester := createTestProcessorNodeAndTrieStorage(t, numOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver, trieStorageResolver := createTestProcessorNodeAndTrieStorage(t, numOfShards, shardID, txSignPrivKeyShardId)

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
	// we have tested even with the 1000000 value and found out that it worked in a reasonable amount of time ~3.5 minutes
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

	numLeaves := getNumLeaves(t, resolverTrie, rootHash)
	assert.Equal(t, numTrieLeaves, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	nilRootHash, _ := requesterTrie.RootHash()

	tss := statistics.NewTrieSyncStatistics()
	arg := trie.ArgTrieSyncer{
		RequestHandler:            nRequester.RequestHandler,
		InterceptedNodes:          nRequester.DataPool.TrieNodes(),
		DB:                        requesterTrie.GetStorageManager(),
		Marshalizer:               integrationTests.TestMarshalizer,
		Hasher:                    integrationTests.TestHasher,
		ShardId:                   shardID,
		Topic:                     factory.AccountTrieNodesTopic,
		TrieSyncStatistics:        tss,
		TimeoutHandler:            testscommon.NewTimeoutHandlerMock(timeout),
		MaxHardCapForMissingNodes: 10000,
		CheckNodesOnDisk:          false,
	}
	trieSyncer, _ := trie.NewDoubleListTrieSyncer(arg)

	ctxPrint, cancel := context.WithCancel(context.Background())
	go printStatistics(ctxPrint, tss)

	err = trieSyncer.StartSyncing(rootHash, context.Background())
	assert.Nil(t, err)
	cancel()

	requesterTrie, err = requesterTrie.Recreate(rootHash)
	require.Nil(t, err)

	newRootHash, _ := requesterTrie.RootHash()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	numLeaves = getNumLeaves(t, requesterTrie, rootHash)
	assert.Equal(t, numTrieLeaves, numLeaves)
}

func printStatistics(ctx context.Context, stats common.SizeSyncStatisticsHandler) {
	lastDataReceived := uint64(0)
	printInterval := time.Second
	for {
		select {
		case <-ctx.Done():
			log.Info("finished trie sync",
				"num received", stats.NumReceived(),
				"num large nodes", stats.NumLarge(),
				"num missing", stats.NumMissing(),
				"data size received", core.ConvertBytes(stats.NumBytesReceived()),
			)
			return
		case <-time.After(printInterval):
			bytesReceivedDelta := stats.NumBytesReceived() - lastDataReceived
			if stats.NumBytesReceived() < lastDataReceived {
				bytesReceivedDelta = 0
			}
			lastDataReceived = stats.NumBytesReceived()

			bytesReceivedPerSec := float64(bytesReceivedDelta) / printInterval.Seconds()
			speed := fmt.Sprintf("%s/s", core.ConvertBytes(uint64(bytesReceivedPerSec)))

			log.Info("trie sync in progress",
				"num received", stats.NumReceived(),
				"num large nodes", stats.NumLarge(),
				"num missing", stats.NumMissing(),
				"data size received", core.ConvertBytes(stats.NumBytesReceived()),
				"speed", speed)
		}
	}
}

func TestNode_RequestInterceptTrieNodesWithMessengerNotSyncingShouldErr(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester, trieStorageRequester := createTestProcessorNodeAndTrieStorage(t, numOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver, trieStorageResolver := createTestProcessorNodeAndTrieStorage(t, numOfShards, shardID, txSignPrivKeyShardId)

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
	// we have tested even with the 1000000 value and found out that it worked in a reasonable amount of time ~3.5 minutes
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

	numLeaves := getNumLeaves(t, resolverTrie, rootHash)
	assert.Equal(t, numTrieLeaves, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))

	tss := statistics.NewTrieSyncStatistics()
	arg := trie.ArgTrieSyncer{
		RequestHandler:            nRequester.RequestHandler,
		InterceptedNodes:          nRequester.DataPool.TrieNodes(),
		DB:                        requesterTrie.GetStorageManager(),
		Marshalizer:               integrationTests.TestMarshalizer,
		Hasher:                    integrationTests.TestHasher,
		ShardId:                   shardID,
		Topic:                     factory.AccountTrieNodesTopic,
		TrieSyncStatistics:        tss,
		TimeoutHandler:            testscommon.NewTimeoutHandlerMock(timeout),
		MaxHardCapForMissingNodes: 10000,
		CheckNodesOnDisk:          false,
	}
	trieSyncer, _ := trie.NewDoubleListTrieSyncer(arg)

	ctxPrint, cancel := context.WithCancel(context.Background())
	go printStatistics(ctxPrint, tss)

	go func() {
		// sudden close of the resolver node after just 2 seconds
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

	var numOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester, trieStorageRequester := createTestProcessorNodeAndTrieStorage(t, numOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver, trieStorageResolver := createTestProcessorNodeAndTrieStorage(t, numOfShards, shardID, txSignPrivKeyShardId)

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
	dataTrieRootHashes := addAccountsToState(t, numAccounts, numDataTrieLeaves, accState, valSize)

	rootHash, _ := accState.RootHash()
	leavesChannel := make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
	err = accState.GetAllLeaves(leavesChannel, context.Background(), rootHash)
	for range leavesChannel {
	}
	require.Nil(t, err)

	requesterTrie := nRequester.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	nilRootHash, _ := requesterTrie.RootHash()

	syncerArgs := getUserAccountSyncerArgs(nRequester)

	userAccSyncer, err := syncer.NewUserAccountsSyncer(syncerArgs)
	assert.Nil(t, err)

	err = userAccSyncer.SyncAccounts(rootHash)
	assert.Nil(t, err)

	_ = nRequester.AccntState.RecreateTrie(rootHash)

	newRootHash, _ := nRequester.AccntState.RootHash()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	leavesChannel = make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
	err = nRequester.AccntState.GetAllLeaves(leavesChannel, context.Background(), rootHash)
	assert.Nil(t, err)
	numLeaves := 0
	for range leavesChannel {
		numLeaves++
	}
	assert.Equal(t, numAccounts, numLeaves)
	checkAllDataTriesAreSynced(t, numDataTrieLeaves, requesterTrie, dataTrieRootHashes)
}

func checkAllDataTriesAreSynced(t *testing.T, numDataTrieLeaves int, tr common.Trie, dataTriesRootHashes [][]byte) {
	for i := range dataTriesRootHashes {
		numLeaves := getNumLeaves(t, tr, dataTriesRootHashes[i])
		assert.Equal(t, numDataTrieLeaves, numLeaves)
	}
}

func addValuesToDataTrie(t *testing.T, adb state.AccountsAdapter, acc state.UserAccountHandler, numVals int, valSize int) []byte {
	for i := 0; i < numVals; i++ {
		keyRandBytes := integrationTests.CreateRandomBytes(32)
		valRandBytes := integrationTests.CreateRandomBytes(valSize)
		_ = acc.SaveKeyValue(keyRandBytes, valRandBytes)
	}

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	return acc.GetRootHash()
}

func TestSyncMissingSnapshotNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numSystemAccounts := 1
	numAccounts := 1000
	numDataTrieLeaves := 50
	valSize := 32
	roundsPerEpoch := uint64(5)
	numOfShards := 1
	nodesPerShard := 2
	numMetachainNodes := 1

	enableEpochsConfig := integrationTests.GetDefaultEnableEpochsConfig()
	enableEpochsConfig.StakingV2EnableEpoch = integrationTests.UnreachableEpoch
	nodes := integrationTests.CreateNodesWithEnableEpochsConfig(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochsConfig,
	)

	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	nRequester := nodes[0]
	nResolver := nodes[1]

	err := nRequester.ConnectTo(nResolver)
	require.Nil(t, err)
	time.Sleep(integrationTests.SyncDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	numDelayRounds := uint32(10)
	for i := uint64(0); i < uint64(numDelayRounds); i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		time.Sleep(integrationTests.StepDelay)
	}

	resolverTrie := nResolver.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	accState := nResolver.AccntState
	dataTrieRootHashes := addAccountsToState(t, numAccounts, numDataTrieLeaves, accState, valSize)
	rootHash, _ := accState.RootHash()
	numLeaves := getNumLeaves(t, resolverTrie, rootHash)
	require.Equal(t, numAccounts+numSystemAccounts, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	nilRootHash, _ := requesterTrie.RootHash()

	copyPartialState(t, nResolver, nRequester, dataTrieRootHashes)

	syncerArgs := getUserAccountSyncerArgs(nRequester)
	userAccSyncer, err := syncer.NewUserAccountsSyncer(syncerArgs)
	assert.Nil(t, err)

	err = nRequester.AccntState.SetSyncer(userAccSyncer)
	assert.Nil(t, err)
	err = nRequester.AccntState.StartSnapshotIfNeeded()
	assert.Nil(t, err)

	tsm := nRequester.TrieStorageManagers[trieFactory.UserAccountTrie]
	_ = tsm.PutInEpoch([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal), 0)
	nRequester.AccntState.SnapshotState(rootHash)
	for tsm.IsPruningBlocked() {
		time.Sleep(time.Millisecond * 100)
	}
	_ = nRequester.AccntState.RecreateTrie(rootHash)

	newRootHash, _ := nRequester.AccntState.RootHash()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	numLeaves = getNumLeaves(t, requesterTrie, rootHash)
	assert.Equal(t, numAccounts+numSystemAccounts, numLeaves)
	checkAllDataTriesAreSynced(t, numDataTrieLeaves, requesterTrie, dataTrieRootHashes)
}

func copyPartialState(t *testing.T, sourceNode, destinationNode *integrationTests.TestProcessorNode, dataTriesRootHashes [][]byte) {
	resolverTrie := sourceNode.TrieContainer.Get([]byte(trieFactory.UserAccountTrie))
	hashes, _ := resolverTrie.GetAllHashes()
	assert.NotEqual(t, 0, len(hashes))

	hashes = append(hashes, getDataTriesHashes(t, resolverTrie, dataTriesRootHashes)...)
	destStorage := destinationNode.TrieContainer.Get([]byte(trieFactory.UserAccountTrie)).GetStorageManager()

	for i, hash := range hashes {
		if i%1000 == 0 {
			continue
		}

		val, err := resolverTrie.GetStorageManager().Get(hash)
		assert.Nil(t, err)

		err = destStorage.Put(hash, val)
		assert.Nil(t, err)
	}

}

func getDataTriesHashes(t *testing.T, tr common.Trie, dataTriesRootHashes [][]byte) [][]byte {
	hashes := make([][]byte, 0)
	for _, rh := range dataTriesRootHashes {
		dt, err := tr.Recreate(rh)
		assert.Nil(t, err)

		dtHashes, err := dt.GetAllHashes()
		assert.Nil(t, err)

		hashes = append(hashes, dtHashes...)
	}

	return hashes
}

func addAccountsToState(t *testing.T, numAccounts int, numDataTrieLeaves int, accState state.AccountsAdapter, valSize int) [][]byte {
	dataTrieRootHashes := make([][]byte, numAccounts)

	for i := 0; i < numAccounts; i++ {
		address := integrationTests.CreateAccount(accState, 1, big.NewInt(100))
		account, _ := accState.LoadAccount(address)
		userAcc, ok := account.(state.UserAccountHandler)
		assert.True(t, ok)

		rootHash := addValuesToDataTrie(t, accState, userAcc, numDataTrieLeaves, valSize)
		dataTrieRootHashes[i] = rootHash
	}

	return dataTrieRootHashes
}

func getNumLeaves(t *testing.T, tr common.Trie, rootHash []byte) int {
	leavesChannel := make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity)
	err := tr.GetAllLeavesOnChannel(leavesChannel, context.Background(), rootHash, keyBuilder.NewDisabledKeyBuilder())
	require.Nil(t, err)

	numLeaves := 0
	for range leavesChannel {
		numLeaves++
	}

	return numLeaves
}

func getUserAccountSyncerArgs(node *integrationTests.TestProcessorNode) syncer.ArgsNewUserAccountsSyncer {
	thr, _ := throttler.NewNumGoRoutinesThrottler(50)
	syncerArgs := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                    integrationTests.TestHasher,
			Marshalizer:               integrationTests.TestMarshalizer,
			TrieStorageManager:        node.TrieStorageManagers[trieFactory.UserAccountTrie],
			RequestHandler:            node.RequestHandler,
			Timeout:                   common.TimeoutGettingTrieNodes,
			Cacher:                    node.DataPool.TrieNodes(),
			MaxTrieLevelInMemory:      200,
			MaxHardCapForMissingNodes: 5000,
			TrieSyncerVersion:         2,
			StorageMarker:             storageMarker.NewTrieStorageMarker(),
		},
		ShardId:                0,
		Throttler:              thr,
		AddressPubKeyConverter: integrationTests.TestAddressPubkeyConverter,
	}

	return syncerArgs
}
