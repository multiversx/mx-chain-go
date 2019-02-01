package block

import (
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTxBlockBodyWithMemMessenger(t *testing.T) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	nRequestor, _, pFactory1 := createMemNode(1, dPoolRequestor)
	nResolver, _, _ := createMemNode(2, dPoolResolver)

	nRequestor.Start()
	nResolver.Start()
	defer func() {
		_ = nRequestor.Stop()
		_ = nResolver.Stop()
	}()

	defer p2p.ReInitializeGloballyRegisteredPeers()

	time.Sleep(time.Second)

	//Step 1. Generate a block body
	txBlock := block.TxBlockBody{
		MiniBlocks: []block.MiniBlock{
			{
				ShardID: 0,
				TxHashes: [][]byte{
					hasher.Compute("tx1"),
				},
			},
		},
		StateBlockBody: block.StateBlockBody{
			RootHash: hasher.Compute("root hash"),
			ShardID:  0,
		},
	}

	txBlockBodyBuff, _ := marshalizer.Marshal(&txBlock)
	txBlockBodyHash := hasher.Compute(string(txBlockBodyBuff))

	//Step 2. resolver has the tx block body
	dPoolResolver.TxBlocks().HasOrAdd(txBlockBodyHash, &txBlock)

	//Step 3. wire up a received handler
	chanDone := make(chan bool)

	dPoolRequestor.TxBlocks().RegisterHandler(func(key []byte) {
		txBlockBodyStored, _ := dPoolRequestor.TxBlocks().Get(key)

		if reflect.DeepEqual(txBlockBodyStored, &txBlock) {
			chanDone <- true
		}

		assert.Equal(t, txBlockBodyStored, &txBlock)

	})

	//Step 4. request tx block body
	res, _ := pFactory1.ResolverContainer().Get(string(factory.TxBlockBodyTopic))
	hdrResolver := res.(*block2.GenericBlockBodyResolver)
	hdrResolver.RequestBlockBodyFromHash(txBlockBodyHash)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
