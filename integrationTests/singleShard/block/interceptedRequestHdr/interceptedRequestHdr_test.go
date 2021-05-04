package interceptedRequestHdr

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var durationTimeout = time.Second * 10

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

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)

	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := nResolver.ConnectTo(nRequester)
	require.Nil(t, err)

	time.Sleep(time.Second)

	hdr1, hdr2 := generateTwoHeaders(integrationTests.ChainID)
	hdrBuff1, _ := marshalizer.Marshal(hdr1)
	hdrHash1 := hasher.Compute(string(hdrBuff1))
	hdrBuff2, _ := marshalizer.Marshal(hdr2)
	hdrHash2 := hasher.Compute(string(hdrBuff2))

	//resolver has the headers
	nResolver.DataPool.Headers().AddHeader(hdrHash1, hdr1)

	_ = nResolver.Storage.GetStorer(dataRetriever.BlockHeaderUnit).Put(hdrHash2, hdrBuff2)
	_ = nResolver.Storage.GetStorer(dataRetriever.ShardHdrNonceHashDataUnit).Put(uint64Converter.ToByteSlice(1), hdrHash2)

	chanDone1, chanDone2 := wireUpHandler(nRequester, hdr1, hdr2)

	//request header from pool
	res, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, core.MetachainShardId)
	assert.Nil(t, err)
	hdrResolver := res.(*resolvers.HeaderResolver)
	_ = hdrResolver.RequestDataFromNonce(0, 0)

	//request header that is stored
	_ = hdrResolver.RequestDataFromNonce(1, 0)

	testChansShouldReadBoth(t, chanDone1, chanDone2)
}

func TestNode_InterceptedHeaderWithWrongChainIDShouldBeDiscarded(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := integrationTests.TestHasher
	marshalizer := integrationTests.TestMarshalizer
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId)

	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := nResolver.ConnectTo(nRequester)
	require.Nil(t, err)

	time.Sleep(time.Second)

	wrongChainID := []byte("wrong chain ID")
	hdr1, hdr2 := generateTwoHeaders(wrongChainID)
	hdrBuff1, _ := marshalizer.Marshal(hdr1)
	hdrHash1 := hasher.Compute(string(hdrBuff1))
	hdrBuff2, _ := marshalizer.Marshal(hdr2)
	hdrHash2 := hasher.Compute(string(hdrBuff2))

	//resolver has the headers
	nResolver.DataPool.Headers().AddHeader(hdrHash1, hdr1)

	_ = nResolver.Storage.GetStorer(dataRetriever.BlockHeaderUnit).Put(hdrHash2, hdrBuff2)
	_ = nResolver.Storage.GetStorer(dataRetriever.ShardHdrNonceHashDataUnit).Put(uint64Converter.ToByteSlice(1), hdrHash2)

	chanDone1, chanDone2 := wireUpHandler(nRequester, hdr1, hdr2)

	//request header from pool
	res, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, core.MetachainShardId)
	assert.Nil(t, err)
	hdrResolver := res.(*resolvers.HeaderResolver)
	_ = hdrResolver.RequestDataFromNonce(0, 0)

	//request header that is stored
	_ = hdrResolver.RequestDataFromNonce(1, 0)

	testChansShouldReadNone(t, chanDone1, chanDone2)
}

func generateTwoHeaders(chainID []byte) (data.HeaderHandler, data.HeaderHandler) {
	hdr1 := &block.Header{
		Nonce:            0,
		PubKeysBitmap:    []byte{255, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardID:          0,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 1),
		RandSeed:         make([]byte, 1),
		MiniBlockHeaders: nil,
		ChainID:          chainID,
		SoftwareVersion:  []byte("version"),
	}

	hdr2 := &block.Header{
		Nonce:            1,
		PubKeysBitmap:    []byte{255, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardID:          0,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 1),
		RandSeed:         make([]byte, 1),
		MiniBlockHeaders: nil,
		ChainID:          chainID,
		SoftwareVersion:  []byte("version"),
	}

	return hdr1, hdr2
}

func wireUpHandler(
	nRequester *integrationTests.TestProcessorNode,
	hdr1 data.HeaderHandler,
	hdr2 data.HeaderHandler,
) (chan struct{}, chan struct{}) {

	//wire up a received handler
	chanDone1 := make(chan struct{}, 1)
	chanDone2 := make(chan struct{}, 1)
	nRequester.DataPool.Headers().RegisterHandler(func(header data.HeaderHandler, key []byte) {
		fmt.Printf("Received hash %v\n", base64.StdEncoding.EncodeToString(key))

		if reflect.DeepEqual(header, hdr1) && hdr1.GetSignature() != nil {
			fmt.Printf("Received header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			chanDone1 <- struct{}{}
		}

		if reflect.DeepEqual(header, hdr2) && hdr2.GetSignature() != nil {
			fmt.Printf("Received header with hash %v\n", base64.StdEncoding.EncodeToString(key))
			chanDone2 <- struct{}{}
		}
	})

	return chanDone1, chanDone2
}

func testChansShouldReadBoth(t *testing.T, chanDone1 chan struct{}, chanDone2 chan struct{}) {
	select {
	case <-chanDone1:
	case <-time.After(durationTimeout):
		assert.Fail(t, "timeout")
		return
	}

	select {
	case <-chanDone2:
	case <-time.After(durationTimeout):
		assert.Fail(t, "timeout")
		return
	}
}

func testChansShouldReadNone(t *testing.T, chanDone1 chan struct{}, chanDone2 chan struct{}) {
	select {
	case <-chanDone1:
	case <-chanDone2:
	case <-time.After(durationTimeout):
		return
	}
	assert.Fail(t, "should have not received header with wrong chain ID")
}
