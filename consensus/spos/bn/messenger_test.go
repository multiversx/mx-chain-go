package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWorker_InitMessageChannelsShouldInitMap(t *testing.T) {
	wrk := initWorker()

	wrk.MessageChannels = nil
	wrk.InitMessageChannels()
	assert.NotNil(t, wrk.MessageChannels[bn.MtBlockBody])
}

func TestWorker_InitReceivedMessagesShouldInitMap(t *testing.T) {
	wrk := initWorker()

	wrk.ReceivedMessages = nil
	wrk.InitReceivedMessages()
	assert.NotNil(t, wrk.ReceivedMessages[bn.MtBlockBody])
}

func TestWorker_CleanReceivedMessagesShouldCleanList(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	wrk.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		-1,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.CleanReceivedMessages()

	assert.Equal(t, 0, len(wrk.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ShouldDropConsensusMessageShouldReturnTrueWhenMessageReceivedIsFromThePastRounds(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		-1,
	)

	assert.True(t, wrk.ShouldDropConsensusMessage(cnsDta))
}

func TestWorker_ShouldDropConsensusMessageShouldReturnTrueWhenMessageIsReceivedAfterEndRound(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	endTime := getEndTime(wrk.SPoS.Chr, bn.SrEndRound)
	wrk.SPoS.Chr.SetClockOffset(time.Duration(endTime))

	assert.True(t, wrk.ShouldDropConsensusMessage(cnsDta))
}

func TestWorker_ShouldDropConsensusMessageShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)
}

func TestWorker_ReceivedMessageUnknown(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))
	message, _ = mock.MarshalizerMock{}.Marshal(hdr)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtUnknown),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.SPoS.Chr.SetSelfSubround(-1)
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
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		nil,
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	time.Sleep(100 * time.Millisecond)

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
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		nil,
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	err := wrk.CheckSignature(cnsDta)

	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestWorker_CheckSignatureShouldReturnPublicKeyFromByteArrayErr(t *testing.T) {
	wrk := initWorker()

	keyGenMock, _, _ := initSingleSigning()

	err := errors.New("error public key from byte array")
	keyGenMock.PublicKeyFromByteArrayMock = func(b []byte) (crypto.PublicKey, error) {
		return nil, err
	}

	wrk.SetKeyGen(keyGenMock)

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitment),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		wrk.SPoS.Chr.RoundTimeStamp(),
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
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := wrk.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	wrk.ReceivedMessages[msgType] = cnsDataList

	wrk.SPoS.SetStatus(bn.SrStartRound, spos.SsFinished)

	wrk.ExecuteMessage(cnsDataList)

	assert.Nil(t, wrk.ReceivedMessages[msgType][0])
}

func TestWorker_CheckChannelTxBlockBody(t *testing.T) {
	wrk := initWorker()

	wrk.Header = nil
	rnd := wrk.SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	cnsGroup := wrk.SPoS.ConsensusGroup()
	rcns := wrk.SPoS.RoundConsensus

	// BLOCK BODY
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.MessageChannels[bn.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isBlockJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrBlock)

	assert.Nil(t, err)
	// Not done since header is missing
	assert.False(t, isBlockJobDone)
}

func TestWorker_CheckChannelBlockHeader(t *testing.T) {
	wrk := initWorker()

	rnd := wrk.SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	cnsGroup := wrk.SPoS.ConsensusGroup()
	rcns := wrk.SPoS.RoundConsensus

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		1,
	)

	wrk.Header = hdr

	wrk.MessageChannels[bn.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	cnsDta.MsgType = int(bn.MtBlockHeader)
	wrk.MessageChannels[bn.MtBlockHeader] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	isBlockJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrBlock)

	assert.Nil(t, err)
	assert.True(t, isBlockJobDone)
}

func TestWorker_CheckChannelsCommitmentHash(t *testing.T) {
	wrk := initWorker()

	rcns := wrk.SPoS.RoundConsensus
	cnsGroup := wrk.SPoS.ConsensusGroup()

	commitmentHash := []byte("commitmentHash")

	// COMMITMENT_HASH
	cnsDta := spos.NewConsensusData(
		wrk.SPoS.Data,
		commitmentHash,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.MessageChannels[bn.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isCommitmentHashJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrCommitmentHash)

	assert.Nil(t, err)
	assert.True(t, isCommitmentHashJobDone)
}

func TestWorker_CheckChannelsBitmap(t *testing.T) {
	wrk := initWorker()

	rcns := wrk.SPoS.RoundConsensus
	cnsGroup := wrk.SPoS.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	// BITMAP
	cnsDta := spos.NewConsensusData(
		wrk.SPoS.Data,
		bitmap,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.Header = &block.Header{}

	wrk.MessageChannels[bn.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < len(cnsGroup); i++ {
		isBitmapJobDone, err := rcns.GetJobDone(cnsGroup[i], bn.SrBitmap)
		assert.Nil(t, err)
		assert.True(t, isBitmapJobDone)
	}
}

func TestWorker_CheckChannelsCommitment(t *testing.T) {
	wrk := initWorker()

	rcns := wrk.SPoS.RoundConsensus
	cnsGroup := wrk.SPoS.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	commHash := mock.HasherMock{}.Compute("commitment")

	// Commitment Hash
	cnsDta := spos.NewConsensusData(
		wrk.SPoS.Data,
		commHash,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.Header = &block.Header{}

	wrk.MessageChannels[bn.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	// Bitmap
	cnsDta.MsgType = int(bn.MtBitmap)
	cnsDta.SubRoundData = bitmap
	wrk.MessageChannels[bn.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	// Commitment
	cnsDta.MsgType = int(bn.MtCommitment)
	cnsDta.SubRoundData = []byte("commitment")
	wrk.MessageChannels[bn.MtCommitment] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	isCommitmentJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrCommitment)

	assert.Nil(t, err)
	assert.True(t, isCommitmentJobDone)
}

func TestWorker_CheckChannelsSignature(t *testing.T) {
	wrk := initWorker()

	rcns := wrk.SPoS.RoundConsensus
	cnsGroup := wrk.SPoS.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	// Bitmap
	cnsDta := spos.NewConsensusData(
		wrk.SPoS.Data,
		bitmap,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.Header = &block.Header{}

	wrk.MessageChannels[bn.MtBitmap] <- cnsDta

	// Signature
	cnsDta.MsgType = int(bn.MtSignature)
	cnsDta.SubRoundData = []byte("signature")
	time.Sleep(10 * time.Millisecond)

	wrk.MessageChannels[bn.MtSignature] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isSigJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrSignature)

	assert.Nil(t, err)
	assert.True(t, isSigJobDone)
}

func TestWorker_SendConsensusMessage(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.SendMessage = nil
	r := wrk.SendConsensusMessage(cnsDta)
	assert.False(t, r)

	wrk.SendMessage = sendMessage
	r = wrk.SendConsensusMessage(cnsDta)

	assert.True(t, r)
}
