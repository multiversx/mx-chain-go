package stateTrieSync

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	testStorage "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/trie/storageMarker"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	logger "github.com/multiversx/mx-chain-logger-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
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
	t.Run("test with double lists version", func(t *testing.T) {
		testNodeRequestInterceptTrieNodesWithMessenger(t, 2)
	})
	t.Run("test with depth version", func(t *testing.T) {
		testNodeRequestInterceptTrieNodesWithMessenger(t, 3)
	})
}

func testNodeRequestInterceptTrieNodesWithMessenger(t *testing.T, version int) {
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

	resolverTrie := nResolver.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	// we have tested even with the 1000000 value and found out that it worked in a reasonable amount of time ~3.5 minutes
	numTrieLeaves := 10000
	for i := 0; i < numTrieLeaves; i++ {
		_ = resolverTrie.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.RootHash()

	numLeaves := getNumLeaves(t, resolverTrie, rootHash)
	assert.Equal(t, numTrieLeaves, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
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
	trieSyncer, _ := trie.CreateTrieSyncer(arg, version)

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
				"num processed", stats.NumProcessed(),
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
				"num processed", stats.NumProcessed(),
				"num large nodes", stats.NumLarge(),
				"num missing", stats.NumMissing(),
				"data size received", core.ConvertBytes(stats.NumBytesReceived()),
				"speed", speed)
		}
	}
}

func TestNode_RequestInterceptTrieNodesWithMessengerNotSyncingShouldErr(t *testing.T) {
	t.Run("test with double lists version", func(t *testing.T) {
		testNodeRequestInterceptTrieNodesWithMessengerNotSyncingShouldErr(t, 2)
	})
	t.Run("test with depth version", func(t *testing.T) {
		testNodeRequestInterceptTrieNodesWithMessengerNotSyncingShouldErr(t, 3)
	})
}

func testNodeRequestInterceptTrieNodesWithMessengerNotSyncingShouldErr(t *testing.T, version int) {
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

	resolverTrie := nResolver.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	// we have tested even with the 1000000 value and found out that it worked in a reasonable amount of time ~3.5 minutes
	numTrieLeaves := 100000
	for i := 0; i < numTrieLeaves; i++ {
		_ = resolverTrie.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.RootHash()

	numLeaves := getNumLeaves(t, resolverTrie, rootHash)
	assert.Equal(t, numTrieLeaves, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))

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
	trieSyncer, _ := trie.CreateTrieSyncer(arg, version)

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
	gasSchedule := wasmConfig.MakeGasMapForTests()
	gasSchedule = defaults.FillGasMapInternal(gasSchedule, 1)

	return gasSchedule
}

func TestMultipleDataTriesSyncSmallValues(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("test with double lists version", func(t *testing.T) {
		testMultipleDataTriesSync(t, 1000, 50, 32, 2)
	})
	t.Run("test with depth version", func(t *testing.T) {
		testMultipleDataTriesSync(t, 1000, 50, 32, 3)
	})
}

func TestMultipleDataTriesSyncLargeValues(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("test with double lists version", func(t *testing.T) {
		testMultipleDataTriesSync(t, 3, 3, 1<<21, 2)
	})
	t.Run("test with depth version", func(t *testing.T) {
		testMultipleDataTriesSync(t, 3, 3, 1<<21, 3)
	})
}

func testMultipleDataTriesSync(t *testing.T, numAccounts int, numDataTrieLeaves int, valSize int, version int) {
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
	leavesChannel := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = accState.GetAllLeaves(leavesChannel, context.Background(), rootHash, parsers.NewMainTrieLeafParser())
	for range leavesChannel.LeavesChan {
	}
	require.Nil(t, err)
	err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
	require.Nil(t, err)

	requesterTrie := nRequester.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	nilRootHash, _ := requesterTrie.RootHash()

	syncerArgs := getUserAccountSyncerArgs(nRequester, version)

	userAccSyncer, err := syncer.NewUserAccountsSyncer(syncerArgs)
	assert.Nil(t, err)

	err = userAccSyncer.SyncAccounts(rootHash, storageMarker.NewDisabledStorageMarker())
	assert.Nil(t, err)

	_ = nRequester.AccntState.RecreateTrie(rootHash)

	newRootHash, _ := nRequester.AccntState.RootHash()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	leavesChannel = &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = nRequester.AccntState.GetAllLeaves(leavesChannel, context.Background(), rootHash, parsers.NewMainTrieLeafParser())
	assert.Nil(t, err)
	numLeaves := 0
	for range leavesChannel.LeavesChan {
		numLeaves++
	}
	err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
	require.Nil(t, err)
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

type snapshotWatcher interface {
	IsSnapshotInProgress() bool
}

func TestSyncMissingSnapshotNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("test with double lists version", func(t *testing.T) {
		testSyncMissingSnapshotNodes(t, 2)
	})
	t.Run("test with depth version", func(t *testing.T) {
		testSyncMissingSnapshotNodes(t, 3)
	})
}

func testSyncMissingSnapshotNodes(t *testing.T, version int) {
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

	resolverTrie := nResolver.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	accState := nResolver.AccntState
	dataTrieRootHashes := addAccountsToState(t, numAccounts, numDataTrieLeaves, accState, valSize)
	rootHash, _ := accState.RootHash()
	numLeaves := getNumLeaves(t, resolverTrie, rootHash)
	require.Equal(t, numAccounts+numSystemAccounts, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	nilRootHash, _ := requesterTrie.RootHash()

	copyPartialState(t, nResolver, nRequester, dataTrieRootHashes)

	syncerArgs := getUserAccountSyncerArgs(nRequester, version)
	userAccSyncer, err := syncer.NewUserAccountsSyncer(syncerArgs)
	assert.Nil(t, err)

	err = nRequester.AccntState.SetSyncer(userAccSyncer)
	assert.Nil(t, err)
	err = nRequester.AccntState.StartSnapshotIfNeeded()
	assert.Nil(t, err)

	tsm := nRequester.TrieStorageManagers[dataRetriever.UserAccountsUnit.String()]
	_ = tsm.PutInEpoch([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal), 0)
	nRequester.AccntState.SnapshotState(rootHash, nRequester.EpochNotifier.CurrentEpoch())
	sw, ok := nRequester.AccntState.(snapshotWatcher)
	assert.True(t, ok)

	for sw.IsSnapshotInProgress() {
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
	resolverTrie := sourceNode.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String()))
	hashes, _ := resolverTrie.GetAllHashes()
	assert.NotEqual(t, 0, len(hashes))

	hashes = append(hashes, getDataTriesHashes(t, resolverTrie, dataTriesRootHashes)...)
	destStorage := destinationNode.TrieContainer.Get([]byte(dataRetriever.UserAccountsUnit.String())).GetStorageManager()

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
	leavesChannel := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err := tr.GetAllLeavesOnChannel(
		leavesChannel,
		context.Background(),
		rootHash,
		keyBuilder.NewDisabledKeyBuilder(),
		parsers.NewMainTrieLeafParser(),
	)
	require.Nil(t, err)

	numLeaves := 0
	for range leavesChannel.LeavesChan {
		numLeaves++
	}

	err = leavesChannel.ErrChan.ReadFromChanNonBlocking()
	require.Nil(t, err)

	return numLeaves
}

func getUserAccountSyncerArgs(node *integrationTests.TestProcessorNode, version int) syncer.ArgsNewUserAccountsSyncer {
	thr, _ := throttler.NewNumGoRoutinesThrottler(50)
	syncerArgs := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                            integrationTests.TestHasher,
			Marshalizer:                       integrationTests.TestMarshalizer,
			TrieStorageManager:                node.TrieStorageManagers[dataRetriever.UserAccountsUnit.String()],
			RequestHandler:                    node.RequestHandler,
			Timeout:                           common.TimeoutGettingTrieNodes,
			Cacher:                            node.DataPool.TrieNodes(),
			MaxTrieLevelInMemory:              200,
			MaxHardCapForMissingNodes:         5000,
			TrieSyncerVersion:                 version,
			UserAccountsSyncStatisticsHandler: statistics.NewTrieSyncStatistics(),
			AppStatusHandler:                  integrationTests.TestAppStatusHandler,
			EnableEpochsHandler:               node.EnableEpochsHandler,
		},
		ShardId:                0,
		Throttler:              thr,
		AddressPubKeyConverter: integrationTests.TestAddressPubkeyConverter,
	}

	return syncerArgs
}
