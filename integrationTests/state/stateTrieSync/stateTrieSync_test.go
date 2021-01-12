package stateTrieSync

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/syncer"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	factory2 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie/statistics"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_RequestInterceptTrieNodesWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	requesterNodeAddr := "0"
	resolverNodeAddr := "1"

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, requesterNodeAddr)
	_ = nRequester.Messenger.CreateTopic(core.ConsensusTopic+nRequester.ShardCoordinator.CommunicationIdentifier(nRequester.ShardCoordinator.SelfId()), true)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nResolver.Messenger.CreateTopic(core.ConsensusTopic+nResolver.ShardCoordinator.CommunicationIdentifier(nResolver.ShardCoordinator.SelfId()), true)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	resolverTrie := nResolver.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	//we have tested even with the 50000 value and found out that it worked in a reasonable amount of time ~21 seconds
	numTrieLeaves := 10000
	for i := 0; i < numTrieLeaves; i++ {
		_ = resolverTrie.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.Root()

	leavesChannel, _ := resolverTrie.GetAllLeavesOnChannel(rootHash, context.Background())
	numLeaves := 0
	for range leavesChannel {
		numLeaves++
	}
	assert.Equal(t, numTrieLeaves, numLeaves)

	requesterTrie := nRequester.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	nilRootHash, _ := requesterTrie.Root()
	whiteListHandler, _ := interceptors.NewWhiteListDataVerifier(
		&testscommon.CacherStub{
			PutCalled: func(_ []byte, _ interface{}, _ int) (evicted bool) {
				return false
			},
		},
	)
	requestHandler, _ := requestHandlers.NewResolverRequestHandler(
		nRequester.ResolverFinder,
		&mock.RequestedItemsHandlerStub{},
		whiteListHandler,
		10000,
		nRequester.ShardCoordinator.SelfId(),
		time.Second,
	)

	timeout := 10 * time.Second
	tss := statistics.NewTrieSyncStatistics()
	arg := trie.ArgTrieSyncer{
		RequestHandler:                 requestHandler,
		InterceptedNodes:               nRequester.DataPool.TrieNodes(),
		Trie:                           requesterTrie,
		ShardId:                        shardID,
		Topic:                          factory.AccountTrieNodesTopic,
		TrieSyncStatistics:             tss,
		TimeoutBetweenTrieNodesCommits: timeout,
	}
	trieSyncer, _ := trie.NewTrieSyncer(arg)

	ctxPrint, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctxPrint.Done():
				fmt.Printf("Sync done: received: %d, missing: %d\n", tss.NumReceived(), tss.NumMissing())
				return
			case <-time.After(time.Millisecond * 100):
				fmt.Printf("Sync in progress: received: %d, missing: %d\n", tss.NumReceived(), tss.NumMissing())
			}
		}
	}()

	err = trieSyncer.StartSyncing(rootHash, context.Background())
	assert.Nil(t, err)
	cancel()

	newRootHash, _ := requesterTrie.Root()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	leavesChannel, _ = requesterTrie.GetAllLeavesOnChannel(newRootHash, context.Background())
	numLeaves = 0
	for range leavesChannel {
		numLeaves++
	}
	assert.Equal(t, numTrieLeaves, numLeaves)
}

func TestMultipleDataTriesSync(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	requesterNodeAddr := "0"
	resolverNodeAddr := "1"

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, requesterNodeAddr)
	_ = nRequester.Messenger.CreateTopic(core.ConsensusTopic+nRequester.ShardCoordinator.CommunicationIdentifier(nRequester.ShardCoordinator.SelfId()), true)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nResolver.Messenger.CreateTopic(core.ConsensusTopic+nResolver.ShardCoordinator.CommunicationIdentifier(nResolver.ShardCoordinator.SelfId()), true)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	numAccounts := 1000
	numDataTrieLeaves := 50
	accState := nResolver.AccntState
	dataTrieRootHashes := make([][]byte, numAccounts)

	for i := 0; i < numAccounts; i++ {
		address := integrationTests.CreateAccount(accState, 1, big.NewInt(100))
		account, _ := accState.LoadAccount(address)
		userAcc, ok := account.(state.UserAccountHandler)
		assert.True(t, ok)

		rootHash := addValuesToDataTrie(t, accState, userAcc, numDataTrieLeaves)
		dataTrieRootHashes[i] = rootHash
	}

	rootHash, _ := accState.RootHash()
	leavesChannel, err := accState.GetAllLeaves(rootHash, context.Background())
	for range leavesChannel {
	}
	require.Nil(t, err)

	requesterTrie := nRequester.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	nilRootHash, _ := requesterTrie.Root()
	whiteListHandler, _ := interceptors.NewWhiteListDataVerifier(
		&testscommon.CacherStub{
			PutCalled: func(_ []byte, _ interface{}, _ int) (evicted bool) {
				return false
			},
		},
	)
	requestHandler, _ := requestHandlers.NewResolverRequestHandler(
		nRequester.ResolverFinder,
		&mock.RequestedItemsHandlerStub{},
		whiteListHandler,
		10000,
		nRequester.ShardCoordinator.SelfId(),
		time.Second,
	)

	thr, _ := throttler.NewNumGoRoutinesThrottler(50)
	syncerArgs := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:               integrationTests.TestHasher,
			Marshalizer:          integrationTests.TestMarshalizer,
			TrieStorageManager:   nRequester.TrieStorageManagers[factory2.UserAccountTrie],
			RequestHandler:       requestHandler,
			Timeout:              time.Second * 10,
			Cacher:               nRequester.DataPool.TrieNodes(),
			MaxTrieLevelInMemory: 5,
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

	leavesChannel, err = nRequester.AccntState.GetAllLeaves(rootHash, context.Background())
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
		dataTrieLeaves, err := adb.GetAllLeaves(dataTriesRootHashes[i], context.Background())
		assert.Nil(t, err)
		numLeaves := 0
		for range dataTrieLeaves {
			numLeaves++
		}
		assert.Equal(t, numDataTrieLeaves, numLeaves)
	}
}

func addValuesToDataTrie(t *testing.T, adb state.AccountsAdapter, acc state.UserAccountHandler, numVals int) []byte {
	for i := 0; i < numVals; i++ {
		randBytes := integrationTests.CreateRandomBytes(32)
		_ = acc.DataTrieTracker().SaveKeyValue(randBytes, randBytes)
	}

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	return acc.GetRootHash()
}
