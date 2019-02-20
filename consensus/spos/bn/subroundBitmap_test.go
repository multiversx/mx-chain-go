package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func initSubroundBitmap() bn.SubroundBitmap {
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		ch,
	)

	srBitmap, _ := bn.NewSubroundBitmap(
		sr,
		blockProcessorMock,
		consensusState,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	return srBitmap
}

func TestSubroundBitmap_NewSubroundBitmapNilSubroundShouldFail(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	srBitmap, err := bn.NewSubroundBitmap(
		nil,
		blockProcessorMock,
		consensusState,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, err, spos.ErrNilSubround)
}

func TestSubroundBitmap_NewSubroundBitmapNilBlockProcessorShouldFail(t *testing.T) {
	consensusState := initConsensusState()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		ch,
	)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		nil,
		consensusState,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, err, spos.ErrNilBlockProcessor)
}

func TestSubroundBitmap_NewSubroundBitmapNilConsensusStateShouldFail(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		ch,
	)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		blockProcessorMock,
		nil,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestSubroundBitmap_NewSubroundBitmapNilRounderShouldFail(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		ch,
	)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		blockProcessorMock,
		consensusState,
		nil,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestSubroundBitmap_NewSubroundBitmapNilSyncTimerShouldFail(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	rounderMock := initRounderMock()

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		ch,
	)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		blockProcessorMock,
		consensusState,
		rounderMock,
		nil,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestSubroundBitmap_NewSubroundBitmapNilSendConsensusMessageFunctionShouldFail(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		ch,
	)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		blockProcessorMock,
		consensusState,
		rounderMock,
		syncTimerMock,
		nil,
		extend,
	)

	assert.Nil(t, srBitmap)
	assert.Equal(t, err, spos.ErrNilSendConsensusMessageFunction)
}

func TestSubroundBitmap_NewSubroundBitmapShouldWork(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitmentHash),
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int64(40*roundTimeDuration/100),
		int64(55*roundTimeDuration/100),
		"(BITMAP)",
		ch,
	)

	srBitmap, err := bn.NewSubroundBitmap(
		sr,
		blockProcessorMock,
		consensusState,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.NotNil(t, srBitmap)
	assert.Nil(t, err)
}

func TestSubroundBitmap_DoBitmapJob(t *testing.T) {
	sr := *initSubroundBitmap()

	sr.ConsensusState().Header = &block.Header{}

	r := sr.DoBitmapJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)

	r = sr.DoBitmapJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsNotFinished)
	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBitmap, true)

	r = sr.DoBitmapJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBitmap, false)
	sr.ConsensusState().RoundConsensus.SetSelfPubKey(sr.ConsensusState().RoundConsensus.ConsensusGroup()[1])

	r = sr.DoBitmapJob()
	assert.False(t, r)

	sr.ConsensusState().RoundConsensus.SetSelfPubKey(sr.ConsensusState().ConsensusGroup()[0])
	sr.ConsensusState().Data = nil

	r = sr.DoBitmapJob()
	assert.False(t, r)

	dta := []byte("X")
	sr.ConsensusState().Data = dta
	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrCommitmentHash, true)

	r = sr.DoBitmapJob()
	assert.True(t, r)
	isBitmapJobDone, _ := sr.ConsensusState().GetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBitmap)
	assert.True(t, isBitmapJobDone)
}

func TestSubroundBitmap_ReceivedBitmap(t *testing.T) {
	sr := *initSubroundBitmap()

	sr.ConsensusState().Header = &block.Header{}

	commitment := []byte("commitment")

	cnsDta := spos.NewConsensusData(
		sr.ConsensusState().Data,
		commitment,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().Data = nil
	r := sr.ReceivedBitmap(cnsDta)
	assert.False(t, r)

	sr.ConsensusState().Data = []byte("X")
	cnsDta.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[0] + "X")
	r = sr.ReceivedBitmap(cnsDta)
	assert.False(t, r)

	cnsDta.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[0])
	sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	r = sr.ReceivedBitmap(cnsDta)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsNotFinished)

	bitmap := make([]byte, 3)

	cnGroup := sr.ConsensusState().ConsensusGroup()

	selfIndexInConsensusGroup, _ := sr.ConsensusState().IndexSelfConsensusGroup()

	// fill ony few of the signers in bitmap
	for i := 0; i < 5; i++ {
		if i != selfIndexInConsensusGroup {
			bitmap[i/8] |= 1 << uint16(i%8)
		}
	}

	cnsDta.SubRoundData = bitmap

	r = sr.ReceivedBitmap(cnsDta)
	assert.False(t, r)

	//fill the rest except self
	for i := 5; i < len(cnGroup); i++ {
		if i != selfIndexInConsensusGroup {
			bitmap[i/8] |= 1 << uint16(i%8)
		}
	}

	cnsDta.SubRoundData = bitmap
	r = sr.ReceivedBitmap(cnsDta)
	assert.False(t, r)

	//fill self
	sr.ConsensusState().ResetRoundState()
	bitmap[selfIndexInConsensusGroup/8] |= 1 << uint16(selfIndexInConsensusGroup%8)

	cnsDta.SubRoundData = bitmap
	r = sr.ReceivedBitmap(cnsDta)
	assert.True(t, r)
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	sr := *initSubroundBitmap()
	sr.ConsensusState().RoundCanceled = true
	assert.False(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	sr := *initSubroundBitmap()
	sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	assert.True(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnTrueWhenBitmapIsReceivedReturnTrue(t *testing.T) {
	sr := *initSubroundBitmap()

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, true)
	}

	assert.True(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_DoBitmapConsensusCheckShouldReturnFalseWhenBitmapIsReceivedReturnFalse(t *testing.T) {
	sr := *initSubroundBitmap()
	assert.False(t, sr.DoBitmapConsensusCheck())
}

func TestSubroundBitmap_IsBitmapReceived(t *testing.T) {
	sr := *initSubroundBitmap()

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitment, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.IsBitmapReceived(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrBitmap, true)
	isJobDone, _ := sr.ConsensusState().GetJobDone("A", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.IsBitmapReceived(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrBitmap, true)
	ok = sr.IsBitmapReceived(2)
	assert.True(t, ok)

	sr.ConsensusState().SetJobDone("C", bn.SrBitmap, true)
	ok = sr.IsBitmapReceived(2)
	assert.True(t, ok)
}
