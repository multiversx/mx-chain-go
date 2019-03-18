package block

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/stretchr/testify/assert"
)

func TestNode_GenerateSendInterceptHeaderByNonceWithMemMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	ti := &testInitializer{}

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	dPoolRequestor := ti.createTestDataPool()
	dPoolResolver := ti.createTestDataPool()

	shardCoordinator := &sharding.OneShardCoordinator{}

	fmt.Println("Requestor:")
	nRequestor, mesRequestor, _, resolversContainer := ti.createNetNode(
		33000,
		dPoolRequestor,
		ti.createAccountsDB(),
		shardCoordinator)

	fmt.Println("Resolver:")
	nResolver, mesResolver, _, _ := ti.createNetNode(
		33001,
		dPoolResolver,
		ti.createAccountsDB(),
		shardCoordinator)

	nRequestor.Start()
	nResolver.Start()
	defer func() {
		_ = nRequestor.Stop()
		_ = nResolver.Stop()
	}()

	//connect messengers together
	time.Sleep(time.Second)
	err := mesRequestor.ConnectToPeer(ti.getConnectableAddress(mesResolver))
	assert.Nil(t, err)

	time.Sleep(time.Second)

	//Step 1. Generate a header
	hdr := block.Header{
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
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
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
	res, err := resolversContainer.Get(factory.HeadersTopic +
		shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId()))
	assert.Nil(t, err)
	hdrResolver := res.(*resolvers.HeaderResolver)
	hdrResolver.RequestDataFromNonce(0)

	select {
	case <-chanDone:
	case <-time.After(time.Second * 10):
		assert.Fail(t, "timeout")
	}
}
