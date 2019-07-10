package block

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data/block"
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

func TestNode_GenerateSendInterceptHeaderByNonceWithMemMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequestor := createTestDataPool()
	dPoolResolver := createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}
	nodesCoordinator1, _ := sharding.NewIndexHashedNodesCoordinator(1, hasher, 0, 1)
	nodesCoordinator2, _ := sharding.NewIndexHashedNodesCoordinator(1, hasher, 0, 1)

	fmt.Println("Requestor:")
	nRequestor, mesRequestor, _, pk1, multiSigner, resolversFinder := createNetNode(
		dPoolRequestor,
		createAccountsDB(),
		shardCoordinator,
		nodesCoordinator1,
	)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, pk2, _, _ := createNetNode(
		dPoolResolver,
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

	//Step 1. Generate a header
	hdr := block.Header{
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

	hdr.PubKeysBitmap = bitmap
	hdr.Signature = aggSig

	hdrBuff, _ = marshalizer.Marshal(&hdr)
	hdrHash := hasher.Compute(string(hdrBuff))

	//Step 2. resolver has the header
	dPoolResolver.Headers().HasOrAdd(hdrHash, &hdr)
	dPoolResolver.HeadersNonces().HasOrAdd(0, hdrHash)

	//Step 3. wire up a received handler
	chanDone := make(chan bool)

	dPoolRequestor.Headers().RegisterHandler(func(key []byte) {
		hdrStored, _ := dPoolRequestor.Headers().Peek(key)

		if reflect.DeepEqual(hdrStored, &hdr) && hdr.Signature != nil {
			chanDone <- true
		}

		assert.Equal(t, hdrStored, &hdr)

	})

	//Step 4. request header
	res, err := resolversFinder.IntraShardResolver(factory.HeadersTopic)
	assert.Nil(t, err)
	hdrResolver := res.(*resolvers.HeaderResolver)
	hdrResolver.RequestDataFromNonce(0)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
