package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWorker_SendSignature(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetSelfPubKey(wrk.SPoS.ConsensusGroup()[0])

	r := wrk.DoSignatureJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	r = wrk.DoSignatureJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsNotFinished)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrSignature, true)

	r = wrk.DoSignatureJob()
	assert.False(t, r)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrSignature, false)

	r = wrk.DoSignatureJob()
	assert.False(t, r)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, true)
	wrk.SPoS.Data = nil

	r = wrk.DoSignatureJob()
	assert.False(t, r)

	dta := []byte("X")
	wrk.SPoS.Data = dta

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrCommitment, true)
	r = wrk.DoSignatureJob()
	assert.True(t, r)
}

func TestWorker_CheckCommitmentsValidityShouldErrNilCommitmet(t *testing.T) {
	wrk := initWorker()

	err := wrk.CheckCommitmentsValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilCommitment, err)
}

func TestWorker_CheckCommitmentsValidityShouldErrInvalidIndex(t *testing.T) {
	wrk := initWorker()

	wrk.MultiSigner().Reset(nil, 0)

	wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrCommitment, true)

	err := wrk.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrInvalidIndex, err)
}

func TestWorker_CheckCommitmentsValidityShouldErrOnCommitmentHash(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

	err := errors.New("error commitment hash")
	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return nil, err
	}

	wrk.SetMultiSigner(multiSignerMock)

	wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrCommitment, true)

	err2 := wrk.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, err, err2)
}

func TestWorker_CheckCommitmentsValidityShouldErrCommitmentHashDoesNotMatch(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	wrk.SetMultiSigner(multiSignerMock)

	wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrCommitment, true)

	err := wrk.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, spos.ErrCommitmentHashDoesNotMatch, err)
}

func TestWorker_CheckCommitmentsValidityShouldReturnNil(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

	multiSignerMock.CommitmentMock = func(uint16) ([]byte, error) {
		return []byte("X"), nil
	}

	multiSignerMock.CommitmentHashMock = func(uint16) ([]byte, error) {
		return mock.HasherMock{}.Compute(string([]byte("X"))), nil
	}

	wrk.SetMultiSigner(multiSignerMock)

	wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrCommitment, true)

	err := wrk.CheckCommitmentsValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}

func TestWorker_ReceivedSignature(t *testing.T) {
	wrk := initWorker()

	cnsDta := spos.NewConsensusData(
		wrk.SPoS.Data,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	r := wrk.ReceivedSignature(cnsDta)
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsNotFinished)

	r = wrk.ReceivedSignature(cnsDta)
	assert.False(t, r)

	wrk.SPoS.RoundConsensus.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrBitmap, true)

	r = wrk.ReceivedSignature(cnsDta)
	assert.True(t, r)

	isSignJobDone, _ := wrk.SPoS.GetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrSignature)
	assert.True(t, isSignJobDone)
}

func TestWorker_CheckSignatureConsensus(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsNotFinished)

	ok := wrk.CheckSignatureConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrSignature))

	for i := 0; i < wrk.SPoS.Threshold(bn.SrBitmap); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	for i := 1; i < wrk.SPoS.Threshold(bn.SrSignature); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrSignature, true)
	}

	ok = wrk.CheckSignatureConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrSignature))

	wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[0], bn.SrSignature, true)

	ok = wrk.CheckSignatureConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrSignature))
}

func TestWorker_SignaturesCollected(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := wrk.SignaturesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrBitmap, true)
	wrk.SPoS.SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := wrk.SPoS.GetJobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = wrk.SignaturesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("B", bn.SrSignature, true)
	isJobDone, _ = wrk.SPoS.GetJobDone("B", bn.SrSignature)
	assert.True(t, isJobDone)

	ok = wrk.SignaturesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("C", bn.SrSignature, true)
	ok = wrk.SignaturesCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrSignature, true)
	ok = wrk.SignaturesCollected(2)
	assert.True(t, ok)
}

func TestWorker_ExtendSignature(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendSignature()
	assert.Equal(t, spos.SsExtended, wrk.SPoS.Status(bn.SrSignature))
}

func TestWorker_IsCommitmentSubroundUnfinishedShouldRetunFalseWhenSubroundIsFinished(t *testing.T) {
	wrk := initWorker()
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)

	assert.False(t, wrk.IsCommitmentSubroundUnfinished())
}

func TestWorker_IsCommitmentSubroundUnfinishedShouldRetunFalseWhenJobIsDone(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, true)

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, true)

		if wrk.SPoS.ConsensusGroup()[i] != wrk.SPoS.SelfPubKey() {
			wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, true)
		}
	}

	assert.False(t, wrk.IsCommitmentSubroundUnfinished())
}

func TestWorker_IsCommitmentSubroundUnfinishedShouldReturnTrueWhenDoCommitmentReturnsFalse(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.IsCommitmentSubroundUnfinished())
}

func TestWorker_IsCommitmentSubroundUnfinishedShouldReturnTrueWhenCheckCommitmentConsensusReturnsFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, true)

	assert.True(t, wrk.IsCommitmentSubroundUnfinished())
}
