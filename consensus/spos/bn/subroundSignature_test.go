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
	//blockChain := blockchain.BlockChain{}
	//blockProcessorMock := initBlockProcessorMock()
	//bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
	//	return false
	//}}

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	//marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	//shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	//validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

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

func TestWorker_SendSignature(t *testing.T) {
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

func TestWorker_CheckCommitmentsValidityShouldErrNilCommitmet(t *testing.T) {
	sr := *initSubroundSignature()

	err := sr.CheckCommitmentsValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilCommitment, err)
}

func TestWorker_CheckCommitmentsValidityShouldErrInvalidIndex(t *testing.T) {
	sr := *initSubroundSignature()

	sr.MultiSigner().Reset(nil, 0)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitment, true)

	err := sr.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrInvalidIndex, err)
}

func TestWorker_CheckCommitmentsValidityShouldErrOnCommitmentHash(t *testing.T) {
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

func TestWorker_CheckCommitmentsValidityShouldErrCommitmentHashDoesNotMatch(t *testing.T) {
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

func TestWorker_CheckCommitmentsValidityShouldReturnNil(t *testing.T) {
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

func TestWorker_ReceivedSignature(t *testing.T) {
	sr := *initSubroundSignature()

	cnsDta := spos.NewConsensusData(
		sr.ConsensusState().Data,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	r := sr.ReceivedSignature(cnsDta)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsNotFinished)

	r = sr.ReceivedSignature(cnsDta)
	assert.False(t, r)

	sr.ConsensusState().RoundConsensus.SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrBitmap, true)

	r = sr.ReceivedSignature(cnsDta)
	assert.True(t, r)

	isSignJobDone, _ := sr.ConsensusState().GetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrSignature)
	assert.True(t, isSignJobDone)
}

func TestWorker_CheckSignatureConsensus(t *testing.T) {
	sr := *initSubroundSignature()

	sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsNotFinished)

	ok := sr.DoSignatureConsensusCheck()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sr.ConsensusState().Status(bn.SrSignature))

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	for i := 1; i < sr.ConsensusState().Threshold(bn.SrSignature); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrSignature, true)
	}

	ok = sr.DoSignatureConsensusCheck()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sr.ConsensusState().Status(bn.SrSignature))

	sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[0], bn.SrSignature, true)

	ok = sr.DoSignatureConsensusCheck()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sr.ConsensusState().Status(bn.SrSignature))
}

func TestWorker_SignaturesCollected(t *testing.T) {
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
	isJobDone, _ := sr.ConsensusState().GetJobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.SignaturesCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrSignature, true)
	isJobDone, _ = sr.ConsensusState().GetJobDone("B", bn.SrSignature)
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
