package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func initConsensusState() *spos.ConsensusState {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)
	rcns.ResetRoundState()

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, 3)
	rthr.SetThreshold(bn.SrBitmap, 3)
	rthr.SetThreshold(bn.SrCommitment, 3)
	rthr.SetThreshold(bn.SrSignature, 3)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	return cns
}

func TestConsensusState__NewConsensusStateShouldWork(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()
	assert.NotNil(t, cns)
}

func TestConsensusState_ResetConsensusStateShouldWork(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()
	cns.RoundCanceled = true
	cns.ResetConsensusState()
	assert.False(t, cns.RoundCanceled)
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnFalseWhenGetLeaderErr(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetConsensusGroup(nil)
	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("2"))
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensusState_IsSelfLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestConsensusState_IsSelfLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestConsensusState_GetLeaderShoudErrNilConsensusGroup(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetConsensusGroup(nil)

	_, err := cns.GetLeader()
	assert.Equal(t, spos.ErrNilConsensusGroup, err)
}

func TestConsensusState_GetLeaderShouldErrEmptyConsensusGroup(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetConsensusGroup(make([]string, 0))

	_, err := cns.GetLeader()
	assert.Equal(t, spos.ErrEmptyConsensusGroup, err)
}

func TestConsensusState_GetLeaderShouldWork(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	leader, err := cns.GetLeader()
	assert.Nil(t, err)
	assert.Equal(t, cns.ConsensusGroup()[0], leader)
}

func TestConsensusState_GetNextConsensusGroupShouldFailWhenComputeValidatorsGroupErr(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	vgs := mock.ValidatorGroupSelectorMock{}
	err := errors.New("error")
	vgs.ComputeValidatorsGroupCalled = func(randomness []byte) ([]consensus.Validator, error) {
		return nil, err
	}

	_, err2 := cns.GetNextConsensusGroup("", vgs)
	assert.Equal(t, err, err2)
}

func TestConsensusState_GetNextConsensusGroupShouldWork(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	vgs := mock.ValidatorGroupSelectorMock{}

	nextConsensusGroup, err := cns.GetNextConsensusGroup("", vgs)
	assert.Nil(t, err)
	assert.NotNil(t, nextConsensusGroup)
}

func TestConsensusState_IsConsensusDataSetShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Data = make([]byte, 0)

	assert.True(t, cns.IsConsensusDataSet())
}

func TestConsensusState_IsConsensusDataNotSetShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Data = nil

	assert.False(t, cns.IsConsensusDataSet())
}

func TestConsensusState_IsJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetJobDone("1", bn.SrBlock, false)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))

	cns.SetJobDone("1", bn.SrCommitment, true)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))

	cns.SetJobDone("2", bn.SrBlock, true)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))
}

func TestConsensusState_IsJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetJobDone("1", bn.SrBlock, true)

	assert.True(t, cns.IsJobDone("1", bn.SrBlock))
}

func TestConsensusState_IsSelfJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))

	cns.SetJobDone(cns.SelfPubKey(), bn.SrCommitment, true)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))

	cns.SetJobDone(cns.SelfPubKey()+"X", bn.SrBlock, true)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))
}

func TestConsensusState_IsSelfJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)

	assert.True(t, cns.IsSelfJobDone(bn.SrBlock))
}

func TestConsensusState_IsCurrentSubroundFinishedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetStatus(bn.SrBlock, spos.SsNotFinished)
	assert.False(t, cns.IsCurrentSubroundFinished(bn.SrBlock))

	cns.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.False(t, cns.IsCurrentSubroundFinished(bn.SrBlock))

}

func TestConsensusState_IsCurrentSubroundFinishedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.SetStatus(bn.SrBlock, spos.SsFinished)
	assert.True(t, cns.IsCurrentSubroundFinished(bn.SrBlock))
}

func TestConsensusState_IsNodeSelfShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	assert.False(t, cns.IsNodeSelf(cns.SelfPubKey()+"X"))
}

func TestConsensusState_IsNodeSelfShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	assert.True(t, cns.IsNodeSelf(cns.SelfPubKey()))
}

func TestConsensusState_IsBlockBodyAlreadyReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.BlockBody = nil

	assert.False(t, cns.IsBlockBodyAlreadyReceived())
}

func TestConsensusState_IsBlockBodyAlreadyReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.BlockBody = &block.TxBlockBody{}

	assert.True(t, cns.IsBlockBodyAlreadyReceived())
}

func TestConsensusState_IsHeaderAlreadyReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Header = nil

	assert.False(t, cns.IsHeaderAlreadyReceived())
}

func TestConsensusState_IsHeaderAlreadyReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Header = &block.Header{}

	assert.True(t, cns.IsHeaderAlreadyReceived())
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenConsensusDataNotSet(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Data = nil

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenSelfJobIsDone(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	cns.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	cns.SetStatus(bn.SrBlock, spos.SsNotFinished)

	assert.True(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedFromItself(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cnsDta := &spos.ConsensusMessage{
		RoundIndex: 0,
		PubKey:     []byte(cns.SelfPubKey()),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedForOtherRound(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cnsDta := &spos.ConsensusMessage{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 1, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenJobIsDone(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cnsDta := &spos.ConsensusMessage{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	cns.SetJobDone("1", bn.SrBlock, true)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cnsDta := &spos.ConsensusMessage{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	cns.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	cnsDta := &spos.ConsensusMessage{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.True(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_GenerateBitmapShouldWork(t *testing.T) {
	t.Parallel()

	cns := initConsensusState()

	bitmapExpected := make([]byte, cns.ConsensusGroupSize()/8+1)
	selfIndexInConsensusGroup, _ := cns.SelfConsensusGroupIndex()
	bitmapExpected[selfIndexInConsensusGroup/8] |= 1 << (uint16(selfIndexInConsensusGroup) % 8)

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)
	bitmap := cns.GenerateBitmap(bn.SrBlock)

	assert.Equal(t, bitmapExpected, bitmap)
}
