package transaction

import (
	"context"
	"encoding/json"
	"math/big"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var receivedTxsPerShard map[uint32]int

// This will test if txs are sent correctly intr-shard (shard 1 -> shard 1) and cross-shard (shard 1 -> shard 0)
func TestNode_SendBulkTransactionsAllTransactionsShouldBeSentCorrectly(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 0
	firstSkInShard := uint32(0)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes[0] = integrationTests.NewTestProcessorNode(uint32(numOfShards), 0, firstSkInShard, integrationTests.GetConnectableAddress(advertiser))
	integrationTests.CreateAccountForNodes(nodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	numOfCrossShardTxsToSend := 300
	numOfIntraShardTxsToSend := 400
	totalNumOfTxs := numOfCrossShardTxsToSend + numOfIntraShardTxsToSend
	numOfAccounts := 40
	receivedTxsPerShard = make(map[uint32]int, 0)
	mutGeneratedTxHashes := sync.Mutex{}
	//wire a new hook for generated txs on a node in sender shard to populate tx hashes generated
	for _, n := range nodes {
		nodeShardId := n.ShardCoordinator.SelfId()
		if nodeShardId != sharding.MetachainShardId {
			txPoolRegister(n, &mutGeneratedTxHashes)
		}
	}

	// generate a given number of accounts
	accounts := generateAccounts(numOfAccounts)
	accountsByShard := make(map[uint32][]keyPair, 0)
	for _, acc := range accounts {
		pkBytes, _ := acc.pk.ToByteArray()
		shardId := nodes[0].ShardCoordinator.ComputeId(state.NewAddress(pkBytes))
		accountsByShard[shardId] = append(accountsByShard[shardId], acc)
	}

	var txs []*transaction.Transaction

	// generate cross shard txs from shard 1 to shard 0
	for len(txs) < numOfCrossShardTxsToSend {
		// generate a tx from shard 0 to shard 1 and check if signature is correct
		randomSenderId := getRandomIndex(accountsByShard[1])
		sender := accountsByShard[1][randomSenderId]

		randomReceiverId := getRandomIndex(accountsByShard[0])
		receiver := accountsByShard[0][randomReceiverId]

		tx := generateTx(sender.sk, receiver.pk)

		txs = append(txs, tx)
	}

	// generate intra shard txs in shard 1
	for len(txs) < totalNumOfTxs {
		// generate a tx from a random account from shard 1 to other random account from shard 1
		randomSenderId := getRandomIndex(accountsByShard[1])
		sender := accountsByShard[1][randomSenderId]

		randomReceiverId := getRandomIndex(accountsByShard[1])
		receiver := accountsByShard[1][randomReceiverId]

		tx := generateTx(sender.sk, receiver.pk)

		txs = append(txs, tx)
	}

	sentTxs, err := nodes[0].Node.SendBulkTransactions(txs)
	assert.Nil(t, err)

	// wait for txs to be sent
	time.Sleep(1 * time.Second)
	// check if all txs were sent
	if len(accountsByShard[0]) > 1 {
		assert.Equal(t, uint64(totalNumOfTxs), sentTxs)
	} else {
		assert.Equal(t, uint64(numOfCrossShardTxsToSend), sentTxs)
	}

	mutGeneratedTxHashes.Lock()
	// check if all nodes in shard 0 got cross shard txs in pool
	assert.Equal(t, numOfCrossShardTxsToSend, receivedTxsPerShard[0]/nodesPerShard)

	// check if all nodes in shard 1 got all txs (cross shard and intra shard)
	assert.Equal(t, totalNumOfTxs, receivedTxsPerShard[1]/nodesPerShard)
	mutGeneratedTxHashes.Unlock()
}

func txPoolRegister(n *integrationTests.TestProcessorNode, mutex *sync.Mutex) {

	n.ShardDataPool.Transactions().RegisterHandler(func(key []byte) {
		mutex.Lock()
		receivedTxsPerShard[n.ShardCoordinator.SelfId()]++
		mutex.Unlock()
	})
}

type keyPair struct {
	pk crypto.PublicKey
	sk crypto.PrivateKey
}

func getRandomIndex(accSlice []keyPair) int {
	if len(accSlice) < 2 {
		return 0
	}

	return rand.Intn(len(accSlice) - 1)
}

func generateAccounts(numOfAccounts int) []keyPair {
	var accounts []keyPair
	for idx := 0; idx < numOfAccounts; idx++ {
		kg := signing.NewKeyGenerator(kyber.NewBlakeSHA256Ed25519())
		sk, pk := kg.GeneratePair()
		accounts = append(accounts, keyPair{
			pk: pk,
			sk: sk,
		})
	}

	return accounts
}

func generateTx(sender crypto.PrivateKey, receiver crypto.PublicKey) *transaction.Transaction {
	receiverBytes, _ := receiver.ToByteArray()
	senderBytes, _ := sender.GeneratePublic().ToByteArray()
	tx := transaction.Transaction{
		Nonce:     1,
		Value:     new(big.Int).SetInt64(10),
		RcvAddr:   receiverBytes,
		SndAddr:   senderBytes,
		GasPrice:  5,
		GasLimit:  1000,
		Data:      "",
		Signature: nil,
		Challenge: nil,
	}
	marshalizedTxBeforeSigning, _ := json.Marshal(tx)
	signer := singlesig.SchnorrSigner{}

	signature, _ := signer.Sign(sender, marshalizedTxBeforeSigning)
	tx.Signature = signature
	return &tx
}
