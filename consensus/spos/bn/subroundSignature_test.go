package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func initSubroundSignature() bn.SubroundSignature {
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, _ := bn.NewSubroundSignature(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	return srSignature
}

func TestSubroundSignature_NewSubroundSignatureNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	srSignature, err := bn.NewSubroundSignature(
		nil,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, err, spos.ErrNilSubround)
}

func TestSubroundSignature_NewSubroundSignatureNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		nil,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestSubroundSignature_NewSubroundSignatureNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		consensusState,
		nil,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, err, spos.ErrNilHasher)
}

func TestSubroundSignature_NewSubroundSignatureNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		consensusState,
		hasherMock,
		nil,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestSubroundSignature_NewSubroundSignatureNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		nil,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestSubroundSignature_NewSubroundSignatureNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		nil,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestSubroundSignature_NewSubroundSignatureNilSendConsensusMessageFunctionShouldFail(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		nil,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, err, spos.ErrNilSendConsensusMessageFunction)
}

func TestSubroundSignature_NewSubroundSignatureShouldWork(t *testing.T) {
	t.Parallel()

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		ch,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		consensusState,
		hasherMock,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.NotNil(t, srSignature)
	assert.Nil(t, err)
}

func TestSubroundSignature_DoSignatureJob(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	sr.ConsensusState().SetSelfPubKey(sr.ConsensusState().ConsensusGroup()[0])

	r := sr.DoSignatureJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	r = sr.DoSignatureJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsNotFinished)
	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrSignature, true)

	r = sr.DoSignatureJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrSignature, false)

	r = sr.DoSignatureJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBitmap, true)
	sr.ConsensusState().Data = nil

	r = sr.DoSignatureJob()
	assert.False(t, r)

	dta := []byte("X")
	sr.ConsensusState().Data = dta
	sr.ConsensusState().Header = &block.Header{}

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CommitmentMock = func(uint16) ([]byte, error) {
		return dta, nil
	}

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return mock.HasherMock{}.Compute(string(dta)), nil
	}

	sr.SetMultiSigner(multiSignerMock)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrCommitment, true)
	r = sr.DoSignatureJob()
	assert.True(t, r)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldErrNilCommitmet(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	err := sr.CheckCommitmentsValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilCommitment, err)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldErrInvalidIndex(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	_ = sr.MultiSigner().Reset(nil, 0)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitment, true)

	multiSignerMock := initMultiSignerMock()
	multiSignerMock.CommitmentMock = func(u uint16) ([]byte, error) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	sr.SetMultiSigner(multiSignerMock)

	err := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldErrOnCommitmentHash(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	multiSignerMock := initMultiSignerMock()

	err := errors.New("error commitment hash")
	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return nil, err
	}

	sr.SetMultiSigner(multiSignerMock)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitment, true)

	err2 := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, err, err2)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldErrCommitmentHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	sr.SetMultiSigner(multiSignerMock)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitment, true)

	err := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, spos.ErrCommitmentHashDoesNotMatch, err)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldReturnNil(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.CommitmentMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return mock.HasherMock{}.Compute(string([]byte("X"))), nil
	}

	sr.SetMultiSigner(multiSignerMock)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitment, true)

	err := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}

func TestSubroundSignature_ReceivedSignature(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	cnsMsg := spos.NewConsensusMessage(
		sr.ConsensusState().Data,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().Data = nil
	r := sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().Data = []byte("X")
	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrBitmap, true)
	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsNotFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.True(t, r)
	isSignJobDone, _ := sr.ConsensusState().JobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrSignature)
	assert.True(t, isSignJobDone)
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	sr.ConsensusState().RoundCanceled = true
	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)
	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenSignaturesCollectedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, true)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrSignature, true)
	}

	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenSignaturesCollectedReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_SignaturesCollected(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitment, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrBitmap, true)
	sr.ConsensusState().SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := sr.ConsensusState().JobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrSignature, true)
	isJobDone, _ = sr.ConsensusState().JobDone("B", bn.SrSignature)
	assert.True(t, isJobDone)

	ok = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("C", bn.SrSignature, true)
	ok = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrSignature, true)
	ok = sr.SignaturesCollected(2)
	assert.True(t, ok)
}
