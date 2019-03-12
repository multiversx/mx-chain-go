package block

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTxBlockBodyWithNetMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	ti := &testInitializer{}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequestor := ti.createTestDataPool()
	dPoolResolver := ti.createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requestor:	")
	nRequestor, mesRequestor, _, resolversContainer := ti.createNetNode(
		32000,
		dPoolRequestor,
		ti.createAccountsDB(),
		shardCoordinator)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, _ := ti.createNetNode(
		32001,
		dPoolResolver,
		ti.createAccountsDB(),
		shardCoordinator)

	nRequestor.Start()
	nResolver.Start()

	defer nRequestor.Stop()
	defer nResolver.Stop()

	//connect messengers together
	time.Sleep(time.Second)
	err := mesRequestor.ConnectToPeer(ti.getConnectableAddress(mesResolver))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	//Step 1. Generate a block body
	body := block.Body{
		{
			ShardID: 0,
			TxHashes: [][]byte{
				hasher.Compute("tx1"),
			},
		},
	}

	miniBlock := body[0]
	miniBlockHashes := make([][]byte, 1)

	txBlockBodyBuff, _ := marshalizer.Marshal(miniBlock)
	txBlockBodyHash := hasher.Compute(string(txBlockBodyBuff))

	//Step 2. resolver has the tx block body
	dPoolResolver.MiniBlocks().HasOrAdd(txBlockBodyHash, miniBlock)
	fmt.Printf("Added %s to dPoolResolver\n", base64.StdEncoding.EncodeToString(txBlockBodyHash))

	//Step 3. wire up a received handler
	chanDone := make(chan bool)

	dPoolRequestor.MiniBlocks().RegisterHandler(func(key []byte) {
		txBlockBodyStored, _ := dPoolRequestor.MiniBlocks().Get(key)

		if reflect.DeepEqual(txBlockBodyStored, miniBlock) {
			chanDone <- true
		}

		assert.Equal(t, txBlockBodyStored, miniBlock)

	})

	//Step 4. request tx block body
	txBlockBodyRequestor, _ := resolversContainer.Get(factory.MiniBlocksTopic +
		shardCoordinator.CommunicationIdentifier(shardCoordinator.ShardForCurrentNode()))
	miniBlockRequestor := txBlockBodyRequestor.(process.MiniBlocksResolver)
	miniBlockHashes[0] = txBlockBodyHash
	miniBlockRequestor.RequestDataFromHashArray(miniBlockHashes)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 1000):
		assert.Fail(t, "timeout")
	}
}
