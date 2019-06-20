package bn_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func initSubroundSignatureWithContainer(container *mock.ConsensusCoreMock) bn.SubroundSignature {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)

	srSignature, _ := bn.NewSubroundSignature(
		sr,
		extend,
	)

	return srSignature
}

func initSubroundSignature() bn.SubroundSignature {
	container := mock.InitConsensusCore()
	return initSubroundSignatureWithContainer(container)
}

func TestSubroundSignature_NewSubroundSignatureNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srSignature, err := bn.NewSubroundSignature(
		nil,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundSignature_NewSubroundSignatureNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)

	sr.ConsensusState = nil
	srSignature, err := bn.NewSubroundSignature(
		sr,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundSignature_NewSubroundSignatureNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetHasher(nil)
	srSignature, err := bn.NewSubroundSignature(
		sr,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundSignature_NewSubroundSignatureNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetMultiSigner(nil)
	srSignature, err := bn.NewSubroundSignature(
		sr,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubroundSignature_NewSubroundSignatureNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetRounder(nil)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundSignature_NewSubroundSignatureNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)
	container.SetSyncTimer(nil)
	srSignature, err := bn.NewSubroundSignature(
		sr,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundSignature_NewSubroundSignatureShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bn.SrCommitment),
		int(bn.SrSignature),
		int(bn.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
	)

	srSignature, err := bn.NewSubroundSignature(
		sr,
		extend,
	)

	assert.NotNil(t, srSignature)
	assert.Nil(t, err)
}

func TestSubroundSignature_DoSignatureJob(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container)

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	r := sr.DoSignatureJob()
	assert.False(t, r)

	sr.SetStatus(bn.SrCommitment, spos.SsFinished)
	sr.SetStatus(bn.SrSignature, spos.SsFinished)

	r = sr.DoSignatureJob()
	assert.False(t, r)

	sr.SetStatus(bn.SrSignature, spos.SsNotFinished)
	sr.SetJobDone(sr.SelfPubKey(), bn.SrSignature, true)

	r = sr.DoSignatureJob()
	assert.False(t, r)

	sr.SetJobDone(sr.SelfPubKey(), bn.SrSignature, false)

	r = sr.DoSignatureJob()
	assert.False(t, r)

	sr.SetJobDone(sr.SelfPubKey(), bn.SrBitmap, true)
	sr.Data = nil

	r = sr.DoSignatureJob()
	assert.False(t, r)

	dta := []byte("X")
	sr.Data = dta
	sr.Header = &block.Header{}

	multiSignerMock := mock.InitMultiSignerMock()

	multiSignerMock.CommitmentMock = func(uint16) ([]byte, error) {
		return dta, nil
	}

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return mock.HasherMock{}.Compute(string(dta)), nil
	}

	container.SetMultiSigner(multiSignerMock)

	sr.SetJobDone(sr.SelfPubKey(), bn.SrCommitment, true)
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

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container)

	_ = sr.MultiSigner().Reset(nil, 0)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrCommitment, true)

	multiSignerMock := mock.InitMultiSignerMock()
	multiSignerMock.CommitmentMock = func(u uint16) ([]byte, error) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	container.SetMultiSigner(multiSignerMock)

	err := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldErrOnCommitmentHash(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container)

	multiSignerMock := mock.InitMultiSignerMock()

	err := errors.New("error commitment hash")
	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return nil, err
	}

	container.SetMultiSigner(multiSignerMock)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrCommitment, true)

	err2 := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, err, err2)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldErrCommitmentHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container)

	multiSignerMock := mock.InitMultiSignerMock()

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	container.SetMultiSigner(multiSignerMock)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrCommitment, true)

	err := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, spos.ErrCommitmentHashDoesNotMatch, err)
}

func TestSubroundSignature_CheckCommitmentsValidityShouldReturnNil(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container)

	multiSignerMock := mock.InitMultiSignerMock()

	multiSignerMock.CommitmentMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return mock.HasherMock{}.Compute(string([]byte("X"))), nil
	}

	container.SetMultiSigner(multiSignerMock)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrCommitment, true)

	err := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}

func TestSubroundSignature_ReceivedSignature(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	commitment := []byte("commitment")
	cnsMsg := consensus.NewConsensusMessage(
		sr.Data,
		commitment,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.Data = nil
	r := sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("X")
	sr.SetStatus(bn.SrSignature, spos.SsFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrBitmap, true)
	sr.SetStatus(bn.SrSignature, spos.SsFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bn.SrSignature, spos.SsNotFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.True(t, r)
	isSignJobDone, _ := sr.JobDone(sr.ConsensusGroup()[0], bn.SrSignature)
	assert.True(t, isSignJobDone)
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	sr.RoundCanceled = true
	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	sr.SetStatus(bn.SrSignature, spos.SsFinished)
	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenSignaturesCollectedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	for i := 0; i < sr.Threshold(bn.SrBitmap); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, true)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrSignature, true)
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

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBlock, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitment, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("A", bn.SrBitmap, true)
	sr.SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := sr.JobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("B", bn.SrSignature, true)
	isJobDone, _ = sr.JobDone("B", bn.SrSignature)
	assert.True(t, isJobDone)

	ok = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("C", bn.SrSignature, true)
	ok = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("A", bn.SrSignature, true)
	ok = sr.SignaturesCollected(2)
	assert.True(t, ok)
}

func TestSubroundSignature_ReceivedSignatureReturnFalseWhenConsensusDataIsNotEqual(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	cnsMsg := consensus.NewConsensusMessage(
		append(sr.Data, []byte("X")...),
		[]byte("commitment"),
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrBitmap, true)

	assert.False(t, sr.ReceivedSignature(cnsMsg))
}
