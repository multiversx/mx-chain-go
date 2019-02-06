package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestWorker_SendCommitment(t *testing.T) {
	cnWorkers := initSposWorkers()

	r := cnWorkers[0].DoCommitmentJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	cnWorkers[0].SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)

	r = cnWorkers[0].DoCommitmentJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitment, spos.SsNotFinished)
	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrCommitment, true)

	r = cnWorkers[0].DoCommitmentJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrCommitment, false)

	r = cnWorkers[0].DoCommitmentJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBitmap, true)
	cnWorkers[0].SPoS.Data = nil

	r = cnWorkers[0].DoCommitmentJob()
	assert.False(t, r)

	dta := []byte("X")
	cnWorkers[0].SPoS.Data = dta

	r = cnWorkers[0].DoCommitmentJob()
	assert.True(t, r)
}

func TestWorker_ReceivedCommitment(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	commitment := []byte("commitment")
	commHash := mock.HasherMock{}.Compute(string(commitment))

	cnsDta := spos.NewConsensusData(
		cnWorkers[0].SPoS.Data,
		commHash,
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0)

	r := cnWorkers[0].ReceivedCommitmentHash(cnsDta)

	cnsDta.MsgType = int(bn.MtCommitment)
	cnsDta.SubRoundData = commitment

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)

	r = cnWorkers[0].ReceivedCommitment(cnsDta)
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitment, spos.SsNotFinished)

	r = cnWorkers[0].ReceivedCommitment(cnsDta)
	assert.False(t, r)

	cnWorkers[0].SPoS.RoundConsensus.SetJobDone(cnWorkers[0].SPoS.ConsensusGroup()[0], bn.SrBitmap, true)

	r = cnWorkers[0].ReceivedCommitment(cnsDta)
	assert.True(t, r)
	isCommJobDone, _ := cnWorkers[0].SPoS.GetJobDone(cnWorkers[0].SPoS.ConsensusGroup()[0], bn.SrCommitment)
	assert.True(t, isCommJobDone)
}

func TestWorker_CheckCommitmentConsensus(t *testing.T) {
	sPoS := InitSposWorker()

	worker, _ := bn.NewWorker(
		sPoS,
		&blockchain.BlockChain{},
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BootstrapMock{ShouldSyncCalled: func() bool {
			return false
		}},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, sPoS, worker)

	sPoS.SetStatus(bn.SrCommitment, spos.SsNotFinished)

	ok := worker.CheckCommitmentConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrCommitment))

	for i := 0; i < sPoS.Threshold(bn.SrBitmap); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	for i := 1; i < len(sPoS.RoundConsensus.ConsensusGroup()); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrCommitment, true)
	}

	ok = worker.CheckCommitmentConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrCommitment))

	sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[0], bn.SrCommitment, true)

	ok = worker.CheckCommitmentConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrCommitment))
}

func TestWorker_CommitmentsCollected(t *testing.T) {
	sPoS := InitSposWorker()

	worker, _ := bn.NewWorker(
		sPoS,
		&blockchain.BlockChain{},
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BootstrapMock{ShouldSyncCalled: func() bool {
			return false
		}},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, sPoS, worker)

	for i := 0; i < len(worker.SPoS.ConsensusGroup()); i++ {
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := worker.CommitmentsCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrBitmap, true)
	worker.SPoS.SetJobDone("3", bn.SrBitmap, true)
	isJobDone, _ := worker.SPoS.GetJobDone("3", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = worker.CommitmentsCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("2", bn.SrCommitment, true)
	isJobDone, _ = worker.SPoS.GetJobDone("2", bn.SrCommitment)
	assert.True(t, isJobDone)

	ok = worker.CommitmentsCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("3", bn.SrCommitment, true)
	ok = worker.CommitmentsCollected(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrCommitment, true)
	ok = worker.CommitmentsCollected(2)
	assert.True(t, ok)
}

func TestWorker_ExtendCommitment(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].ExtendCommitment()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].SPoS.Status(bn.SrCommitment))
}
