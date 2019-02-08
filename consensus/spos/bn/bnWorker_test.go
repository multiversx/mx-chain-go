package bn_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/chronology"
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

func sendMessage(cnsDta *spos.ConsensusData) {
	fmt.Println(cnsDta.Signature)
}

func broadcastMessage(msg []byte) {
	fmt.Println(msg)
}

func sendConsensusMessage(cnsDta *spos.ConsensusData) bool {
	fmt.Println(cnsDta)
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

func initWorker() *bn.Worker {
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusState := initConsensusState()

	keyGeneratorMock, privateKeyMock, _ := initSingleSigning()
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()

	wrk, _ := bn.NewWorker(
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
	)

	wrk.SendMessage = sendMessage
	wrk.BroadcastBlockBody = broadcastMessage
	wrk.BroadcastHeader = broadcastMessage

	return wrk
}

func initConsensusState() *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := createEligibleList(consensusGroupSize)
	indexLeader := 1

	rcns := initRoundConsensus(eligibleList, consensusGroupSize, indexLeader)
	rthr := initRoundThreshold(consensusGroupSize)
	rstatus := initRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	cns.Data = []byte("X")

	return cns
}

func initChronologyHandlerMock() chronology.ChronologyHandler {
	chr := &mock.ChronologyHandlerMock{}

	return chr
}

func initRoundThreshold(consensusGroupSize int) *spos.RoundThreshold {
	PBFTThreshold := consensusGroupSize*2/3 + 1

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, PBFTThreshold)
	rthr.SetThreshold(bn.SrBitmap, PBFTThreshold)
	rthr.SetThreshold(bn.SrCommitment, PBFTThreshold)
	rthr.SetThreshold(bn.SrSignature, PBFTThreshold)

	return rthr
}

func initRoundStatus() *spos.RoundStatus {

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrBlock, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitment, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrSignature, spos.SsNotFinished)

	return rstatus
}

func initRoundConsensus(eligibleList []string, consensusGroupSize int, indexLeader int) *spos.RoundConsensus {
	rcns := spos.NewRoundConsensus(
		eligibleList,
		consensusGroupSize,
		eligibleList[indexLeader])

	rcns.SetConsensusGroup(eligibleList)

	for j := 0; j < len(rcns.ConsensusGroup()); j++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[j], bn.SrSignature, false)
	}

	return rcns
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

	multiSigner.CreateCommitmentMock = func() ([]byte, []byte, error) {
		return []byte("commSecret"), []byte("commitment"), nil
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

func initSingleSigning() (*mock.KeyGeneratorMock, *mock.PrivateKeyMock, *mock.PublicKeyMock) {
	toByteArrayMock := func() ([]byte, error) {
		return []byte("byteArray"), nil
	}

	signMock := func(message []byte) ([]byte, error) {
		return []byte("signature"), nil
	}

	verifyMock := func(data []byte, signature []byte) error {
		return nil
	}

	privKeyMock := &mock.PrivateKeyMock{
		SignMock:        signMock,
		ToByteArrayMock: toByteArrayMock,
	}

	pubKeyMock := &mock.PublicKeyMock{
		ToByteArrayMock: toByteArrayMock,
		VerifyMock:      verifyMock,
	}

	privKeyFromByteArr := func(b []byte) (crypto.PrivateKey, error) {
		return privKeyMock, nil
	}

	pubKeyFromByteArr := func(b []byte) (crypto.PublicKey, error) {
		return pubKeyMock, nil
	}

	keyGenMock := &mock.KeyGeneratorMock{
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
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGeneratorMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()

	wrk, err := bn.NewWorker(
		nil,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilBlootstraper)
}

func TestWorker_NewWorkerConsensusStateNilShouldFail(t *testing.T) {
	bootstraperMock := &mock.BootstraperMock{}
	keyGeneratorMock := &mock.KeyGeneratorMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()

	wrk, err := bn.NewWorker(
		bootstraperMock,
		nil,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestWorker_NewWorkerKeyGeneratorNilShouldFail(t *testing.T) {
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()

	wrk, err := bn.NewWorker(
		bootstraperMock,
		consensusState,
		nil,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilKeyGenerator)
}

func TestWorker_NewWorkerMarshalizerNilShouldFail(t *testing.T) {
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGeneratorMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()

	wrk, err := bn.NewWorker(
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		nil,
		privateKeyMock,
		rounderMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilMarshalizer)
}

func TestWorker_NewWorkerPrivateKeyNilShouldFail(t *testing.T) {
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGeneratorMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()

	wrk, err := bn.NewWorker(
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		nil,
		rounderMock,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilPrivateKey)
}

func TestWorker_NewWorkerRounderNilShouldFail(t *testing.T) {
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGeneratorMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}

	wrk, err := bn.NewWorker(
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		nil,
	)

	assert.Nil(t, wrk)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestWorker_NewWorkerShouldWork(t *testing.T) {
	bootstraperMock := &mock.BootstraperMock{}
	consensusState := initConsensusState()
	keyGeneratorMock := &mock.KeyGeneratorMock{}
	marshalizerMock := mock.MarshalizerMock{}
	privateKeyMock := &mock.PrivateKeyMock{}
	rounderMock := initRounderMock()

	wrk, err := bn.NewWorker(
		bootstraperMock,
		consensusState,
		keyGeneratorMock,
		marshalizerMock,
		privateKeyMock,
		rounderMock,
	)

	assert.NotNil(t, wrk)
	assert.Nil(t, err)
}

func TestWorker_InitReceivedMessagesShouldInitMap(t *testing.T) {
	wrk := initWorker()

	wrk.ReceivedMessages = nil
	wrk.InitReceivedMessages()
	assert.NotNil(t, wrk.ReceivedMessages[bn.MtBlockBody])
}

func TestWorker_ShouldDropConsensusMessageShouldReturnTrueWhenMessageReceivedIsFromThePastRounds(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		-1,
	)

	assert.True(t, wrk.ShouldDropConsensusMessage(cnsDta))
}

func TestWorker_ShouldDropConsensusMessageShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	assert.False(t, wrk.ShouldDropConsensusMessage(cnsDta))
}

func TestWorker_ReceivedMessageTxBlockBody(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)
}

func TestWorker_ReceivedMessageUnknown(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))
	message, _ = mock.MarshalizerMock{}.Marshal(hdr)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtUnknown),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)
}

func TestWorker_ReceivedMessageShouldReturnWhenIsCanceled(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.ConsensusState().RoundCanceled = true
	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenDataReceivedIsInvalid(t *testing.T) {
	wrk := initWorker()

	wrk.ReceivedMessage(string(consensusTopic), nil, nil)
	assert.Equal(t, 0, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenNodeIsNotInTheConsensusGroup(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte("X"),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenShouldDropMessage(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		-1,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenReceivedMessageIsFromSelf(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenSignatureIsInvalid(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		nil,
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldSendReceivedMesageOnChannel(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_CheckSignatureShouldReturnErrNilConsensusData(t *testing.T) {
	wrk := initWorker()

	err := wrk.CheckSignature(nil)
	assert.Equal(t, spos.ErrNilConsensusData, err)
}

func TestWorker_CheckSignatureShouldReturnErrNilPublicKey(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		nil,
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	err := wrk.CheckSignature(cnsDta)

	assert.Equal(t, spos.ErrNilPublicKey, err)
}

func TestWorker_CheckSignatureShouldReturnErrNilSignature(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		nil,
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	err := wrk.CheckSignature(cnsDta)

	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestWorker_CheckSignatureShouldReturnPublicKeyFromByteArrayErr(t *testing.T) {
	wrk := initWorker()

	keyGeneratorMock, _, _ := initSingleSigning()

	err := errors.New("error public key from byte array")
	keyGeneratorMock.PublicKeyFromByteArrayMock = func(b []byte) (crypto.PublicKey, error) {
		return nil, err
	}

	wrk.SetKeyGenerator(keyGeneratorMock)

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	err2 := wrk.CheckSignature(cnsDta)

	assert.Equal(t, err, err2)
}

func TestWorker_CheckSignatureShouldReturnNilErr(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	err := wrk.CheckSignature(cnsDta)

	assert.Nil(t, err)
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenConsensusDataIsNil(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, nil)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.Nil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenShouldDropConsensusMessage(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		-1,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenShouldSync(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteBlockBodyMessagesShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteBlockHeaderMessagesShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteCommitmentHashMessagesShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteBitmapMessagesShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteCommitmentMessagesShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitment),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteSignatureMessagesShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ExecuteMessage(cnsDataList)

	assert.NotNil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteMessagesShouldExecute(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.ConsensusState().SetStatus(bn.SrStartRound, spos.SsFinished)

	wrk.ExecuteMessage(cnsDataList)

	assert.Nil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_CheckChannelsShouldWork(t *testing.T) {
	wrk := initWorker()

	wrk.ReceivedMessagesCalls[bn.MtBlockHeader] = func(cnsDta *spos.ConsensusData) bool {
		wrk.ConsensusState().SetJobDone(wrk.ConsensusState().ConsensusGroup()[0], bn.SrBlock, true)
		return true
	}

	rnd := wrk.Rounder()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	cnsGroup := wrk.ConsensusState().ConsensusGroup()
	rcns := wrk.ConsensusState().RoundConsensus

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		1,
	)

	wrk.ConsensusState().Header = hdr

	wrk.ExecuteMessageChannel <- cnsDta
	time.Sleep(10 * time.Millisecond)

	cnsDta.MsgType = int(bn.MtBlockHeader)
	wrk.ExecuteMessageChannel <- cnsDta
	time.Sleep(10 * time.Millisecond)

	isBlockJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrBlock)

	assert.Nil(t, err)
	assert.True(t, isBlockJobDone)
}

func TestWorker_SendConsensusMessage(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(wrk.Rounder().TimeStamp().Unix())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.ConsensusState().SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(wrk.Rounder().TimeStamp().Unix()),
		0,
	)

	wrk.SendMessage = nil
	r := wrk.SendConsensusMessage(cnsDta)
	assert.False(t, r)

	wrk.SendMessage = sendMessage
	r = wrk.SendConsensusMessage(cnsDta)

	assert.True(t, r)
}

func TestWorker_GetMessageTypeName(t *testing.T) {
	r := bn.GetMessageTypeName(bn.MtBlockBody)
	assert.Equal(t, "(BLOCK_BODY)", r)

	r = bn.GetMessageTypeName(bn.MtBlockHeader)
	assert.Equal(t, "(BLOCK_HEADER)", r)

	r = bn.GetMessageTypeName(bn.MtCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)

	r = bn.GetMessageTypeName(bn.MtBitmap)
	assert.Equal(t, "(BITMAP)", r)

	r = bn.GetMessageTypeName(bn.MtCommitment)
	assert.Equal(t, "(COMMITMENT)", r)

	r = bn.GetMessageTypeName(bn.MtSignature)
	assert.Equal(t, "(SIGNATURE)", r)

	r = bn.GetMessageTypeName(bn.MtUnknown)
	assert.Equal(t, "(UNKNOWN)", r)

	r = bn.GetMessageTypeName(bn.MessageType(-1))
	assert.Equal(t, "Undefined message type", r)
}

func TestWorker_GetSubroundName(t *testing.T) {
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
