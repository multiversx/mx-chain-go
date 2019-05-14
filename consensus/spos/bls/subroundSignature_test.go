package bls_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bls"
	"github.com/stretchr/testify/assert"
)

func initSubroundSignatureWithContainer(container *mock.ConsensusCoreMock) bls.SubroundSignature {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)

	srSignature, _ := bls.NewSubroundSignature(
		sr,
		sendConsensusMessage,
		extend,
	)

	return srSignature
}

func initSubroundSignature() bls.SubroundSignature {
	container := mock.InitConsensusCore()
	return initSubroundSignatureWithContainer(container)
}

func TestSubroundSignature_NewSubroundSignatureNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srSignature, err := bls.NewSubroundSignature(
		nil,
		sendConsensusMessage,
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
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)

	sr.ConsensusState = nil
	srSignature, err := bls.NewSubroundSignature(
		sr,
		sendConsensusMessage,
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
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)
	container.SetHasher(nil)
	srSignature, err := bls.NewSubroundSignature(
		sr,
		sendConsensusMessage,
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
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)
	container.SetMultiSigner(nil)
	srSignature, err := bls.NewSubroundSignature(
		sr,
		sendConsensusMessage,
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
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)
	container.SetRounder(nil)

	srSignature, err := bls.NewSubroundSignature(
		sr,
		sendConsensusMessage,
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
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)
	container.SetSyncTimer(nil)
	srSignature, err := bls.NewSubroundSignature(
		sr,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundSignature_NewSubroundSignatureNilSendConsensusMessageFunctionShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)

	srSignature, err := bls.NewSubroundSignature(
		sr,
		nil,
		extend,
	)

	assert.Nil(t, srSignature)
	assert.Equal(t, spos.ErrNilSendConsensusMessageFunction, err)
}

func TestSubroundSignature_NewSubroundSignatureShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		int(bls.SrBlock),
		int(bls.SrSignature),
		int(bls.SrEndRound),
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		container,
	)

	srSignature, err := bls.NewSubroundSignature(
		sr,
		sendConsensusMessage,
		extend,
	)

	assert.NotNil(t, srSignature)
	assert.Nil(t, err)
}

//
//func TestSubroundSignature_DoSignatureJob(t *testing.T) {
//	t.Parallel()
//
//	container := mock.InitConsensusCore()
//	sr := *initSubroundSignatureWithContainer(container)
//
//	sr.SetSelfPubKey(sr.ConsensusGroup()[1])
//
//	//sr.SetStatus(bls.SrBlock, spos.SsNotFinished)
//	//sr.SetStatus(bls.SrSignature, spos.SsNotFinished)
//
//	r := sr.DoSignatureJob()
//	assert.False(t, r)
//
//	sr.SetStatus(bls.SrBlock, spos.SsFinished)
//	sr.SetStatus(bls.SrSignature, spos.SsFinished)
//
//	r = sr.DoSignatureJob()
//	assert.False(t, r)
//
//	sr.SetStatus(bls.SrSignature, spos.SsNotFinished)
//	sr.SetJobDone(sr.SelfPubKey(), bls.SrSignature, true)
//
//	r = sr.DoSignatureJob()
//	assert.False(t, r)
//
//	sr.SetJobDone(sr.SelfPubKey(), bls.SrSignature, false)
//
//	r = sr.DoSignatureJob()
//	assert.False(t, r)
//
//	//sr.SetJobDone(sr.SelfPubKey(), bls.SrBitmap, true)
//	sr.Data = nil
//
//	r = sr.DoSignatureJob()
//	assert.False(t, r)
//
//	dta := []byte("X")
//	sr.Data = dta
//	sr.Header = &block.Header{}
//
//	multiSignerMock := mock.InitMultiSignerMock()
//
//	multiSignerMock.CommitmentMock = func(uint16) ([]byte, error) {
//		return dta, nil
//	}
//
//	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
//		return mock.HasherMock{}.Compute(string(dta)), nil
//	}
//
//	container.SetMultiSigner(multiSignerMock)
//
//	sr.SetJobDone(sr.SelfPubKey(), bls.SrBlock, true)
//	r = sr.DoSignatureJob()
//	assert.True(t, r)
//}
//
//func TestSubroundSignature_SignaturesCollected(t *testing.T) {
//	t.Parallel()
//
//	sr := *initSubroundSignature()
//
//	for i := 0; i < len(sr.ConsensusGroup()); i++ {
//		sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrBlock, false)
//		sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, false)
//	}
//
//	ok := sr.SignaturesCollected(2)
//	assert.False(t, ok)
//
//	sr.SetJobDone("A", bls.SrBitmap, true)
//	sr.SetJobDone("C", bls.SrBitmap, true)
//	isJobDone, _ := sr.JobDone("C", bls.SrBitmap)
//	assert.True(t, isJobDone)
//
//	ok = sr.SignaturesCollected(2)
//	assert.False(t, ok)
//
//	sr.SetJobDone("B", bls.SrSignature, true)
//	isJobDone, _ := sr.JobDone("B", bls.SrSignature)
//	assert.True(t, isJobDone)
//
//	ok = sr.SignaturesCollected(2)
//	assert.False(t, ok)
//
//	sr.SetJobDone("C", bls.SrSignature, true)
//	ok = sr.SignaturesCollected(2)
//	assert.False(t, ok)
//
//	sr.SetJobDone("A", bls.SrSignature, true)
//	ok = sr.SignaturesCollected(2)
//	assert.True(t, ok)
//}

func TestSubroundSignature_ReceivedSignature(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	commitment := []byte("commitment")
	cnsMsg := consensus.NewConsensusMessage(
		sr.Data,
		commitment,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bls.MtSignature),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.Data = nil
	r := sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("X")
	sr.SetStatus(bls.SrSignature, spos.SsFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	//sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrBitmap, true)
	sr.SetStatus(bls.SrSignature, spos.SsFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bls.SrSignature, spos.SsNotFinished)

	r = sr.ReceivedSignature(cnsMsg)
	assert.True(t, r)
	isSignJobDone, _ := sr.JobDone(sr.ConsensusGroup()[0], bls.SrSignature)
	assert.True(t, isSignJobDone)
}

func TestSubroundSignature_SignaturesCollected(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrBlock, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, false)
	}

	ok, n := sr.SignaturesCollected(2)
	assert.False(t, ok)
	assert.Equal(t, 0, n)

	ok, n = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("B", bls.SrSignature, true)
	isJobDone, _ := sr.JobDone("B", bls.SrSignature)
	assert.True(t, isJobDone)

	ok, n = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.SetJobDone("C", bls.SrSignature, true)
	ok, n = sr.SignaturesCollected(2)
	assert.True(t, ok)
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
	sr.SetStatus(bls.SrSignature, spos.SsFinished)
	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenSignaturesCollectedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	for i := 0; i < sr.Threshold(bls.SrSignature); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}

	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenSignaturesCollectedReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_ReceivedSignatureReturnFalseWhenConsensusDataIsNotEqual(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	cnsMsg := consensus.NewConsensusMessage(
		append(sr.Data, []byte("X")...),
		[]byte("commitment"),
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bls.MtSignature),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	//sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrBitmap, true)

	assert.False(t, sr.ReceivedSignature(cnsMsg))
}
