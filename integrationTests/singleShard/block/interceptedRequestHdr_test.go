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
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
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
	dPoolRequester := createTestDataPool()
	storeRequester := createTestStore()
	dPoolResolver := createTestDataPool()
	storeResolver := createTestStore()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requester:")
	nRequester, mesRequester, _, resolversFinder := createNetNode(
		dPoolRequester,
		storeRequester,
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
	_, _ = dPoolResolver.Headers().HasOrAdd(hdrHash1, &hdr1)

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(0, hdrHash1)
	dPoolResolver.HeadersNonces().Merge(0, syncMap)
	_ = storeResolver.GetStorer(dataRetriever.BlockHeaderUnit).Put(hdrHash2, hdrBuff2)
	_ = storeResolver.GetStorer(dataRetriever.ShardHdrNonceHashDataUnit).Put(uint64Converter.ToByteSlice(1), hdrHash2)

	//Step 3. wire up a received handler
	chanDone := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- struct{}{}
	}()

	dPoolRequester.Headers().RegisterHandler(func(key []byte) {
		hdrStored, _ := dPoolRequester.Headers().Peek(key)
		fmt.Printf("Recieved hash %v\n", base64.StdEncoding.EncodeToString(key))

		if reflect.DeepEqual(hdrStored, &hdr1) && hdr1.Signature != nil {
			fmt.Printf("Recieved header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			wg.Done()
		}

		if reflect.DeepEqual(hdrStored, &hdr2) && hdr2.Signature != nil {
			fmt.Printf("Recieved header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			wg.Done()
		}
	})

	//Step 4. request header from pool
	res, err := resolversFinder.IntraShardResolver(factory.HeadersTopic)
	assert.Nil(t, err)
	hdrResolver := res.(*resolvers.HeaderResolver)
	_ = hdrResolver.RequestDataFromNonce(0)

	//Step 5. request header that is stored
	_ = hdrResolver.RequestDataFromNonce(1)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
