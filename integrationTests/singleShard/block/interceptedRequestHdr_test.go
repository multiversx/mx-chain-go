package block

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptHeaderByNonceWithNetMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	dPoolRequestor := createTestDataPool()
	storeRequestor := createTestStore()
	dPoolResolver := createTestDataPool()
	storeResolver := createTestStore()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requestor:")
	nRequestor, mesRequestor, _, resolversFinder := createNetNode(
		dPoolRequestor,
		storeRequestor,
		createAccountsDB(),
		shardCoordinator,
	)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, _ := createNetNode(
		dPoolResolver,
		storeResolver,
		createAccountsDB(),
		shardCoordinator,
	)

	nRequestor.Start()
	nResolver.Start()
	defer func() {
		_ = nRequestor.Stop()
		_ = nResolver.Stop()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := mesRequestor.ConnectToPeer(getConnectableAddress(mesResolver))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	//Step 1. Generate 2 headers, one will be stored in datapool, the other one in storage
	hdr1 := block.Header{
		Nonce:            0,
		PubKeysBitmap:    []byte{255, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardId:          0,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 0),
		RandSeed:         make([]byte, 0),
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
	}

	hdr2 := block.Header{
		Nonce:            1,
		PubKeysBitmap:    []byte{255, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardId:          0,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 0),
		RandSeed:         make([]byte, 0),
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
	}

	hdrBuff1, _ := marshalizer.Marshal(&hdr1)
	hdrHash1 := hasher.Compute(string(hdrBuff1))
	hdrBuff2, _ := marshalizer.Marshal(&hdr2)
	hdrHash2 := hasher.Compute(string(hdrBuff2))

	//Step 2. resolver has the headers
	dPoolResolver.Headers().HasOrAdd(hdrHash1, &hdr1)
	dPoolResolver.HeadersNonces().HasOrAdd(0, hdrHash1)
	storeResolver.GetStorer(dataRetriever.BlockHeaderUnit).Put(hdrHash2, hdrBuff2)
	storeResolver.GetStorer(dataRetriever.ShardHdrNonceHashDataUnit).Put(uint64Converter.ToByteSlice(1), hdrHash2)

	//Step 3. wire up a received handler
	chanDone := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- struct{}{}
	}()

	dPoolRequestor.Headers().RegisterHandler(func(key []byte) {
		hdrStored, _ := dPoolRequestor.Headers().Peek(key)
		fmt.Printf("Recieved hash %v\n", base64.StdEncoding.EncodeToString(key))

		if reflect.DeepEqual(hdrStored, &hdr1) && hdr1.Signature != nil {
			assert.Equal(t, hdrStored, &hdr1)
			fmt.Printf("Recieved header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			wg.Done()
		}

		if reflect.DeepEqual(hdrStored, &hdr2) && hdr2.Signature != nil {
			assert.Equal(t, hdrStored, &hdr2)
			fmt.Printf("Recieved header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			wg.Done()
		}
	})

	//Step 4. request header from pool
	res, err := resolversFinder.IntraShardResolver(factory.HeadersTopic)
	assert.Nil(t, err)
	hdrResolver := res.(*resolvers.HeaderResolver)
	hdrResolver.RequestDataFromNonce(0)

	//Step 5. request header that is stored
	hdrResolver.RequestDataFromNonce(1)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
