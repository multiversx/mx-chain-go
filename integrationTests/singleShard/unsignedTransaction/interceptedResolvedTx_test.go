package unsignedTransaction

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptUnsignedTransactionWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequester := createTestDataPool()
	dPoolResolver := createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requester:")
	nRequester, mesRequester, sk1, resolversFinder := createNetNode(
		dPoolRequester,
		createAccountsDB(),
		shardCoordinator)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, _ := createNetNode(
		dPoolResolver,
		createAccountsDB(),
		shardCoordinator)

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

	//Step 1. Generate an unsigned transaction
	scr := &smartContractResult.SmartContractResult{
		Nonce:   0,
		Value:   big.NewInt(0),
		RcvAddr: hasher.Compute("receiver"),
		SndAddr: buffPk1,
		Data:    "tx notarized data",
		TxHash:  []byte("tx hash"),
	}

	scrBuff, _ := marshalizer.Marshal(scr)

	fmt.Printf("Unsigned transaction: %v\n%v\n", scr, string(scrBuff))

	chanDone := make(chan bool)

	scrHash := hasher.Compute(string(scrBuff))

	//step 2. wire up a received handler for requester
	dPoolRequester.UnsignedTransactions().RegisterHandler(func(key []byte) {
		scrStored, _ := dPoolRequester.UnsignedTransactions().ShardDataStore(
			process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
		).Get(key)

		if reflect.DeepEqual(scrStored, scr) {
			chanDone <- true
		}

		assert.Equal(t, scrStored, scr)
		assert.Equal(t, scrHash, key)
	})

	//Step 3. add the unsigned transaction in resolver pool
	dPoolResolver.UnsignedTransactions().AddData(
		scrHash,
		scr,
		process.ShardCacherIdentifier(shardCoordinator.SelfId(), shardCoordinator.SelfId()),
	)

	//Step 4. request unsigned tx
	scrResolver, _ := resolversFinder.IntraShardResolver(factory.UnsignedTransactionTopic)
	err = scrResolver.RequestDataFromHash(scrHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
