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

func TestNode_GenerateSendInterceptHeaderByNonce(t *testing.T) {
	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	keyGen := schnorr.NewKeyGenerator()
	//sk, pk := keyGen.GeneratePair()
	//buffPk, _ := pk.ToByteArray()

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	blkcRequestor := createTestBlockChain()
	blkcResolver := createTestBlockChain()

	cp1, _ := p2p.NewConnectParamsFromPort(1)
	mes1, _ := p2p.NewMemMessenger(marshalizer, hasher, cp1)

	nRequestor, _ := node.NewNode(
		node.WithMessenger(mes1),
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

	cp2, _ := p2p.NewConnectParamsFromPort(2)
	mes2, _ := p2p.NewMemMessenger(marshalizer, hasher, cp2)

	nResolver, _ := node.NewNode(
		node.WithMessenger(mes2),
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

	mes1.Bootstrap(context.Background())
	mes2.Bootstrap(context.Background())

	defer p2p.ReInitializeGloballyRegisteredPeers()

	time.Sleep(time.Second)

	_ = nRequestor.BindInterceptorsResolvers()
	_ = nResolver.BindInterceptorsResolvers()

	//Step 1. Generate a header
	hdr := block.Header{
		Nonce:         0,
		PubKeysBitmap: []byte{255, 0},
		Commitment:    []byte("commitment"),
		Signature:     []byte("signature"),
		BlockBodyHash: []byte("block body hash"),
		PrevHash:      []byte("prev hash"),
		TimeStamp:     uint64(time.Now().Unix()),
		Round:         1,
		Epoch:         2,
		ShardId:       0,
		BlockBodyType: block.TxBlock,
	}

	hdrBuff, _ := marshalizer.Marshal(&hdr)
	hdrHash := hasher.Compute(string(hdrBuff))

	//Step 2. resolver has the header
	dPoolResolver.Headers().AddData(hdrHash, &hdr, 0)
	dPoolResolver.HeadersNonces().HasOrAdd(0, hdrHash)

	//Step 3. wire up a received handler
	chanDone := make(chan bool)

	dPoolRequestor.Headers().RegisterHandler(func(key []byte) {
		hdrStored, _ := dPoolRequestor.Headers().ShardDataStore(0).Get(key)

		if reflect.DeepEqual(hdrStored, &hdr) && hdr.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, hdrStored, &hdr)

	})

	//Step 4. request header
	hdrResolver := nRequestor.GetResolvers()[1].(*block2.HeaderResolver)
	hdrResolver.RequestHeaderFromNonce(0)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
