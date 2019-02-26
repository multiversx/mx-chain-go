package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func initSubroundCommitment() bn.SubroundCommitment {
	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int64(55*roundTimeDuration/100),
		int64(70*roundTimeDuration/100),
		"(COMMITMENT)",
		ch,
	)

	srCommitment, _ := bn.NewSubroundCommitment(
		sr,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	return srCommitment
}

func TestSubroundCommitment_NewSubroundCommitmentNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	srCommitment, err := bn.NewSubroundCommitment(
		nil,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitment)
	assert.Equal(t, err, spos.ErrNilSubround)
}

func TestSubroundCommitment_NewSubroundCommitmentNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int64(55*roundTimeDuration/100),
		int64(70*roundTimeDuration/100),
		"(COMMITMENT)",
		ch,
	)

	srCommitment, err := bn.NewSubroundCommitment(
		sr,
		nil,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitment)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestSubroundCommitment_NewSubroundCommitmentNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int64(55*roundTimeDuration/100),
		int64(70*roundTimeDuration/100),
		"(COMMITMENT)",
		ch,
	)

	srCommitment, err := bn.NewSubroundCommitment(
		sr,
		consensusState,
		nil,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitment)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestSubroundCommitment_NewSubroundCommitmentNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int64(55*roundTimeDuration/100),
		int64(70*roundTimeDuration/100),
		"(COMMITMENT)",
		ch,
	)

	srCommitment, err := bn.NewSubroundCommitment(
		sr,
		consensusState,
		multiSignerMock,
		nil,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitment)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestSubroundCommitment_NewSubroundCommitmentNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int64(55*roundTimeDuration/100),
		int64(70*roundTimeDuration/100),
		"(COMMITMENT)",
		ch,
	)

	srCommitment, err := bn.NewSubroundCommitment(
		sr,
		consensusState,
		multiSignerMock,
		rounderMock,
		nil,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srCommitment)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestSubroundCommitment_NewSubroundCommitmentNilSendConsensusMessageFunctionShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int64(55*roundTimeDuration/100),
		int64(70*roundTimeDuration/100),
		"(COMMITMENT)",
		ch,
	)

	srCommitment, err := bn.NewSubroundCommitment(
		sr,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		nil,
		extend,
	)

	assert.Nil(t, srCommitment)
	assert.Equal(t, err, spos.ErrNilSendConsensusMessageFunction)
}

func TestSubroundCommitment_NewSubroundCommitmentShouldWork(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrBitmap),
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int64(55*roundTimeDuration/100),
		int64(70*roundTimeDuration/100),
		"(COMMITMENT)",
		ch,
	)

	srCommitment, err := bn.NewSubroundCommitment(
		sr,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.NotNil(t, srCommitment)
	assert.Nil(t, err)
}

func TestSubroundCommitment_DoCommitmentJob(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitment()

	r := sr.DoCommitmentJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)

	r = sr.DoCommitmentJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsNotFinished)
	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrCommitment, true)

	r = sr.DoCommitmentJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrCommitment, false)

	r = sr.DoCommitmentJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBitmap, true)
	sr.ConsensusState().Data = nil

	r = sr.DoCommitmentJob()
	assert.False(t, r)

	dta := []byte("X")
	sr.ConsensusState().Data = dta

	r = sr.DoCommitmentJob()
	assert.True(t, r)
}

func TestSubroundCommitment_ReceivedCommitment(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitment()

	commitment := []byte("commitment")

	cnsMsg := spos.NewConsensusMessage(
		sr.ConsensusState().Data,
		commitment,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitment),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0)

	sr.ConsensusState().Data = nil
	r := sr.ReceivedCommitment(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().Data = []byte("X")
	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)

	r = sr.ReceivedCommitment(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrBitmap, true)
	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)

	r = sr.ReceivedCommitment(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsNotFinished)

	r = sr.ReceivedCommitment(cnsMsg)
	assert.True(t, r)
	isCommJobDone, _ := sr.ConsensusState().JobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitment)
	assert.True(t, isCommJobDone)
}

func TestSubroundCommitment_DoCommitmentConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitment()
	sr.ConsensusState().RoundCanceled = true
	assert.False(t, sr.DoCommitmentConsensusCheck())
}

func TestSubroundCommitment_DoCommitmentConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitment()
	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	assert.True(t, sr.DoCommitmentConsensusCheck())
}

func TestSubroundCommitment_DoCommitmentConsensusCheckShouldReturnTrueWhenCommitmentsCollectedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitment()

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, true)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitment, true)
	}

	assert.True(t, sr.DoCommitmentConsensusCheck())
}

func TestSubroundCommitment_DoCommitmentConsensusCheckShouldReturnFalseWhenCommitmentsCollectedReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitment()
	assert.False(t, sr.DoCommitmentConsensusCheck())
}

func TestSubroundCommitment_CommitmentsCollected(t *testing.T) {
	t.Parallel()

	sr := *initSubroundCommitment()

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitment, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.CommitmentsCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrBitmap, true)
	sr.ConsensusState().SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := sr.ConsensusState().JobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.CommitmentsCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrCommitment, true)
	isJobDone, _ = sr.ConsensusState().JobDone("B", bn.SrCommitment)
	assert.True(t, isJobDone)

	ok = sr.CommitmentsCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("C", bn.SrCommitment, true)
	ok = sr.CommitmentsCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrCommitment, true)
	ok = sr.CommitmentsCollected(2)
	assert.True(t, ok)
}
