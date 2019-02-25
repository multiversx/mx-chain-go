package bn_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type topicName string

const (
	consensusTopic topicName = "cns"
)

const roundTimeDuration = time.Duration(100 * time.Millisecond)

func sendMessage(cnsMsg *spos.ConsensusMessage) {
	fmt.Println(cnsMsg.Signature)
}

func broadcastMessage(msg []byte) {
	fmt.Println(msg)
}

func sendConsensusMessage(cnsMsg *spos.ConsensusMessage) bool {
	fmt.Println(cnsMsg)
	return true
}

func broadcastTxBlockBody(txBlockBody *block.TxBlockBody) error {
	fmt.Println(txBlockBody)
	return nil
}

func broadcastHeader(header *block.Header) error {
	fmt.Println(header)
	return nil
}

func extend(subroundId int) {
	fmt.Println(subroundId)
}

func initWorker() bn.Worker {
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
	wrk.BroadcastTxBlockBody = broadcastMessage
	wrk.BroadcastHeader = broadcastMessage

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

	return cns
}

func initChronologyHandlerMock() consensus.ChronologyHandler {
	chr := &mock.ChronologyHandlerMock{}

	return chr
}

func initBlockProcessorMock() *mock.BlockProcessorMock {
	blockProcessorMock := &mock.BlockProcessorMock{}

	blockProcessorMock.RemoveBlockTxsFromPoolCalled = func(*block.TxBlockBody) error { return nil }
	blockProcessorMock.CreateTxBlockCalled = func(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
		return &block.TxBlockBody{}, nil
	}

	blockProcessorMock.CommitBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
		return nil
	}

	blockProcessorMock.RevertAccountStateCalled = func() {}

	blockProcessorMock.ProcessAndCommitCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return nil
	}

	blockProcessorMock.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return nil
	}

	blockProcessorMock.CreateEmptyBlockBodyCalled = func(shardId uint32, round int32) *block.TxBlockBody {
		return &block.TxBlockBody{}
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

	multiSigner.AggregateCommitmentsMock = func(bitmap []byte) ([]byte, error) {
		return []byte("aggregatedCommitments"), nil
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

func TestWorker_NewWorkerBoostraperNilShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}

	wrk, err := bn.NewWorker(
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

	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGenMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	singleSignerMock := &mock.SingleSignerMock{}

	wrk, err := bn.NewWorker(
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

func TestWorker_ReceivedMessageTxBlockBody(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	blk := &block.TxBlockBody{}
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

	err := wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)
	assert.Nil(t, err)
}

func TestWorker_ReceivedMessageUnknown(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))
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

	err := wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)
	assert.Nil(t, err)
}

func TestWorker_ReceivedMessageShouldReturnWhenIsCanceled(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	blk := &block.TxBlockBody{}
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

	wrk.ConsensusState().RoundCanceled = true
	err := wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrRoundCanceled, err)
}

func TestWorker_ReceivedMessageShouldReturnWhenDataReceivedIsInvalid(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	err := wrk.ReceivedMessage(string(consensusTopic), nil, nil)
	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrInvalidConsensusData, err)
}

func TestWorker_ReceivedMessageShouldReturnWhenNodeIsNotInTheEligibleList(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	blk := &block.TxBlockBody{}
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

	err := wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrSenderNotOk, err)
}

func TestWorker_ReceivedMessageShouldReturnWhenMessageIsForPastRound(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	blk := &block.TxBlockBody{}
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

	err := wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrMessageForPastRound, err)
}

func TestWorker_ReceivedMessageShouldReturnWhenReceivedMessageIsFromSelf(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	blk := &block.TxBlockBody{}
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

	err := wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrMessageFromItself, err)
}

func TestWorker_ReceivedMessageShouldReturnWhenSignatureIsInvalid(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	blk := &block.TxBlockBody{}
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

	err := wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
	assert.Equal(t, bn.ErrInvalidSignature, err)
}

func TestWorker_ReceivedMessageShouldSendReceivedMesageOnChannel(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	blk := &block.TxBlockBody{}
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

	wrk.ReceivedMessage(string(consensusTopic), cnsMsg, nil)

	time.Sleep(1000 * time.Millisecond)

	assert.Equal(t, 1, len(wrk.ReceivedMessages()[bn.MtBlockBody]))
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	blk := &block.TxBlockBody{}
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

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

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

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

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

func TestWorker_BroadcastTxBlockBodyShouldFailWhenBlockBodyNil(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	err := wrk.BroadcastTxBlockBody2(nil)
	assert.Equal(t, spos.ErrNilTxBlockBody, err)
}

func TestWorker_BroadcastTxBlockBodyShouldFailWhenMarshalErr(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	marshalizerMock := mock.MarshalizerMock{}
	marshalizerMock.Fail = true
	wrk.SetMarshalizer(marshalizerMock)

	err := wrk.BroadcastTxBlockBody2(&block.TxBlockBody{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestWorker_BroadcastTxBlockBodyShouldFailWhenBroadcastTxBlockBodyFunctionIsNil(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	wrk.BroadcastTxBlockBody = nil

	err := wrk.BroadcastTxBlockBody2(&block.TxBlockBody{})
	assert.Equal(t, spos.ErrNilOnBroadcastTxBlockBody, err)
}

func TestWorker_BroadcastTxBlockBodyShouldWork(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	err := wrk.BroadcastTxBlockBody2(&block.TxBlockBody{})
	assert.Nil(t, err)
}

func TestWorker_BroadcastHeaderShouldFailWhenHeaderNil(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	err := wrk.BroadcastHeader2(nil)
	assert.Equal(t, spos.ErrNilBlockHeader, err)
}

func TestWorker_BroadcastHeaderShouldFailWhenMarshalErr(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	marshalizerMock := mock.MarshalizerMock{}
	marshalizerMock.Fail = true
	wrk.SetMarshalizer(marshalizerMock)

	err := wrk.BroadcastHeader2(&block.Header{})
	assert.Equal(t, mock.ErrMockMarshalizer, err)
}

func TestWorker_BroadcastHeaderShouldFailWhenBroadcastHeaderFunctionIsNil(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	wrk.BroadcastHeader = nil

	err := wrk.BroadcastHeader2(&block.Header{})
	assert.Equal(t, spos.ErrNilOnBroadcastHeader, err)
}

func TestWorker_BroadcastHeaderShouldWork(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	err := wrk.BroadcastHeader2(&block.Header{})
	assert.Nil(t, err)
}

func TestWorker_ExtendShouldReturnWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	executed := false

	bootstraperMock := &mock.BootstraperMock{
		ShouldSyncCalled: func() bool {
			return true
		},
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (*block.TxBlockBody, *block.Header) {
			executed = true
			return nil, nil
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
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (*block.TxBlockBody, *block.Header) {
			executed = true
			return nil, nil
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

	wrk.BroadcastTxBlockBody = func(msg []byte) {
		executed = true
	}

	wrk.BroadcastHeader = func(msg []byte) {
		executed = true
	}

	bootstraperMock := &mock.BootstraperMock{
		CreateAndCommitEmptyBlockCalled: func(shardForCurrentNode uint32) (*block.TxBlockBody, *block.Header) {
			return nil, nil
		}}

	wrk.SetBootstraper(bootstraperMock)
	wrk.Extend(0)
	assert.False(t, executed)
}

func TestWorker_ExtendShouldReturnWhenBroadcastTxBlockBodyIsNil(t *testing.T) {
	t.Parallel()

	wrk := *initWorker()

	executed := int32(0)

	wrk.BroadcastTxBlockBody = func(msg []byte) {
		atomic.AddInt32(&executed, 1)
	}

	wrk.BroadcastHeader = func(msg []byte) {
		atomic.AddInt32(&executed, 1)
	}

	wrk.Extend(0)
	time.Sleep(1000 * time.Millisecond)
	assert.Equal(t, int32(2), atomic.LoadInt32(&executed))
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
