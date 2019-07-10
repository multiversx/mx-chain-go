package block

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
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

func genValidatorsFromPubKeys(pubKeysMap map[uint32][]string) map[uint32][]sharding.Validator {
	validatorsMap := make(map[uint32][]sharding.Validator)

	for shardId, shardNodesPks := range pubKeysMap {
		shardValidators := make([]sharding.Validator, 0)
		for i := 0; i < len(shardNodesPks); i++ {
			v, _ := consensus.NewValidator(big.NewInt(0), 1, []byte(shardNodesPks[i]))
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

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
	nodesCoordinator1, _ := sharding.NewIndexHashedNodesCoordinator(1, hasher, 0, 1)
	nodesCoordinator2, _ := sharding.NewIndexHashedNodesCoordinator(1, hasher, 0, 1)

	fmt.Println("Requester:")
	nRequester, mesRequester, _, pk1, multiSigner, resolversFinder := createNetNode(
		dPoolRequester,
		storeRequester,
		createAccountsDB(),
		shardCoordinator,
		nodesCoordinator1,
	)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, pk2, _, _ := createNetNode(
		dPoolResolver,
		storeResolver,
		createAccountsDB(),
		shardCoordinator,
		nodesCoordinator2,
	)

	pubKeyMap := make(map[uint32][]string)
	shard0PubKeys := make([]string, 2)
	pk1Bytes, _ := pk1.ToByteArray()
	pk2Bytes, _ := pk2.ToByteArray()

	shard0PubKeys[0] = string(pk1Bytes)
	shard0PubKeys[1] = string(pk2Bytes)

	pubKeyMap[0] = shard0PubKeys
	validatorsMap := genValidatorsFromPubKeys(pubKeyMap)
	nodesCoordinator1.LoadNodesPerShards(validatorsMap)
	nodesCoordinator2.LoadNodesPerShards(validatorsMap)

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
		PubKeysBitmap:    nil,
		Signature:        nil,
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
		Nonce:            0,
		PubKeysBitmap:    nil,
		Signature:        nil,
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

	hdrBuff, _ := marshalizer.Marshal(&hdr)
	msig, _ := multiSigner.Create(shard0PubKeys, 0)
	bitmap := []byte{1, 0, 0}
	_, _ = msig.CreateSignatureShare(hdrBuff, bitmap)
	aggSig, _ := msig.AggregateSigs(bitmap)

	hdr1.PubKeysBitmap = bitmap
	hdr1.Signature = aggSig
	hdr2.PubKeysBitmap = bitmap
	hdr2.Signature = aggSig

	hdrBuff1, _ := marshalizer.Marshal(&hdr1)
	hdrHash1 := hasher.Compute(string(hdrBuff1))
	hdrBuff2, _ := marshalizer.Marshal(&hdr2)
	hdrHash2 := hasher.Compute(string(hdrBuff2))

	//Step 2. resolver has the headers
	_, _ = dPoolResolver.Headers().HasOrAdd(hdrHash1, &hdr1)
	_, _ = dPoolResolver.HeadersNonces().HasOrAdd(0, hdrHash1)
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
