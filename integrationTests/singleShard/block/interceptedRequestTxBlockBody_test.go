package block

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTxBlockBodyWithNetMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequester := createTestDataPool()
	dPoolResolver := createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requester:	")
	nRequester, mesRequester, _, resolversFinder := createNetNode(
		dPoolRequester,
		createTestStore(),
		createAccountsDB(),
		shardCoordinator,
	)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, _ := createNetNode(
		dPoolResolver,
		createTestStore(),
		createAccountsDB(),
		shardCoordinator,
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

	//Step 1. Generate a block body
	body := block.Body{
		{
			ReceiverShardID: 0,
			SenderShardID:   0,
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

	dPoolRequester.MiniBlocks().RegisterHandler(func(key []byte) {
		txBlockBodyStored, _ := dPoolRequester.MiniBlocks().Get(key)

		if reflect.DeepEqual(txBlockBodyStored, miniBlock) {
			chanDone <- true
		}

		assert.Equal(t, txBlockBodyStored, miniBlock)

	})

	//Step 4. request tx block body
	txBlockBodyRequester, _ := resolversFinder.IntraShardResolver(factory.MiniBlocksTopic)
	miniBlockRequester := txBlockBodyRequester.(dataRetriever.MiniBlocksResolver)
	miniBlockHashes[0] = txBlockBodyHash
	_ = miniBlockRequester.RequestDataFromHashArray(miniBlockHashes)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
