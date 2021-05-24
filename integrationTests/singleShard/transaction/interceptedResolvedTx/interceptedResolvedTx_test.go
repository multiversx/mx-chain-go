package interceptedResolvedTx

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_RequestInterceptTransactionWithMessengerAndWhitelist(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := nRequester.ConnectTo(nResolver)
	require.Nil(t, err)

	time.Sleep(time.Second)

	buffPk1, _ := nRequester.OwnAccount.SkTxSign.GeneratePublic().ToByteArray()

	//minting the sender is no longer required as the requests are whitelisted

	//Step 1. Generate a signed transaction
	txData := "tx notarized data"
	//TODO change here when gas limit will no longer be linear with the tx data length
	txDataCost := uint64(len(txData))
	tx := transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  integrationTests.TestHasher.Compute("receiver"),
		SndAddr:  buffPk1,
		Data:     []byte(txData),
		GasLimit: integrationTests.MinTxGasLimit + txDataCost,
		GasPrice: integrationTests.MinTxGasPrice,
		ChainID:  integrationTests.ChainID,
		Version:  integrationTests.MinTransactionVersion,
	}

	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	signer := &ed25519SingleSig.Ed25519Signer{}
	tx.Signature, _ = signer.Sign(nRequester.OwnAccount.SkTxSign, txBuff)
	signedTxBuff, _ := integrationTests.TestMarshalizer.Marshal(&tx)

	chanDone := make(chan bool)

	txHash := integrationTests.TestHasher.Compute(string(signedTxBuff))

	//step 2. wire up a received handler for requester
	nRequester.DataPool.Transactions().RegisterOnAdded(func(key []byte, value interface{}) {
		txStored, _ := nRequester.DataPool.Transactions().ShardDataStore(
			process.ShardCacherIdentifier(nRequester.ShardCoordinator.SelfId(), nRequester.ShardCoordinator.SelfId()),
		).Get(key)

		if reflect.DeepEqual(txStored, &tx) && tx.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, txStored, &tx)
		assert.Equal(t, txHash, key)
	})

	//Step 3. add the transaction in resolver pool
	nResolver.DataPool.Transactions().AddData(
		txHash,
		&tx,
		tx.Size(),
		process.ShardCacherIdentifier(nRequester.ShardCoordinator.SelfId(), nRequester.ShardCoordinator.SelfId()),
	)

	//Step 4. request tx through request handler that will whitelist the hash
	nRequester.RequestHandler.RequestTransaction(0, [][]byte{txHash})
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

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := nRequester.ConnectTo(nResolver)
	require.Nil(t, err)

	time.Sleep(time.Second)

	//Step 1. Generate a reward Transaction
	_, pubKey, _ := integrationTests.GenerateSkAndPkInShard(nRequester.ShardCoordinator, nRequester.ShardCoordinator.SelfId())
	pubKeyArray, _ := pubKey.ToByteArray()
	tx := rewardTx.RewardTx{
		Value:   big.NewInt(0),
		RcvAddr: pubKeyArray,
		Round:   0,
		Epoch:   0,
	}

	marshaledTxBuff, _ := integrationTests.TestMarshalizer.Marshal(&tx)

	chanDone := make(chan bool)

	txHash := integrationTests.TestHasher.Compute(string(marshaledTxBuff))

	//step 2. wire up a received handler for requester
	nRequester.DataPool.RewardTransactions().RegisterOnAdded(func(key []byte, value interface{}) {
		rewardTxStored, _ := nRequester.DataPool.RewardTransactions().ShardDataStore(
			process.ShardCacherIdentifier(core.MetachainShardId, nRequester.ShardCoordinator.SelfId()),
		).Get(key)

		if reflect.DeepEqual(rewardTxStored, &tx) {
			chanDone <- true
		}

		assert.Equal(t, rewardTxStored, &tx)
		assert.Equal(t, txHash, key)
	})

	//Step 3. add the transaction in resolver pool
	nResolver.DataPool.RewardTransactions().AddData(
		txHash,
		&tx,
		tx.Size(),
		process.ShardCacherIdentifier(nRequester.ShardCoordinator.SelfId(), core.MetachainShardId),
	)

	//Step 4. request tx
	rewardTxResolver, _ := nRequester.ResolverFinder.CrossShardResolver(factory.RewardsTransactionTopic, core.MetachainShardId)
	err = rewardTxResolver.RequestDataFromHash(txHash, 0)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
