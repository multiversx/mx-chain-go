package spos_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

const roundTimeDuration = 100 * time.Millisecond

func initWorker() *spos.Worker {
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{
		DecodeBlockHeaderCalled: func(dta []byte) data.HeaderHandler {
			return nil
		},
		RevertAccountStateCalled: func() {
		},
	}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	forkDetectorMock.AddHeaderCalled = func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalsHeaders []data.HeaderHandler, finalHeadersHashes [][]byte, isNotarizedShardStuck bool) error {
		return nil
	}
	keyGeneratorMock, _, _ := mock.InitKeys()
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte("signed"), nil
		},
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
	}
	syncTimerMock := &mock.SyncTimerMock{}

	bnService, _ := bn.NewConsensusService()

	sposWorker, _ := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	return sposWorker
}

func initRounderMock() *mock.RounderMock {
	return &mock.RounderMock{
		RoundIndex: 0,
		TimeStampCalled: func() time.Time {
			return time.Unix(0, 0)
		},
		TimeDurationCalled: func() time.Duration {
			return roundTimeDuration
		},
	}
}

func TestWorker_NewWorkerConsensusServiceNilShouldFail(t *testing.T) {
	t.Parallel()

	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}

	wrk, err := spos.NewWorker(
		nil,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilConsensusService, err)
}

func TestWorker_NewWorkerBlockChainNilShouldFail(t *testing.T) {
	t.Parallel()

	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		nil,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestWorker_NewWorkerBlockProcessorNilShouldFail(t *testing.T) {
	t.Parallel()

	blockchainMock := &mock.BlockChainMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		nil,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestWorker_NewWorkerBootstrapperNilShouldFail(t *testing.T) {
	t.Parallel()

	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		nil,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestWorker_NewWorkerBroadcastMessengerNilShouldFail(t *testing.T) {
	t.Parallel()

	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		nil,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilBroadcastMessenger, err)
}

func TestWorker_NewWorkerConsensusStateNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		nil,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestWorker_NewWorkerForkDetectorNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		nil,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilForkDetector, err)
}

func TestWorker_NewWorkerKeyGeneratorNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		nil,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilKeyGenerator, err)
}

func TestWorker_NewWorkerMarshalizerNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		nil,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestWorker_NewWorkerRounderNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		nil,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestWorker_NewWorkerShardCoordinatorNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		nil,
		singleSignerMock,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestWorker_NewWorkerSingleSignerNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		nil,
		syncTimerMock)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilSingleSigner, err)
}

func TestWorker_NewWorkerSyncTimerNilShouldFail(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		nil)

	assert.Nil(t, wrk)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestWorker_NewWorkerShouldWork(t *testing.T) {
	t.Parallel()
	blockchainMock := &mock.BlockChainMock{}
	blockProcessor := &mock.BlockProcessorMock{}
	bootstrapperMock := &mock.BootstrapperMock{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	consensusState := initConsensusState()
	forkDetectorMock := &mock.ForkDetectorMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	bnService, _ := bn.NewConsensusService()

	wrk, err := spos.NewWorker(
		bnService,
		blockchainMock,
		blockProcessor,
		bootstrapperMock,
		broadcastMessengerMock,
		consensusState,
		forkDetectorMock,
		keyGeneratorMock,
		marshalizerMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
		syncTimerMock)

	assert.NotNil(t, wrk)
	assert.Nil(t, err)
}

func TestWorker_ReceivedSyncStateShouldNotSendOnChannelWhenInputIsFalse(t *testing.T) {
	t.Parallel()
	wrk := initWorker()
	wrk.ReceivedSyncState(false)
	rcv := false
	select {
	case rcv = <-wrk.ConsensusStateChangedChannel():
	case <-time.After(100 * time.Millisecond):
	}

	assert.False(t, rcv)
}

func TestWorker_ReceivedSyncStateShouldNotSendOnChannelWhenChannelIsBusy(t *testing.T) {
	t.Parallel()
	wrk := initWorker()
	wrk.ConsensusStateChangedChannel() <- false
	wrk.ReceivedSyncState(true)
	rcv := false
	select {
	case rcv = <-wrk.ConsensusStateChangedChannel():
	case <-time.After(100 * time.Millisecond):
	}

	assert.False(t, rcv)
}

func TestWorker_ReceivedSyncStateShouldSendOnChannel(t *testing.T) {
	t.Parallel()
	wrk := initWorker()
	wrk.ReceivedSyncState(true)
	rcv := false
	select {
	case rcv = <-wrk.ConsensusStateChangedChannel():
	case <-time.After(100 * time.Millisecond):
	}

	assert.True(t, rcv)
}

func TestWorker_InitReceivedMessagesShouldInitMap(t *testing.T) {
	t.Parallel()
	wrk := initWorker()
	wrk.NilReceivedMessages()
	wrk.InitReceivedMessages()

	assert.NotNil(t, wrk.ReceivedMessages()[bn.MtBlockBody])
}

func TestWorker_AddReceivedMessageCallShouldWork(t *testing.T) {
	t.Parallel()
	wrk := initWorker()
	receivedMessageCall := func(*consensus.Message) bool {
		return true
	}
	wrk.AddReceivedMessageCall(bn.MtBlockBody, receivedMessageCall)
	receivedMessageCalls := wrk.ReceivedMessagesCalls()

	assert.Equal(t, 1, len(receivedMessageCalls))
	assert.NotNil(t, receivedMessageCalls[bn.MtBlockBody])
	assert.True(t, receivedMessageCalls[bn.MtBlockBody](nil))
}

func TestWorker_RemoveAllReceivedMessageCallsShouldWork(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	receivedMessageCall := func(*consensus.Message) bool {
		return true
	}
	wrk.AddReceivedMessageCall(bn.MtBlockBody, receivedMessageCall)
	receivedMessageCalls := wrk.ReceivedMessagesCalls()

	assert.Equal(t, 1, len(receivedMessageCalls))
	assert.NotNil(t, receivedMessageCalls[bn.MtBlockBody])
	assert.True(t, receivedMessageCalls[bn.MtBlockBody](nil))

	wrk.RemoveAllReceivedMessagesCalls()
	receivedMessageCalls = wrk.ReceivedMessagesCalls()

	assert.Equal(t, 0, len(receivedMessageCalls))
	assert.Nil(t, receivedMessageCalls[bn.MtBlockBody])
}

func TestWorker_ProcessReceivedMessageTxBlockBodyShouldRetNil(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	time.Sleep(time.Second)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)

	assert.Nil(t, err)
}

func TestWorker_ProcessReceivedMessageHeaderShouldRetNil(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())
	message, _ := mock.MarshalizerMock{}.Marshal(hdr)
	message, _ = mock.MarshalizerMock{}.Marshal(hdr)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtUnknown),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	time.Sleep(time.Second)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)

	assert.Nil(t, err)
}

func TestWorker_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	err := wrk.ProcessReceivedMessage(nil, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, spos.ErrNilMessage, err)
}

func TestWorker_ProcessReceivedMessageNilMessageDataFieldShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{}, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, spos.ErrNilDataToProcess, err)
}

func TestWorker_ProcessReceivedMessageNodeNotInEligibleListShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte("X"),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, spos.ErrSenderNotOk, err)
}

func TestWorker_ProcessReceivedMessageMessageIsForPastRoundShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		-1,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, spos.ErrMessageForPastRound, err)
}

func TestWorker_ProcessReceivedMessageInvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		nil,
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, spos.ErrInvalidSignature, err)
}

func TestWorker_ProcessReceivedMessageReceivedMessageIsFromSelfShouldRetNilAndNotProcess(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Nil(t, err)
}

func TestWorker_ProcessReceivedMessageWhenRoundIsCanceledShouldRetNilAndNotProcess(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	wrk.ConsensusState().RoundCanceled = true
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Nil(t, err)
}

func TestWorker_ProcessReceivedMessageOkValsShouldWork(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff}, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 1, len(wrk.ReceivedMessages()[bn.MtBlockHeader]))
	assert.Nil(t, err)
}

func TestWorker_CheckSelfStateShouldErrMessageFromItself(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		nil,
		0,
		0,
		0,
	)
	err := wrk.CheckSelfState(cnsMsg)
	assert.Equal(t, spos.ErrMessageFromItself, err)
}

func TestWorker_CheckSelfStateShouldErrRoundCanceled(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	wrk.ConsensusState().RoundCanceled = true
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		nil,
		0,
		0,
		0,
	)
	err := wrk.CheckSelfState(cnsMsg)
	assert.Equal(t, spos.ErrRoundCanceled, err)
}

func TestWorker_CheckSelfStateShouldNotErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		nil,
		0,
		0,
		0,
	)
	err := wrk.CheckSelfState(cnsMsg)
	assert.Nil(t, err)
}

func TestWorker_CheckSignatureShouldReturnErrNilConsensusData(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	err := wrk.CheckSignature(nil)

	assert.Equal(t, spos.ErrNilConsensusData, err)
}

func TestWorker_CheckSignatureShouldReturnErrNilPublicKey(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		nil,
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	err := wrk.CheckSignature(cnsMsg)

	assert.Equal(t, spos.ErrNilPublicKey, err)
}

func TestWorker_CheckSignatureShouldReturnErrNilSignature(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		nil,
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	err := wrk.CheckSignature(cnsMsg)

	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestWorker_CheckSignatureShouldReturnPublicKeyFromByteArrayErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	keyGeneratorMock, _, _ := mock.InitKeys()
	err := errors.New("error public key from byte array")
	keyGeneratorMock.PublicKeyFromByteArrayMock = func(b []byte) (crypto.PublicKey, error) {
		return nil, err
	}
	wrk.SetKeyGenerator(keyGeneratorMock)
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	err2 := wrk.CheckSignature(cnsMsg)

	assert.Equal(t, err, err2)
}

func TestWorker_CheckSignatureShouldReturnMarshalizerErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	marshalizerMock := mock.MarshalizerMock{}
	marshalizerMock.Fail = true
	wrk.SetMarshalizer(marshalizerMock)
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	err := wrk.CheckSignature(cnsMsg)

	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestWorker_CheckSignatureShouldReturnNilErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	err := wrk.CheckSignature(cnsMsg)

	assert.Nil(t, err)
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenConsensusDataIsNil(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, nil)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.Nil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenMessageIsForOtherRound(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		-1,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteBlockBodyMessagesShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteBlockHeaderMessagesShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteCommitmentHashMessagesShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteBitmapMessagesShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteCommitmentMessagesShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitment),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteSignatureMessagesShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_ExecuteMessagesShouldExecute(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ConsensusState().SetStatus(bn.SrStartRound, spos.SsFinished)
	wrk.ExecuteMessage(cnsDataList)

	assert.Nil(t, wrk.ReceivedMessages()[msgType][0])
}

func TestWorker_CheckChannelsShouldWork(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	wrk.SetReceivedMessagesCalls(bn.MtBlockHeader, func(cnsMsg *consensus.Message) bool {
		_ = wrk.ConsensusState().SetJobDone(wrk.ConsensusState().ConsensusGroup()[0], bn.SrBlock, true)
		return true
	})
	rnd := wrk.Rounder()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))
	cnsGroup := wrk.ConsensusState().ConsensusGroup()
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())
	message, _ := mock.MarshalizerMock{}.Marshal(hdr)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		message,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		1,
	)
	wrk.ExecuteMessageChannel() <- cnsMsg
	time.Sleep(1000 * time.Millisecond)
	isBlockJobDone, err := wrk.ConsensusState().JobDone(cnsGroup[0], bn.SrBlock)

	assert.Nil(t, err)
	assert.True(t, isBlockJobDone)
}

func TestWorker_ExtendShouldReturnWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := false
	bootstrapperMock := &mock.BootstrapperMock{
		ShouldSyncCalled: func() bool {
			return true
		},
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
			executed = true
			return nil, nil, errors.New("error")
		},
	}
	wrk.SetBootstrapper(bootstrapperMock)
	wrk.ConsensusState().RoundCanceled = true
	wrk.Extend(0)

	assert.False(t, executed)
}

func TestWorker_ExtendShouldReturnWhenShouldSync(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := false
	bootstrapperMock := &mock.BootstrapperMock{
		ShouldSyncCalled: func() bool {
			return true
		},
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
			executed = true
			return nil, nil, errors.New("error")
		},
	}
	wrk.SetBootstrapper(bootstrapperMock)
	wrk.Extend(0)

	assert.False(t, executed)
}

func TestWorker_ExtendShouldReturnWhenCreateEmptyBlockFail(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := false
	bmm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			executed = true
			return nil
		},
	}
	wrk.SetBroadcastMessenger(bmm)
	bootstrapperMock := &mock.BootstrapperMock{
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
			return nil, nil, errors.New("error")
		}}
	wrk.SetBootstrapper(bootstrapperMock)
	wrk.Extend(0)

	assert.False(t, executed)
}

func TestWorker_ExtendShouldWorkAfterAWhile(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := int32(0)
	blockProcessor := &mock.BlockProcessorMock{
		RevertAccountStateCalled: func() {
			atomic.AddInt32(&executed, 1)
		},
	}
	wrk.SetBlockProcessor(blockProcessor)
	wrk.ConsensusState().SetProcessingBlock(true)
	n := 10
	go func() {
		for n > 0 {
			time.Sleep(100 * time.Millisecond)
			n--
		}
		wrk.ConsensusState().SetProcessingBlock(false)
	}()
	wrk.Extend(0)

	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
	assert.Equal(t, 0, n)
}

func TestWorker_ExtendShouldWork(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := int32(0)
	blockProcessor := &mock.BlockProcessorMock{
		RevertAccountStateCalled: func() {
			atomic.AddInt32(&executed, 1)
		},
	}
	wrk.SetBlockProcessor(blockProcessor)
	wrk.Extend(0)
	time.Sleep(1000 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
}

func TestWorker_ExecuteStoredMessagesShouldWork(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	wrk.InitReceivedMessages()
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := consensus.MessageType(cnsMsg.MsgType)
	cnsDataList := wrk.ReceivedMessages()[msgType]
	cnsDataList = append(cnsDataList, cnsMsg)
	wrk.SetReceivedMessages(msgType, cnsDataList)
	wrk.ConsensusState().SetStatus(bn.SrStartRound, spos.SsFinished)

	rcvMsg := wrk.ReceivedMessages()
	assert.Equal(t, 1, len(rcvMsg[msgType]))

	wrk.ExecuteStoredMessages()

	rcvMsg = wrk.ReceivedMessages()
	assert.Equal(t, 0, len(rcvMsg[msgType]))
}
