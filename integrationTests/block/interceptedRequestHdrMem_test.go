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

func TestNode_GenerateSendInterceptHeaderByNonceWithMemMessenger(t *testing.T) {
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
	res, err := pFactory1.ResolverContainer().Get(string(factory.HeadersTopic))
	assert.Nil(t, err)
	hdrResolver := res.(*block2.HeaderResolver)
	hdrResolver.RequestHeaderFromNonce(0)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
