package transaction

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptUnsignedTransactionWithMessenger(t *testing.T) {
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

	//Step 1. Generate an unsigned transaction
	scr := &smartContractResult.SmartContractResult{
		Nonce:   0,
		Value:   big.NewInt(0),
		RcvAddr: integrationTests.TestHasher.Compute("receiver"),
		SndAddr: buffPk1,
		Data:    []byte("tx notarized data"),
		TxHash:  []byte("tx hash"),
	}

	scrBuff, _ := integrationTests.TestMarshalizer.Marshal(scr)
	fmt.Printf("Unsigned transaction: %v\n%v\n", scr, string(scrBuff))
	chanDone := make(chan bool)
	scrHash := integrationTests.TestHasher.Compute(string(scrBuff))

	//step 2. wire up a received handler for requester
	nRequester.ShardDataPool.UnsignedTransactions().RegisterHandler(func(key []byte) {
		selfId := nRequester.ShardCoordinator.SelfId()
		scrStored, _ := nRequester.ShardDataPool.UnsignedTransactions().ShardDataStore(
			process.ShardCacherIdentifier(selfId, selfId),
		).Get(key)

		if reflect.DeepEqual(scrStored, scr) {
			chanDone <- true
		}

		assert.Equal(t, scrStored, scr)
		assert.Equal(t, scrHash, key)
	})

	//Step 3. add the unsigned transaction in resolver pool
	nResolver.ShardDataPool.UnsignedTransactions().AddData(
		scrHash,
		scr,
		process.ShardCacherIdentifier(nResolver.ShardCoordinator.SelfId(), nResolver.ShardCoordinator.SelfId()),
	)

	//Step 4. request unsigned tx
	scrResolver, _ := nRequester.ResolverFinder.IntraShardResolver(factory.UnsignedTransactionTopic)
	err = scrResolver.RequestDataFromHash(scrHash)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
