package broadcast_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func createDelayData(prefix string) ([]byte, *block.Header, map[uint32][]byte, map[string][][]byte) {
	miniblocks := make(map[uint32][]byte)
	receiverShardID := uint32(1)
	miniblocks[receiverShardID] = []byte(prefix + "miniblock data")

	transactions := make(map[string][][]byte)
	topic := "txBlockBodies_0_1"
	transactions[topic] = [][]byte{
		[]byte(prefix + "tx0"),
		[]byte(prefix + "tx1"),
	}
	headerHash := []byte(prefix + "header hash")
	header := &block.Header{
		Round:        0,
		PrevRandSeed: []byte(prefix),
	}

	return headerHash, header, miniblocks, transactions
}

func createInterceptorContainer() process.InterceptorsContainer {
	return &mock.InterceptorsContainerStub{
		GetCalled: func(topic string) (process.Interceptor, error) {
			return &mock.InterceptorStub{
				ProcessReceivedMessageCalled: func(message p2p.MessageP2P) error {
					return nil
				},
			}, nil
		},
	}
}

func createDefaultShardChainArgs() broadcast.ShardChainMessengerArgs {
	marshalizerMock := &mock.MarshalizerMock{}
	hasher := &mock.HasherMock{}
	messengerMock := &mock.MessengerStub{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := &mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptorsContainer := createInterceptorContainer()
	peerSigHandler := &mock.PeerSignatureHandler{
		Signer: singleSignerMock,
	}
	alarmScheduler := &mock.AlarmSchedulerStub{}

	return broadcast.ShardChainMessengerArgs{
		CommonMessengerArgs: broadcast.CommonMessengerArgs{
			Marshalizer:                marshalizerMock,
			Hasher:                     hasher,
			Messenger:                  messengerMock,
			PrivateKey:                 privateKeyMock,
			ShardCoordinator:           shardCoordinatorMock,
			PeerSignatureHandler:       peerSigHandler,
			HeadersSubscriber:          headersSubscriber,
			InterceptorsContainer:      interceptorsContainer,
			MaxDelayCacheSize:          1,
			MaxValidatorDelayCacheSize: 1,
			AlarmScheduler:             alarmScheduler,
		},
	}
}

func TestShardChainMessenger_NewShardChainMessengerNilMarshalizerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.Marshalizer = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilMessengerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.Messenger = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilMessenger, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilPrivateKeyShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.PrivateKey = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilPrivateKey, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilShardCoordinatorShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.ShardCoordinator = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilPeerSignatureHandlerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.PeerSignatureHandler = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilPeerSignatureHandler, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilInterceptorsContainerShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.InterceptorsContainer = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilInterceptorsContainer, err)
}

func TestShardChainMessenger_NewShardChainMessengerNilHeadersSubscriberShouldFail(t *testing.T) {
	args := createDefaultShardChainArgs()
	args.HeadersSubscriber = nil
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.Nil(t, scm)
	assert.Equal(t, spos.ErrNilHeadersSubscriber, err)
}

func TestShardChainMessenger_NewShardChainMessengerShouldWork(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, err := broadcast.NewShardChainMessenger(args)

	assert.NotNil(t, scm)
	assert.Equal(t, nil, err)
	assert.False(t, scm.IsInterfaceNil())
}

func TestShardChainMessenger_BroadcastBlockShouldErrNilBody(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(nil, &block.Header{})
	assert.Equal(t, spos.ErrNilBody, err)
}

func TestShardChainMessenger_BroadcastBlockShouldErrNilHeader(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(newTestBlockBody(), nil)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestShardChainMessenger_BroadcastBlockShouldErrMockMarshalizer(t *testing.T) {
	marshalizer := mock.MarshalizerMock{
		Fail: true,
	}
	args := createDefaultShardChainArgs()
	args.Marshalizer = marshalizer
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(newTestBlockBody(), &block.Header{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestShardChainMessenger_BroadcastBlockShouldWork(t *testing.T) {
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastBlock(newTestBlockBody(), &block.Header{})
	assert.Nil(t, err)
}

func TestShardChainMessenger_BroadcastMiniBlocksShouldBeDone(t *testing.T) {
	channelCalled := make(chan bool, 100)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	miniBlocks := make(map[uint32][]byte)
	miniBlocks[0] = make([]byte, 0)
	miniBlocks[1] = make([]byte, 0)
	miniBlocks[2] = make([]byte, 0)
	miniBlocks[3] = make([]byte, 0)
	err := scm.BroadcastMiniBlocks(miniBlocks)

	called := 0
	for i := 0; i < 4; i++ {
		select {
		case <-channelCalled:
			called++
		case <-time.After(time.Millisecond * 100):
			break
		}
	}

	assert.Nil(t, err)
	assert.Equal(t, 4, called)
}

func TestShardChainMessenger_BroadcastTransactionsShouldNotBeCalled(t *testing.T) {
	channelCalled := make(chan bool)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	transactions := make(map[string][][]byte)
	err := scm.BroadcastTransactions(transactions)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.False(t, wasCalled)

	transactions[factory.TransactionTopic] = make([][]byte, 0)
	err = scm.BroadcastTransactions(transactions)

	wasCalled = false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.False(t, wasCalled)
}

func TestShardChainMessenger_BroadcastTransactionsShouldBeCalled(t *testing.T) {
	channelCalled := make(chan bool)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}

	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	transactions := make(map[string][][]byte)
	txs := make([][]byte, 0)
	txs = append(txs, []byte(""))
	transactions[factory.TransactionTopic] = txs
	err := scm.BroadcastTransactions(transactions)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestShardChainMessenger_BroadcastHeaderNilHeaderShouldErr(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	err := scm.BroadcastHeader(nil)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestShardChainMessenger_BroadcastHeaderShouldWork(t *testing.T) {
	channelCalled := make(chan bool)

	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			channelCalled <- true
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	hdr := block.MetaBlock{Nonce: 10}
	err := scm.BroadcastHeader(&hdr)

	wasCalled := false
	select {
	case <-channelCalled:
		wasCalled = true
	case <-time.After(time.Millisecond * 100):
	}

	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestShardChainMessenger_BroadcastBlockDataLeaderNilHeaderShouldErr(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	_, _, miniblocks, transactions := createDelayData("1")

	err := scm.BroadcastBlockDataLeader(nil, miniblocks, transactions)
	assert.Equal(t, spos.ErrNilHeader, err)
}

func TestShardChainMessenger_BroadcastBlockDataLeaderNilMiniblocksShouldReturnNil(t *testing.T) {
	args := createDefaultShardChainArgs()
	scm, _ := broadcast.NewShardChainMessenger(args)

	_, header, _, transactions := createDelayData("1")

	err := scm.BroadcastBlockDataLeader(header, nil, transactions)
	assert.Nil(t, err)
}

func TestShardChainMessenger_BroadcastBlockDataLeaderShouldTriggerWaitingDelayedMessage(t *testing.T) {
	wasCalled := atomic.Flag{}
	messenger := &mock.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			wasCalled.Set()
		},
	}
	args := createDefaultShardChainArgs()
	args.Messenger = messenger
	scm, _ := broadcast.NewShardChainMessenger(args)

	_, header, miniBlocksMarshalled, transactions := createDelayData("1")
	err := scm.BroadcastBlockDataLeader(header, miniBlocksMarshalled, transactions)
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, err)
	assert.False(t, wasCalled.IsSet())

	wasCalled.Unset()
	_, header2, miniBlocksMarshalled2, transactions2 := createDelayData("2")
	err = scm.BroadcastBlockDataLeader(header2, miniBlocksMarshalled2, transactions2)
	time.Sleep(10 * time.Millisecond)
	assert.Nil(t, err)
	assert.True(t, wasCalled.IsSet())
}
