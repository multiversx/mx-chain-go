package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func TestWorker_SendCommitment(t *testing.T) {
	wrk := initWorker()

	r := wrk.DoCommitmentJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)

	r = wrk.DoCommitmentJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsNotFinished)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrCommitment, true)

	r = wrk.DoCommitmentJob()
	assert.False(t, r)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrCommitment, false)

	r = wrk.DoCommitmentJob()
	assert.False(t, r)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, true)
	wrk.SPoS.Data = nil

	r = wrk.DoCommitmentJob()
	assert.False(t, r)

	dta := []byte("X")
	wrk.SPoS.Data = dta

	r = wrk.DoCommitmentJob()
	assert.True(t, r)
}

func TestWorker_ReceivedCommitment(t *testing.T) {
	wrk := initWorker()

	commitment := []byte("commitment")
	commHash := mock.HasherMock{}.Compute(string(commitment))

	cnsDta := spos.NewConsensusData(
		wrk.SPoS.Data,
		commHash,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0)

	r := wrk.ReceivedCommitmentHash(cnsDta)

	cnsDta.MsgType = int(bn.MtCommitment)
	cnsDta.SubRoundData = commitment

	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)

	r = wrk.ReceivedCommitment(cnsDta)
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsNotFinished)

	r = wrk.ReceivedCommitment(cnsDta)
	assert.False(t, r)

	wrk.SPoS.RoundConsensus.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrBitmap, true)

	r = wrk.ReceivedCommitment(cnsDta)
	assert.True(t, r)
	isCommJobDone, _ := wrk.SPoS.GetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrCommitment)
	assert.True(t, isCommJobDone)
}

func TestWorker_CheckCommitmentConsensus(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsNotFinished)

	ok := wrk.CheckCommitmentConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrCommitment))

	for i := 0; i < wrk.SPoS.Threshold(bn.SrBitmap); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	for i := 1; i < len(wrk.SPoS.RoundConsensus.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitment, true)
	}

	ok = wrk.CheckCommitmentConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrCommitment))

	wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[0], bn.SrCommitment, true)

	ok = wrk.CheckCommitmentConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrCommitment))
}

func TestWorker_CommitmentsCollected(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := wrk.CommitmentsCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrBitmap, true)
	wrk.SPoS.SetJobDone("C", bn.SrBitmap, true)
	isJobDone, _ := wrk.SPoS.GetJobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = wrk.CommitmentsCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("B", bn.SrCommitment, true)
	isJobDone, _ = wrk.SPoS.GetJobDone("B", bn.SrCommitment)
	assert.True(t, isJobDone)

	ok = wrk.CommitmentsCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("C", bn.SrCommitment, true)
	ok = wrk.CommitmentsCollected(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrCommitment, true)
	ok = wrk.CommitmentsCollected(2)
	assert.True(t, ok)
}

func TestWorker_ExtendCommitment(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendCommitment()
	assert.Equal(t, spos.SsExtended, wrk.SPoS.Status(bn.SrCommitment))
}

func TestWorker_IsBitmapSubroundUnfinishedShouldRetunFalseWhenSubroundIsFinished(t *testing.T) {
	wrk := initWorker()
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)

	assert.False(t, wrk.IsBitmapSubroundUnfinished())
}

func TestWorker_IsBitmapSubroundUnfinishedShouldRetunFalseWhenJobIsDone(t *testing.T) {
	wrk := initWorker()

	wrk.Header = &block.Header{}

	wrk.SPoS.SetSelfPubKey(wrk.SPoS.ConsensusGroup()[0])
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		if wrk.SPoS.ConsensusGroup()[i] != wrk.SPoS.SelfPubKey() {
			wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, true)
		}
	}

	assert.False(t, wrk.IsBitmapSubroundUnfinished())
}

func TestWorker_IsBitmapSubroundUnfinishedShouldReturnTrueWhenDoBitmapJobReturnsFalse(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.IsBitmapSubroundUnfinished())
}

func TestWorker_IsBitmapSubroundUnfinishedShouldReturnTrueWhenCheckBitmapConsensusReturnsFalse(t *testing.T) {
	wrk := initWorker()

	wrk.Header = &block.Header{}

	wrk.SPoS.SetSelfPubKey(wrk.SPoS.ConsensusGroup()[0])
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	assert.True(t, wrk.IsBitmapSubroundUnfinished())
}
