package block

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptTxBlockBodyWithNetMessenger(t *testing.T) {
	t.Skip("TODO: fix tests that run on the same local network")

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	suite := kv2.NewBlakeSHA256Ed25519()
	keyGen := signing.NewKeyGenerator(suite)

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	blkcRequestor := createTestBlockChain()
	blkcResolver := createTestBlockChain()
	reqMessenger := createMessenger(context.Background(), marshalizer, hasher, 4, 32000)
	resMessenger := createMessenger(context.Background(), marshalizer, hasher, 4, 32001)
	shardCoordinatorReq := &sharding.OneShardCoordinator{}
	shardCoordinatorRes := &sharding.OneShardCoordinator{}
	uint64BsReq := uint64ByteSlice.NewBigEndianConverter()
	uint64BsRes := uint64ByteSlice.NewBigEndianConverter()

	pFactoryReq, _ := factory.NewProcessorsCreator(factory.ProcessorsCreatorConfig{
		InterceptorContainer:     interceptor.NewContainer(),
		ResolverContainer:        resolver.NewContainer(),
		Messenger:                reqMessenger,
		Blockchain:               blkcRequestor,
		DataPool:                 dPoolRequestor,
		ShardCoordinator:         shardCoordinatorReq,
		AddrConverter:            addrConverter,
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		SingleSignKeyGen:         keyGen,
		Uint64ByteSliceConverter: uint64BsReq,
	})

	pFactoryRes, _ := factory.NewProcessorsCreator(factory.ProcessorsCreatorConfig{
		InterceptorContainer:     interceptor.NewContainer(),
		ResolverContainer:        resolver.NewContainer(),
		Messenger:                resMessenger,
		Blockchain:               blkcResolver,
		DataPool:                 dPoolResolver,
		ShardCoordinator:         shardCoordinatorRes,
		AddrConverter:            addrConverter,
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		SingleSignKeyGen:         keyGen,
		Uint64ByteSliceConverter: uint64BsRes,
	})

	nRequestor, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithContext(context.Background()),
		node.WithDataPool(dPoolRequestor),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(shardCoordinatorReq),
		node.WithBlockChain(blkcRequestor),
		node.WithUint64ByteSliceConverter(uint64BsReq),
		node.WithMessenger(reqMessenger),
		node.WithProcessorCreator(pFactoryReq),
	)

	nResolver, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithContext(context.Background()),
		node.WithDataPool(dPoolResolver),
		node.WithAddressConverter(addrConverter),
		node.WithSingleSignKeyGenerator(keyGen),
		node.WithShardCoordinator(shardCoordinatorRes),
		node.WithBlockChain(blkcResolver),
		node.WithUint64ByteSliceConverter(uint64BsRes),
		node.WithMessenger(resMessenger),
		node.WithProcessorCreator(pFactoryRes),
	)

	nRequestor.Start()
	nResolver.Start()

	defer nRequestor.Stop()
	defer nResolver.Stop()

	nRequestor.P2PBootstrap()
	nResolver.P2PBootstrap()

	time.Sleep(time.Second)

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
	res, _ := pFactoryRes.ResolverContainer().Get(string(factory.TxBlockBodyTopic))
	txBlockBodyResolver := res.(*block2.GenericBlockBodyResolver)
	txBlockBodyResolver.RequestBlockBodyFromHash(txBlockBodyHash)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}

func createMessenger(ctx context.Context, marshalizer marshal.Marshalizer, hasher hashing.Hasher, maxAllowedPeers int, port int) p2p.Messenger {
	cp := &p2p.ConnectParams{}
	cp.Port = port
	cp.GeneratePrivPubKeys(time.Now().UnixNano())
	cp.GenerateIDFromPubKey()

	nm, _ := p2p.NewNetMessenger(ctx, marshalizer, hasher, cp, maxAllowedPeers, p2p.GossipSub)
	return nm
}
