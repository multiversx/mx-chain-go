package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

func initSubroundCommitment() bn.SubroundCommitment {
	//blockChain := blockchain.BlockChain{}
	//blockProcessorMock := initBlockProcessorMock()
	//bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
	//	return false
	//}}

	consensusState := initConsensusState()
	//hasherMock := mock.HasherMock{}
	//marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	//shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	//validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

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

func TestWorker_SendCommitment(t *testing.T) {
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

func TestWorker_ReceivedCommitment(t *testing.T) {
	sr := *initSubroundCommitment()

	commitment := []byte("commitment")

	cnsDta := spos.NewConsensusData(
		sr.ConsensusState().Data,
		commitment,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtCommitment),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0)

	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)

	r := sr.ReceivedCommitment(cnsDta)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsNotFinished)

	r = sr.ReceivedCommitment(cnsDta)
	assert.False(t, r)

	sr.ConsensusState().RoundConsensus.SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrBitmap, true)

	r = sr.ReceivedCommitment(cnsDta)
	assert.True(t, r)
	isCommJobDone, _ := sr.ConsensusState().GetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrCommitment)
	assert.True(t, isCommJobDone)
}

func TestWorker_CheckCommitmentConsensus(t *testing.T) {
	sr := *initSubroundCommitment()

	sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsNotFinished)

	ok := sr.DoCommitmentConsensusCheck()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sr.ConsensusState().Status(bn.SrCommitment))

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBitmap); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrBitmap, true)
	}

	for i := 1; i < len(sr.ConsensusState().RoundConsensus.ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[i], bn.SrCommitment, true)
	}

	ok = sr.DoCommitmentConsensusCheck()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sr.ConsensusState().Status(bn.SrCommitment))

	sr.ConsensusState().SetJobDone(sr.ConsensusState().RoundConsensus.ConsensusGroup()[0], bn.SrCommitment, true)

	ok = sr.DoCommitmentConsensusCheck()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sr.ConsensusState().Status(bn.SrCommitment))
}

func TestWorker_CommitmentsCollected(t *testing.T) {
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
	isJobDone, _ := sr.ConsensusState().GetJobDone("C", bn.SrBitmap)
	assert.True(t, isJobDone)

	ok = sr.CommitmentsCollected(2)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("B", bn.SrCommitment, true)
	isJobDone, _ = sr.ConsensusState().GetJobDone("B", bn.SrCommitment)
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
