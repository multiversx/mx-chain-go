package interceptedResolvedUnsignedTx

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
	"github.com/stretchr/testify/require"
)

func TestNode_RequestInterceptUnsignedTransactionWithMessenger(t *testing.T) {
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

	//Step 1. Generate an unsigned transaction
	scr := &smartContractResult.SmartContractResult{
		Nonce:      0,
		Value:      big.NewInt(0),
		RcvAddr:    integrationTests.TestHasher.Compute("receiver"),
		SndAddr:    buffPk1,
		Data:       []byte("tx notarized data"),
		PrevTxHash: []byte("tx hash"),
	}

	scrBuff, _ := integrationTests.TestMarshalizer.Marshal(scr)
	fmt.Printf("Unsigned transaction: %v\n%v\n", scr, string(scrBuff))
	chanDone := make(chan bool)
	scrHash := integrationTests.TestHasher.Compute(string(scrBuff))

	//step 2. wire up a received handler for requester
	nRequester.DataPool.UnsignedTransactions().RegisterOnAdded(func(key []byte, value interface{}) {
		selfId := nRequester.ShardCoordinator.SelfId()
		scrStored, _ := nRequester.DataPool.UnsignedTransactions().ShardDataStore(
			process.ShardCacherIdentifier(selfId, selfId),
		).Get(key)

		if reflect.DeepEqual(scrStored, scr) {
			chanDone <- true
		}

		assert.Equal(t, scrStored, scr)
		assert.Equal(t, scrHash, key)
	})

	//Step 3. add the unsigned transaction in resolver pool
	nResolver.DataPool.UnsignedTransactions().AddData(
		scrHash,
		scr,
		scr.Size(),
		process.ShardCacherIdentifier(nResolver.ShardCoordinator.SelfId(), nResolver.ShardCoordinator.SelfId()),
	)

	//Step 4. request unsigned tx
	scrResolver, _ := nRequester.ResolverFinder.IntraShardResolver(factory.UnsignedTransactionTopic)
	err = scrResolver.RequestDataFromHash(scrHash, 0)
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 3):
		assert.Fail(t, "timeout")
	}
}
