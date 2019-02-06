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

func TestWorker_SendBitmap(t *testing.T) {
	cnWorkers := initSposWorkers()

	r := cnWorkers[0].DoBitmapJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	cnWorkers[0].SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)

	r = cnWorkers[0].DoBitmapJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBitmap, true)

	r = cnWorkers[0].DoBitmapJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBitmap, false)
	cnWorkers[0].SPoS.RoundConsensus.SetSelfPubKey(cnWorkers[0].SPoS.RoundConsensus.ConsensusGroup()[1])

	r = cnWorkers[0].DoBitmapJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.RoundConsensus.SetSelfPubKey(cnWorkers[0].SPoS.RoundConsensus.ConsensusGroup()[0])
	cnWorkers[0].SPoS.Data = nil

	r = cnWorkers[0].DoBitmapJob()
	assert.False(t, r)

	dta := []byte("X")
	cnWorkers[0].SPoS.Data = dta
	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrCommitmentHash, true)

	r = cnWorkers[0].DoBitmapJob()
	assert.True(t, r)
	isBitmapJobDone, _ := cnWorkers[0].SPoS.GetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBitmap)
	assert.True(t, isBitmapJobDone)
}

func TestWorker_ReceivedBitmap(t *testing.T) {
	cnWorkers := initSposWorkers()
	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	cnsDta := spos.NewConsensusData(
		cnWorkers[0].SPoS.Data,
		[]byte("commHash"),
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)

	r := cnWorkers[0].ReceivedBitmap(cnsDta)
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	bitmap := make([]byte, 3)

	cnGroup := cnWorkers[0].SPoS.ConsensusGroup()

	// fill ony few of the signers in bitmap
	for i := 0; i < 5; i++ {
		bitmap[i/8] |= 1 << uint16(i%8)
	}

	cnsDta.SubRoundData = bitmap

	r = cnWorkers[0].ReceivedBitmap(cnsDta)
	assert.False(t, r)

	//fill the rest
	for i := 5; i < len(cnGroup); i++ {
		bitmap[i/8] |= 1 << uint16(i%8)
	}

	cnsDta.SubRoundData = bitmap

	r = cnWorkers[0].ReceivedBitmap(cnsDta)
	assert.True(t, r)
}

func TestWorker_IsValidatorInBitmapGroup(t *testing.T) {
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

	assert.False(t, worker.IsValidatorInBitmap(worker.SPoS.SelfPubKey()))
	worker.SPoS.SetJobDone(worker.SPoS.SelfPubKey(), bn.SrBitmap, true)
	assert.True(t, worker.IsValidatorInBitmap(worker.SPoS.SelfPubKey()))
}

func TestWorker_IsSelfInBitmapGroupShoudReturnFalse(t *testing.T) {
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
		if worker.SPoS.ConsensusGroup()[i] == worker.SPoS.SelfPubKey() {
			continue
		}

		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	assert.False(t, worker.IsSelfInBitmap())
}

func TestWorker_IsSelfInBitmapGroupShoudReturnTrue(t *testing.T) {
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

	worker.SPoS.SetJobDone(worker.SPoS.SelfPubKey(), bn.SrBitmap, true)
	assert.True(t, worker.IsSelfInBitmap())
}

func TestWorker_CheckBitmapConsensus(t *testing.T) {
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

	sPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	ok := worker.CheckBitmapConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrBitmap))

	for i := 1; i < sPoS.Threshold(bn.SrBitmap); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	ok = worker.CheckBitmapConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrBitmap))

	sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[0], bn.SrBitmap, true)

	ok = worker.CheckBitmapConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrBitmap))

	for i := 1; i < len(sPoS.RoundConsensus.ConsensusGroup()); i++ {
		sPoS.SetJobDone(sPoS.RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	sPoS.SetJobDone(sPoS.RoundConsensus.SelfPubKey(), bn.SrBitmap, false)

	sPoS.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	ok = worker.CheckBitmapConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrBitmap))
}

func TestWorker_IsBitmapReceived(t *testing.T) {
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

	ok := worker.IsBitmapReceived(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrBitmap, true)
	isJobDone, _ := worker.SPoS.GetJobDone("1", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = worker.IsBitmapReceived(2)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("2", bn.SrBitmap, true)
	ok = worker.IsBitmapReceived(2)
	assert.True(t, ok)

	worker.SPoS.SetJobDone("3", bn.SrBitmap, true)
	ok = worker.IsBitmapReceived(2)
	assert.True(t, ok)
}

func TestWorker_ExtendBitmap(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].ExtendBitmap()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].SPoS.Status(bn.SrBitmap))
}
