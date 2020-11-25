package broadcast_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultMetaChainArgs() broadcast.MetaChainMessengerArgs {
	marshalizerMock := &mock.MarshalizerMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	hasher := mock.HasherMock{}
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptorsContainer := createInterceptorContainer()
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock}
	alarmScheduler := &mock.AlarmSchedulerStub{}

	return broadcast.MetaChainMessengerArgs{
		CommonMessengerArgs: broadcast.CommonMessengerArgs{
			Marshalizer:                marshalizerMock,
			Hasher:                     hasher,
			Messenger:                  messengerMock,
			PrivateKey:                 privateKeyMock,
			ShardCoordinator:           shardCoordinatorMock,
			PeerSignatureHandler:       peerSigHandler,
			HeadersSubscriber:          headersSubscriber,
			InterceptorsContainer:      interceptorsContainer,
			MaxValidatorDelayCacheSize: 2,
			MaxDelayCacheSize:          2,
			AlarmScheduler:             alarmScheduler,
		},
	}
}

func TestMetaChainMessenger_NewMetaChainMessengerNilMarshalizerShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.Marshalizer = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilMessengerShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.Messenger = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilMessenger, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilPrivateKeyShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.PrivateKey = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilPrivateKey, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilShardCoordinatorShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.ShardCoordinator = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerNilPeerSignatureHandlerShouldFail(t *testing.T) {
	args := createDefaultMetaChainArgs()
	args.PeerSignatureHandler = nil
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.Nil(t, mcm)
	assert.Equal(t, spos.ErrNilPeerSignatureHandler, err)
}

func TestMetaChainMessenger_NewMetaChainMessengerShouldWork(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, err := broadcast.NewMetaChainMessenger(args)

	assert.NotNil(t, mcm)
	assert.Equal(t, nil, err)
	assert.False(t, mcm.IsInterfaceNil())
}

func TestMetaChainMessenger_BroadcastBlockShouldErrNilMetaHeader(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastBlock(newTestBlockBody(), nil)
	assert.Equal(t, spos.ErrNilMetaHeader, err)
}

func TestMetaChainMessenger_BroadcastBlockShouldErrMockMarshalizer(t *testing.T) {
	marshalizer := &mock.MarshalizerMock{
		Fail: true,
	}
	args := createDefaultMetaChainArgs()
	args.Marshalizer = marshalizer
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastBlock(newTestBlockBody(), &block.MetaBlock{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestMetaChainMessenger_BroadcastBlockShouldWork(t *testing.T) {
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	args := createDefaultMetaChainArgs()
	args.Messenger = messenger
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastBlock(newTestBlockBody(), &block.MetaBlock{})
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastMiniBlocksShouldWork(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastMiniBlocks(nil)
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastTransactionsShouldWork(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastTransactions(nil)
	assert.Nil(t, err)
}

func TestMetaChainMessenger_BroadcastHeaderNilHeaderShouldErr(t *testing.T) {
	args := createDefaultMetaChainArgs()
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	err := mcm.BroadcastHeader(nil)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestMetaChainMessenger_BroadcastHeaderOkHeaderShouldWork(t *testing.T) {
	channelCalled := make(chan bool)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	args := createDefaultMetaChainArgs()
	args.Messenger = messenger
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	hdr := block.Header{
		Nonce: 10,
	}

	err := mcm.BroadcastHeader(&hdr)
	assert.Nil(t, err)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestMetaChainMessenger_BroadcastBlockDataLeader(t *testing.T) {
	countersBroadcast := make(map[string]int)
	mutCounters := &sync.Mutex{}

	messengerMock := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			mutCounters.Lock()
			countersBroadcast[topic]++
			mutCounters.Unlock()
		},
	}

	args := createDefaultMetaChainArgs()
	args.Messenger = messengerMock
	mcm, _ := broadcast.NewMetaChainMessenger(args)

	miniBlocks := map[uint32][]byte{0: []byte("mbs data1"), 1: []byte("mbs data2")}
	transactions := map[string][][]byte{"topic1": {[]byte("txdata1"), []byte("txdata2")}, "topic2": {[]byte("txdata3")}}

	err := mcm.BroadcastBlockDataLeader(nil, miniBlocks, transactions)
	require.Nil(t, err)
	sleepTime := core.ExtraDelayBetweenBroadcastMbsAndTxs +
		core.ExtraDelayForBroadcastBlockInfo +
		time.Millisecond*100
	time.Sleep(sleepTime)

	mutCounters.Lock()
	defer mutCounters.Unlock()

	assert.Equal(t, len(miniBlocks), countersBroadcast["txBlockBodies_0"]+countersBroadcast["txBlockBodies_0_1"])
	assert.Equal(t, len(transactions), countersBroadcast["topic1"]+countersBroadcast["topic2"])
}
