package bn_test

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

const roundTimeDuration = time.Duration(100 * time.Millisecond)

func sendMessage(cnsMsg *spos.ConsensusMessage) {
	fmt.Println(cnsMsg.Signature)
}

func sendConsensusMessage(cnsMsg *spos.ConsensusMessage) bool {
	fmt.Println(cnsMsg)
	return true
}

func broadcastBlock(txBlockBody data.BodyHandler, header data.HeaderHandler) error {
	fmt.Println(txBlockBody)
	fmt.Println(header)
	return nil
}

func extend(subroundId int) {
	fmt.Println(subroundId)
}

func initWorker() bn.Worker {
	blockProcessor := &mock.BlockProcessorMock{
		RevertAccountStateCalled: func() {
		},
	}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock, privateKeyMock, _ := initKeys()
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

	wrk, _ := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	wrk.SendMessage = sendMessage
	wrk.BroadcastBlock = broadcastBlock
	return wrk
}

func initConsensusState() *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := createEligibleList(consensusGroupSize)
	indexLeader := 1
	rcns := spos.NewRoundConsensus(
		eligibleList,
		consensusGroupSize,
		eligibleList[indexLeader])

	rcns.SetConsensusGroup(eligibleList)
	rcns.ResetRoundState()

	PBFTThreshold := consensusGroupSize*2/3 + 1

	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, PBFTThreshold)
	rthr.SetThreshold(bn.SrBitmap, PBFTThreshold)
	rthr.SetThreshold(bn.SrCommitment, PBFTThreshold)
	rthr.SetThreshold(bn.SrSignature, PBFTThreshold)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	cns.Data = []byte("X")
	cns.RoundIndex = 0
	return cns
}

func initChronologyHandlerMock() consensus.ChronologyHandler {
	chr := &mock.ChronologyHandlerMock{}
	return chr
}

func initBlockProcessorMock() *mock.BlockProcessorMock {
	blockProcessorMock := &mock.BlockProcessorMock{}
	blockProcessorMock.RemoveBlockInfoFromPoolCalled = func(body data.BodyHandler) error { return nil }
	blockProcessorMock.CreateBlockCalled = func(round int32, haveTime func() bool) (data.BodyHandler, error) {
		emptyBlock := make(block.Body, 0)

		return emptyBlock, nil
	}
	blockProcessorMock.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
		return nil
	}
	blockProcessorMock.RevertAccountStateCalled = func() {}
	blockProcessorMock.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return nil
	}
	blockProcessorMock.GetRootHashCalled = func() []byte {
		return []byte{}
	}
	blockProcessorMock.CreateBlockHeaderCalled = func(body data.BodyHandler) (header data.HeaderHandler, e error) {
		return &block.Header{RootHash: blockProcessorMock.GetRootHashCalled()}, nil
	}
	return blockProcessorMock
}

func initMultiSignerMock() *mock.BelNevMock {
	multiSigner := mock.NewMultiSigner()
	multiSigner.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("commitment")
	}
	multiSigner.VerifySignatureShareMock = func(index uint16, sig []byte, bitmap []byte) error {
		return nil
	}
	multiSigner.VerifyMock = func(bitmap []byte) error {
		return nil
	}
	multiSigner.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return []byte("aggregatedSig"), nil
	}
	multiSigner.AggregateCommitmentsMock = func(bitmap []byte) error {
		return nil
	}
	multiSigner.CreateSignatureShareMock = func(bitmap []byte) ([]byte, error) {
		return []byte("partialSign"), nil
	}
	return multiSigner
}

func initKeys() (*mock.KeyGenMock, *mock.PrivateKeyMock, *mock.PublicKeyMock) {
	toByteArrayMock := func() ([]byte, error) {
		return []byte("byteArray"), nil
	}
	privKeyMock := &mock.PrivateKeyMock{
		ToByteArrayMock: toByteArrayMock,
	}
	pubKeyMock := &mock.PublicKeyMock{
		ToByteArrayMock: toByteArrayMock,
	}
	privKeyFromByteArr := func(b []byte) (crypto.PrivateKey, error) {
		return privKeyMock, nil
	}
	pubKeyFromByteArr := func(b []byte) (crypto.PublicKey, error) {
		return pubKeyMock, nil
	}
	keyGenMock := &mock.KeyGenMock{
		PrivateKeyFromByteArrayMock: privKeyFromByteArr,
		PublicKeyFromByteArrayMock:  pubKeyFromByteArr,
	}
	return keyGenMock, privKeyMock, pubKeyMock
}

func initRounderMock() *mock.RounderMock {
	return &mock.RounderMock{
		RoundIndex:        0,
		RoundTimeStamp:    time.Unix(0, 0),
		RoundTimeDuration: roundTimeDuration,
	}
}

func createEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, string(i+65))
	}
	return eligibleList
}

func TestWorker_NewWorkerBlockprocessorNilShouldFail(t *testing.T) {
	t.Parallel()
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		nil,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilBlockProcessor)
}

func TestWorker_NewWorkerBoostraperNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		nil,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilBlootstraper)
}

func TestWorker_NewWorkerConsensusStateNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		nil,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestWorker_NewWorkerKeyGeneratorNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		nil,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilKeyGenerator)
}

func TestWorker_NewWorkerMarshalizerNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		nil,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilMarshalizer)
}

func TestWorker_NewWorkerPrivateKeyNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		nil,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilPrivateKey)
}

func TestWorker_NewWorkerRounderNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		nil,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestWorker_NewWorkerShardCoordinatorNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		nil,
		singleSignerMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilShardCoordinator)
}

func TestWorker_NewWorkerSingleSignerNilShouldFail(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		nil,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilSingleSigner)
}

func TestWorker_NewWorkerShouldWork(t *testing.T) {
	t.Parallel()
	blockProcessor := &mock.BlockProcessorMock{}
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}
	wrk, err := bn.NewWorker(
		blockProcessor,
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
		shardCoordinatorMock,
		singleSignerMock,
	)

	assert.NotNil(t, wrk)
	assert.Nil(t, err)
}

func TestWorker_ReceivedSyncStateShouldNotSendOnChannelWhenInputIsFalse(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	wrk.ReceivedSyncState(false)
	rcv := false
	select {
	case rcv = <-wrk.ConsensusStateChangedChannels():
	case <-time.After(100 * time.Millisecond):
	}

	assert.False(t, rcv)
}

func TestWorker_ReceivedSyncStateShouldNotSendOnChannelWhenChannelIsBusy(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	wrk.ConsensusStateChangedChannels() <- false
	wrk.ReceivedSyncState(true)
	rcv := false
	select {
	case rcv = <-wrk.ConsensusStateChangedChannels():
	case <-time.After(100 * time.Millisecond):
	}

	assert.False(t, rcv)
}

func TestWorker_ReceivedSyncStateShouldSendOnChannel(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	wrk.ReceivedSyncState(true)
	rcv := false
	select {
	case rcv = <-wrk.ConsensusStateChangedChannels():
	case <-time.After(100 * time.Millisecond):
	}

	assert.True(t, rcv)
}

func TestWorker_InitReceivedMessagesShouldInitMap(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	wrk.NilReceivedMessages()
	wrk.InitReceivedMessages()

	assert.NotNil(t, wrk.ReceivedMessages()[bn.MtBlockBody])
}

func TestWorker_AddReceivedMessageCallShouldWork(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	receivedMessageCall := func(*spos.ConsensusMessage) bool {
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
	receivedMessageCall := func(*spos.ConsensusMessage) bool {
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
	cnsMsg := spos.NewConsensusMessage(
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
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff})

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
	cnsMsg := spos.NewConsensusMessage(
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
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff})

	assert.Nil(t, err)
}

func TestWorker_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	err := wrk.ProcessReceivedMessage(nil)
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrNilMessage, err)
}

func TestWorker_ProcessReceivedMessageNilMessageDataFieldShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{})
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrNilDataToProcess, err)
}

func TestWorker_ProcessReceivedMessageNodeNotInEligibleListShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte("X"),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff})
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrSenderNotOk, err)
}

func TestWorker_ProcessReceivedMessageMessageIsForPastRoundShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		-1,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff})
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrMessageForPastRound, err)
}

func TestWorker_ProcessReceivedMessageReceivedMessageIsFromSelfShouldRetNilAndNotProcess(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff})
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Nil(t, err)
}

func TestWorker_ProcessReceivedMessageInvalidSignatureShouldErr(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		nil,
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff})
	time.Sleep(time.Second)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrInvalidSignature, err)
}

func TestWorker_ProcessReceivedMessageOkValsShouldWork(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	buff, _ := wrk.Marshalizer().Marshal(cnsMsg)
	err := wrk.ProcessReceivedMessage(&mock.P2PMessageMock{DataField: buff})
	time.Sleep(time.Second)

	assert.Equal(t, 1, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
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
	cnsMsg := spos.NewConsensusMessage(
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
	cnsMsg := spos.NewConsensusMessage(
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
	keyGeneratorMock, _, _ := initKeys()
	err := errors.New("error public key from byte array")
	keyGeneratorMock.PublicKeyFromByteArrayMock = func(b []byte) (crypto.PublicKey, error) {
		return nil, err
	}
	wrk.SetKeyGenerator(keyGeneratorMock)
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
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
	cnsMsg := spos.NewConsensusMessage(
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
	cnsMsg := spos.NewConsensusMessage(
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		-1,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitment),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)
	msgType := bn.MessageType(cnsMsg.MsgType)
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
	wrk.SetReceivedMessagesCalls(bn.MtBlockHeader, func(cnsMsg *spos.ConsensusMessage) bool {
		wrk.ConsensusState().SetJobDone(wrk.ConsensusState().ConsensusGroup()[0], bn.SrBlock, true)
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
	cnsMsg := spos.NewConsensusMessage(
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

func TestWorker_SendConsensusMessage(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	marshalizerMock := mock.MarshalizerMock{}
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())
	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	message, _ = mock.MarshalizerMock{}.Marshal(hdr)
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.SendMessage = nil
	r := wrk.SendConsensusMessage(cnsMsg)

	assert.False(t, r)

	wrk.SendMessage = sendMessage
	marshalizerMock.Fail = true
	wrk.SetMarshalizer(marshalizerMock)
	r = wrk.SendConsensusMessage(cnsMsg)

	assert.False(t, r)

	marshalizerMock.Fail = false
	wrk.SetMarshalizer(marshalizerMock)
	r = wrk.SendConsensusMessage(cnsMsg)

	assert.True(t, r)
}

func TestWorker_ExtendShouldReturnWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := false
	bootstraperMock := &mock.BootstraperMock{
		ShouldSyncCalled: func() bool {
			return true
		},
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
			executed = true
			return nil, nil, errors.New("error")
		},
	}
	wrk.SetBootstraper(bootstraperMock)
	wrk.ConsensusState().RoundCanceled = true
	wrk.Extend(0)

	assert.False(t, executed)
}

func TestWorker_ExtendShouldReturnWhenShouldSync(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := false
	bootstraperMock := &mock.BootstraperMock{
		ShouldSyncCalled: func() bool {
			return true
		},
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
			executed = true
			return nil, nil, errors.New("error")
		},
	}
	wrk.SetBootstraper(bootstraperMock)
	wrk.Extend(0)

	assert.False(t, executed)
}

func TestWorker_ExtendShouldReturnWhenCreateEmptyBlockFail(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := false
	wrk.BroadcastBlock = func(data.BodyHandler, data.HeaderHandler) error {
		executed = true
		return nil
	}
	bootstraperMock := &mock.BootstraperMock{
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
			return nil, nil, errors.New("error")
		}}
	wrk.SetBootstraper(bootstraperMock)
	wrk.Extend(0)

	assert.False(t, executed)
}

func TestWorker_ExtendShouldWorkAfterAWhile(t *testing.T) {
	t.Parallel()
	wrk := *initWorker()
	executed := int32(0)
	wrk.BroadcastBlock = func(data.BodyHandler, data.HeaderHandler) error {
		atomic.AddInt32(&executed, 1)
		return nil
	}
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
	wrk.BroadcastBlock = func(data.BodyHandler, data.HeaderHandler) error {
		atomic.AddInt32(&executed, 1)
		return nil
	}
	wrk.Extend(0)
	time.Sleep(1000 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&executed))
}

func TestWorker_GetSubroundName(t *testing.T) {
	t.Parallel()
	r := bn.GetSubroundName(bn.SrStartRound)
	assert.Equal(t, "(START_ROUND)", r)
	r = bn.GetSubroundName(bn.SrBlock)
	assert.Equal(t, "(BLOCK)", r)
	r = bn.GetSubroundName(bn.SrCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)
	r = bn.GetSubroundName(bn.SrBitmap)
	assert.Equal(t, "(BITMAP)", r)
	r = bn.GetSubroundName(bn.SrCommitment)
	assert.Equal(t, "(COMMITMENT)", r)
	r = bn.GetSubroundName(bn.SrSignature)
	assert.Equal(t, "(SIGNATURE)", r)
	r = bn.GetSubroundName(bn.SrEndRound)
	assert.Equal(t, "(END_ROUND)", r)
	r = bn.GetSubroundName(-1)
	assert.Equal(t, "Undefined subround", r)
}
