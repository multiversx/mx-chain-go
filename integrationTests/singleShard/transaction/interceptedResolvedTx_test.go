package transaction

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptTransactionWithMessenger(t *testing.T) {
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

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nRequester.Node.Start()
	_ = nResolver.Node.Start()
	defer func() {
		_ = nRequester.Node.Stop()
		_ = nResolver.Node.Stop()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	buffPk1, _ := nRequester.OwnAccount.SkTxSign.GeneratePublic().ToByteArray()

	valMinting := big.NewInt(100000)
	integrationTests.CreateMintingForSenders([]*integrationTests.TestProcessorNode{nRequester}, 0, []crypto.PrivateKey{nRequester.OwnAccount.SkTxSign}, valMinting)
	//Step 1. Generate a signed transaction
	tx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  integrationTests.TestHasher.Compute("receiver"),
		SndAddr:  buffPk1,
		Data:     "tx notarized data",
		GasLimit: integrationTests.MinTxGasLimit,
	}

	txBuff, _ := integrationTests.TestMarshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}
	tx.Signature, _ = signer.Sign(nRequester.OwnAccount.SkTxSign, txBuff)
	signedTxBuff, _ := integrationTests.TestMarshalizer.Marshal(&tx)

	fmt.Printf("Transaction: %v\n%v\n", tx, string(signedTxBuff))

	chanDone := make(chan bool)

	txHash := integrationTests.TestHasher.Compute(string(signedTxBuff))

	//step 2. wire up a received handler for requester
	nRequester.ShardDataPool.Transactions().RegisterHandler(func(key []byte) {
		txStored, _ := nRequester.ShardDataPool.Transactions().ShardDataStore(
			process.ShardCacherIdentifier(nRequester.ShardCoordinator.SelfId(), nRequester.ShardCoordinator.SelfId()),
		).Get(key)

		if reflect.DeepEqual(txStored, &tx) && tx.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, txStored, &tx)
		assert.Equal(t, txHash, key)
	})

	//Step 3. add the transaction in resolver pool
	nResolver.ShardDataPool.Transactions().AddData(
		txHash,
		&tx,
		process.ShardCacherIdentifier(nRequester.ShardCoordinator.SelfId(), nRequester.ShardCoordinator.SelfId()),
	)

	//Step 4. request tx
	txResolver, _ := nRequester.ResolverFinder.IntraShardResolver(factory.TransactionTopic)
	err = txResolver.RequestDataFromHash(txHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}

func TestNode_RequestInterceptRewardTransactionWithMessenger(t *testing.T) {
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

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nRequester.Node.Start()
	_ = nResolver.Node.Start()
	defer func() {
		_ = nRequester.Node.Stop()
		_ = nResolver.Node.Stop()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	//Step 1. Generate a reward Transaction
	tx := rewardTx.RewardTx{
		Value:   big.NewInt(0),
		RcvAddr: integrationTests.TestHasher.Compute("receiver"),
		Round:   0,
		Epoch:   0,
		ShardId: 0,
	}

	marshaledTxBuff, _ := integrationTests.TestMarshalizer.Marshal(&tx)

	fmt.Printf("Transaction: %v\n%v\n", tx, string(marshaledTxBuff))

	chanDone := make(chan bool)

	txHash := integrationTests.TestHasher.Compute(string(marshaledTxBuff))

	//step 2. wire up a received handler for requester
	nRequester.ShardDataPool.RewardTransactions().RegisterHandler(func(key []byte) {
		rewardTxStored, _ := nRequester.ShardDataPool.RewardTransactions().ShardDataStore(
			process.ShardCacherIdentifier(nRequester.ShardCoordinator.SelfId(), nRequester.ShardCoordinator.SelfId()),
		).Get(key)

		if reflect.DeepEqual(rewardTxStored, &tx) {
			chanDone <- true
		}

		assert.Equal(t, rewardTxStored, &tx)
		assert.Equal(t, txHash, key)
	})

	//Step 3. add the transaction in resolver pool
	nResolver.ShardDataPool.RewardTransactions().AddData(
		txHash,
		&tx,
		process.ShardCacherIdentifier(nRequester.ShardCoordinator.SelfId(), nRequester.ShardCoordinator.SelfId()),
	)

	//Step 4. request tx
	rewardTxResolver, _ := nRequester.ResolverFinder.IntraShardResolver(factory.RewardsTransactionTopic)
	err = rewardTxResolver.RequestDataFromHash(txHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
