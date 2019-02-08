package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func TestWorker_SendBitmap(t *testing.T) {
	wrk := initWorker()

	wrk.Header = &block.Header{}

	r := wrk.DoBitmapJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)

	r = wrk.DoBitmapJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, true)

	r = wrk.DoBitmapJob()
	assert.False(t, r)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, false)
	wrk.SPoS.RoundConsensus.SetSelfPubKey(wrk.SPoS.RoundConsensus.ConsensusGroup()[1])

	r = wrk.DoBitmapJob()
	assert.False(t, r)

	wrk.SPoS.RoundConsensus.SetSelfPubKey(wrk.SPoS.RoundConsensus.ConsensusGroup()[0])
	wrk.SPoS.Data = nil

	r = wrk.DoBitmapJob()
	assert.False(t, r)

	dta := []byte("X")
	wrk.SPoS.Data = dta
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrCommitmentHash, true)

	r = wrk.DoBitmapJob()
	assert.True(t, r)
	isBitmapJobDone, _ := wrk.SPoS.GetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap)
	assert.True(t, isBitmapJobDone)
}

func TestWorker_ReceivedBitmap(t *testing.T) {
	wrk := initWorker()

	wrk.Header = &block.Header{}

	cnsDta := spos.NewConsensusData(
		wrk.SPoS.Data,
		[]byte("commHash"),
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)

	r := wrk.ReceivedBitmap(cnsDta)
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	bitmap := make([]byte, 3)

	cnGroup := wrk.SPoS.ConsensusGroup()

	// fill ony few of the signers in bitmap
	for i := 0; i < 5; i++ {
		bitmap[i/8] |= 1 << uint16(i%8)
	}

	cnsDta.SubRoundData = bitmap

	r = wrk.ReceivedBitmap(cnsDta)
	assert.False(t, r)

	//fill the rest
	for i := 5; i < len(cnGroup); i++ {
		bitmap[i/8] |= 1 << uint16(i%8)
	}

	cnsDta.SubRoundData = bitmap

	r = wrk.ReceivedBitmap(cnsDta)
	assert.True(t, r)
}

func TestWorker_IsValidatorInBitmapGroup(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	assert.False(t, wrk.IsValidatorInBitmap(wrk.SPoS.SelfPubKey()))
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, true)
	assert.True(t, wrk.IsValidatorInBitmap(wrk.SPoS.SelfPubKey()))
}

func TestWorker_IsSelfInBitmapGroupShoudReturnFalse(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		if wrk.SPoS.ConsensusGroup()[i] == wrk.SPoS.SelfPubKey() {
			continue
		}

		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	assert.False(t, wrk.IsSelfInBitmap())
}

func TestWorker_IsSelfInBitmapGroupShoudReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBitmap, true)
	assert.True(t, wrk.IsSelfInBitmap())
}

func TestWorker_CheckBitmapConsensus(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	ok := wrk.CheckBitmapConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrBitmap))

	for i := 1; i < wrk.SPoS.Threshold(bn.SrBitmap); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	ok = wrk.CheckBitmapConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrBitmap))

	wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[0], bn.SrBitmap, true)

	ok = wrk.CheckBitmapConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrBitmap))

	for i := 1; i < len(wrk.SPoS.RoundConsensus.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	wrk.SPoS.SetJobDone(wrk.SPoS.RoundConsensus.SelfPubKey(), bn.SrBitmap, false)

	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	ok = wrk.CheckBitmapConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrBitmap))
}

func TestWorker_IsBitmapReceived(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := wrk.IsBitmapReceived(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrBitmap, true)
	isJobDone, _ := wrk.SPoS.GetJobDone("A", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = wrk.IsBitmapReceived(2)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("B", bn.SrBitmap, true)
	ok = wrk.IsBitmapReceived(2)
	assert.True(t, ok)

	wrk.SPoS.SetJobDone("C", bn.SrBitmap, true)
	ok = wrk.IsBitmapReceived(2)
	assert.True(t, ok)
}

func TestWorker_ExtendBitmap(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendBitmap()
	assert.Equal(t, spos.SsExtended, wrk.SPoS.Status(bn.SrBitmap))
}

func TestWorker_IsCommitmentHashSubroundUnfinishedShouldRetunFalseWhenSubroundIsFinished(t *testing.T) {
	wrk := initWorker()
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)

	assert.False(t, wrk.IsCommitmentHashSubroundUnfinished())
}

func TestWorker_IsCommitmentHashSubroundUnfinishedShouldRetunFalseWhenJobIsDone(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		if wrk.SPoS.ConsensusGroup()[i] != wrk.SPoS.SelfPubKey() {
			wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, true)
		}
	}

	assert.False(t, wrk.IsCommitmentHashSubroundUnfinished())
}

func TestWorker_IsCommitmentHashSubroundUnfinishedShouldReturnTrueWhenDoCommitmentHashJobReturnsFalse(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.IsCommitmentHashSubroundUnfinished())
}

func TestWorker_IsCommitmentHashSubroundUnfinishedShouldReturnTrueWhenCheckCommitmentHashConsensusReturnsFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.True(t, wrk.IsCommitmentHashSubroundUnfinished())
}
