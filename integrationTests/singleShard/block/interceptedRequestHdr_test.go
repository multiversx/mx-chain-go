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
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptHeaderByNonceWithNetMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := integrationTests.TestHasher
	marshalizer := integrationTests.TestMarshalizer
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	requesterNodeAddr := "0"
	resolverNodeAddr := "1"

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, requesterNodeAddr)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nRequester.Node.Start()
	_ = nResolver.Node.Start()
	defer func() {
		_ = nRequester.Node.Stop()
		_ = nResolver.Node.Stop()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	//Step 1. Generate 2 headers, one will be stored in datapool, the other one in storage
	hdr1 := block.Header{
		Nonce:         1,
		PubKeysBitmap: []byte{255, 0},
		Signature:     []byte("signature"),
		PrevHash:      []byte("prev hash"),
		TimeStamp:     uint64(time.Now().Unix()),
		Round:         1,
		Epoch:         2,
		ShardID:       0,
		BlockBodyType: block.TxBlock,
		RootHash:      []byte{255, 255},
		PrevRandSeed:  []byte{1, 2},
		RandSeed:      []byte{3, 4},
	}

	hdr2 := block.Header{
		Nonce:         2,
		PubKeysBitmap: []byte{255, 0},
		Signature:     []byte("signature"),
		PrevHash:      []byte("prev hash"),
		TimeStamp:     uint64(time.Now().Unix()),
		Round:         1,
		Epoch:         2,
		ShardID:       0,
		BlockBodyType: block.TxBlock,
		RootHash:      []byte{255, 255},
		PrevRandSeed:  []byte{1, 2},
		RandSeed:      []byte{3, 4},
	}

	hdrBuff1, err := marshalizer.Marshal(&hdr1)
	assert.Nil(t, err)
	hdrHash1 := hasher.Compute(string(hdrBuff1))
	t.Logf("Hdr1:  %v,\n%#v", hdrHash1, hdr1)

	hdrBuff2, err := marshalizer.Marshal(&hdr2)
	assert.Nil(t, err)
	hdrHash2 := hasher.Compute(string(hdrBuff2))
	t.Logf("Hdr2:  %v,\n%#v", hdrHash2, hdr2)

	//Step 2. resolver has the headers
	_, _ = nResolver.ShardDataPool.Headers().HasOrAdd(hdrHash1, &hdr1)

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(0, hdrHash1)
	nResolver.ShardDataPool.HeadersNonces().Merge(0, syncMap)
	err = nResolver.Storage.GetStorer(dataRetriever.BlockHeaderUnit).Put(hdrHash2, hdrBuff2)
	assert.Nil(t, err)
	err = nResolver.Storage.GetStorer(dataRetriever.ShardHdrNonceHashDataUnit).Put(uint64Converter.ToByteSlice(1), hdrHash2)
	assert.Nil(t, err)

	//Step 3. wire up a received handler
	chanDone := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		wg.Wait()
		chanDone <- struct{}{}
	}()

	nRequester.ShardDataPool.Headers().RegisterHandler(func(key []byte) {
		fmt.Printf("Received hash %v\n", base64.StdEncoding.EncodeToString(key))

		hdrStored, ok := nRequester.ShardDataPool.Headers().Peek(key)
		assert.Truef(t, ok, "key lookup faild for %s", base64.StdEncoding.EncodeToString(key))

		if reflect.DeepEqual(hdrStored, &hdr1) && hdr1.Signature != nil {
			fmt.Printf("Received header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			wg.Done()
			return
		}

		if reflect.DeepEqual(hdrStored, &hdr2) && hdr2.Signature != nil {
			fmt.Printf("Received header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			wg.Done()
			return
		}

		assert.Failf(t, "Unknown header ", "%#v\nh1:\n%#v\nh2:\n%#v", hdrStored, &hdr1, &hdr2)

	})

	//Step 4. request header from pool
	res, err := nRequester.ResolverFinder.IntraShardResolver(factory.HeadersTopic)
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
