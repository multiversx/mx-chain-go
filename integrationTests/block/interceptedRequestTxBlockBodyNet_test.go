package block

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTxBlockBodyWithNetMessenger(t *testing.T) {
	t.Skip("TODO: fix tests that run on the same local network")

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	keyGen := schnorr.NewKeyGenerator()

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	blkcRequestor := createTestBlockChain()
	blkcResolver := createTestBlockChain()

	nRequestor, _ := node.NewNode(
		node.WithPort(4000),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithDataPool(dPoolRequestor),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(blkcRequestor),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
	)

	nResolver, _ := node.NewNode(
		node.WithPort(4001),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithMaxAllowedPeers(4),
		node.WithContext(context.Background()),
		node.WithPubSubStrategy(p2p.GossipSub),
		node.WithDataPool(dPoolResolver),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(&sharding.OneShardCoordinator{}),
		node.WithBlockChain(blkcResolver),
		node.WithUint64ByteSliceConverter(uint64ByteSlice.NewBigEndianConverter()),
	)

	nRequestor.Start()
	nResolver.Start()

	defer nRequestor.Stop()
	defer nResolver.Stop()

	nRequestor.P2PBootstrap()
	nResolver.P2PBootstrap()

	time.Sleep(time.Second)

	_ = nRequestor.BindInterceptorsResolvers()
	_ = nResolver.BindInterceptorsResolvers()

	//TODO remove this
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
	txBlockBodyResolver := nRequestor.GetResolvers()[2].(*block2.GenericBlockBodyResolver)
	txBlockBodyResolver.RequestBlockBodyFromHash(txBlockBodyHash)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
