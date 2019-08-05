package transaction

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptTransactionWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	dPoolRequester := createTestDataPool()
	dPoolResolver := createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}
	nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		testHasher,
		0,
		1,
		make(map[uint32][]sharding.Validator),
	)

	fmt.Println("Requester:")
	nRequester, mesRequester, sk1, resolversFinder := createNetNode(
		dPoolRequester,
		createAccountsDB(),
		shardCoordinator,
		nodesCoordinator,
	)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, _ := createNetNode(
		dPoolResolver,
		createAccountsDB(),
		shardCoordinator,
		nodesCoordinator,
	)

	_ = nRequester.Start()
	_ = nResolver.Start()
	defer func() {
		_ = nRequester.Stop()
		_ = nResolver.Stop()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := mesRequester.ConnectToPeer(getConnectableAddress(mesResolver))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	buffPk1, _ := sk1.GeneratePublic().ToByteArray()

	//Step 1. Generate a signed transaction
	tx := transaction.Transaction{
		Nonce:   0,
		Value:   big.NewInt(0),
		RcvAddr: testHasher.Compute("receiver"),
		SndAddr: buffPk1,
		Data:    "tx notarized data",
	}

	txBuff, _ := testMarshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}

	tx.Signature, _ = signer.Sign(sk1, txBuff)

	signedTxBuff, _ := testMarshalizer.Marshal(&tx)

	fmt.Printf("Transaction: %v\n%v\n", tx, string(signedTxBuff))

	chanDone := make(chan bool)

	txHash := testHasher.Compute(string(signedTxBuff))

	//step 2. wire up a received handler for requester
	dPoolRequester.Transactions().RegisterHandler(func(key []byte) {
		txStored, _ := dPoolRequester.Transactions().ShardDataStore(
			process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
		).Get(key)

		if reflect.DeepEqual(txStored, &tx) && tx.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, txStored, &tx)
		assert.Equal(t, txHash, key)
	})

	//Step 3. add the transaction in resolver pool
	dPoolResolver.Transactions().AddData(
		txHash,
		&tx,
		process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
	)

	//Step 4. request tx
	txResolver, _ := resolversFinder.IntraShardResolver(factory.TransactionTopic)
	err = txResolver.RequestDataFromHash(txHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
