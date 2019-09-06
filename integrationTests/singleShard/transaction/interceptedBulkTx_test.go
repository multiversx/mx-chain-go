package transaction

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

var stepDelay = time.Second

func TestNode_GenerateSendInterceptBulkTransactionsWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startingNonce := uint64(6)
	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	nodeAddr := "0"

	n := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, nodeAddr)
	_ = n.Node.Start()

	defer func() {
		_ = n.Node.Stop()
	}()

	_ = n.Node.P2PBootstrap()

	time.Sleep(time.Second)

	//set the account's nonce to startingNonce
	_ = n.SetAccountNonce(startingNonce)
	noOfTx := 8000

	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(noOfTx)

	chanDone := make(chan bool)

	go func() {
		wg.Wait()
		chanDone <- true
	}()

	mut := sync.Mutex{}
	txHashes := make([][]byte, 0)
	transactions := make([]data.TransactionHandler, 0)

	//wire up handler
	n.ShardDataPool.Transactions().RegisterHandler(func(key []byte) {
		mut.Lock()
		defer mut.Unlock()

		txHashes = append(txHashes, key)

		dataStore := n.ShardDataPool.Transactions().ShardDataStore(
			process.ShardCacherIdentifier(n.ShardCoordinator.SelfId(), n.ShardCoordinator.SelfId()),
		)
		val, _ := dataStore.Get(key)

		if val == nil {
			assert.Fail(t, fmt.Sprintf("key %s not in store?", base64.StdEncoding.EncodeToString(key)))
			return
		}

		transactions = append(transactions, val.(*transaction.Transaction))
		wg.Done()
	})

	err := n.Node.GenerateAndSendBulkTransactions(
		integrationTests.CreateRandomHexString(64),
		big.NewInt(1),
		uint64(noOfTx),
	)

	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 60):
		assert.Fail(t, "timeout")
		return
	}

	integrationTests.CheckTxPresentAndRightNonce(
		t,
		startingNonce,
		noOfTx,
		txHashes,
		transactions,
		n.ShardDataPool.Transactions(),
		n.ShardCoordinator,
	)
}

func TestNode_GenerateSendInterceptBulkTransactions(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(1)
	currentShard := uint32(0)
	numOfTxs := uint64(1000)
	numOfNodes := 4
	idxProposer := 0
	mintingValue := big.NewInt(100000000)
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	}

	integrationTests.CreateMintingForSenders(
		nodes,
		currentShard,
		[]crypto.PrivateKey{nodes[idxProposer].OwnAccount.SkTxSign},
		mintingValue,
	)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	mutReceivedTxs := sync.Mutex{}
	receivedTxs := make(map[string]int, 0)

	nodes[idxProposer].ShardDataPool.Transactions().RegisterHandler(func(key []byte) {
		mutReceivedTxs.Lock()
		defer mutReceivedTxs.Unlock()

		receivedTxs[string(key)]++
	})

	_ = nodes[0].Node.GenerateAndSendBulkTransactions(integrationTests.CreateRandomHexString(64), big.NewInt(1), numOfTxs)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)

		fmt.Println(integrationTests.MakeDisplayTable(nodes))
	}

	mutReceivedTxs.Lock()
	for _, v := range receivedTxs {
		assert.Equal(t, 1, v)
	}

	assert.Equal(t, numOfTxs, uint64(len(receivedTxs)))

	mutReceivedTxs.Unlock()
}
