package spos_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func internalInitConsensusState() *spos.ConsensusState {
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

	cns := internalInitConsensusState()
	assert.NotNil(t, cns)
}

func TestConsensusState_ResetConsensusStateShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()
	cns.RoundCanceled = true
	cns.ResetConsensusState()
	assert.False(t, cns.RoundCanceled)
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnFalseWhenGetLeaderErr(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetConsensusGroup(nil)
	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("2"))
}

func TestConsensusState_IsNodeLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensusState_IsSelfLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestConsensusState_IsSelfLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestConsensusState_GetLeaderShoudErrNilConsensusGroup(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetConsensusGroup(nil)

	_, err := cns.GetLeader()
	assert.Equal(t, spos.ErrNilConsensusGroup, err)
}

func TestConsensusState_GetLeaderShouldErrEmptyConsensusGroup(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetConsensusGroup(make([]string, 0))

	_, err := cns.GetLeader()
	assert.Equal(t, spos.ErrEmptyConsensusGroup, err)
}

func TestConsensusState_GetLeaderShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	leader, err := cns.GetLeader()
	assert.Nil(t, err)
	assert.Equal(t, cns.ConsensusGroup()[0], leader)
}

func TestConsensusState_GetNextConsensusGroupShouldFailWhenComputeValidatorsGroupErr(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	nodesCoordinator := mock.NodesCoordinatorMock{}
	err := errors.New("error")
	nodesCoordinator.ComputeValidatorsGroupCalled = func(
		randomness []byte,
		round uint64,
		shardId uint32,
	) ([]sharding.Validator, error) {
		return nil, err
	}

	_, err2 := cns.GetNextConsensusGroup([]byte(""), 0, 0, nodesCoordinator)
	assert.Equal(t, err, err2)
}

func TestConsensusState_GetNextConsensusGroupShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	nodesCoordinator := mock.NodesCoordinatorMock{}

	nextConsensusGroup, err := cns.GetNextConsensusGroup(nil, 0, 0, nodesCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, nextConsensusGroup)
}

func TestConsensusState_IsConsensusDataSetShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)

	assert.True(t, cns.IsConsensusDataSet())
}

func TestConsensusState_IsConsensusDataSetShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = nil

	assert.False(t, cns.IsConsensusDataSet())
}

func TestConsensusState_IsConsensusDataEqualShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	data := []byte("consensus data")

	cns.Data = data

	assert.True(t, cns.IsConsensusDataEqual(data))
}

func TestConsensusState_IsConsensusDataEqualShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	data := []byte("consensus data")

	cns.Data = data

	assert.False(t, cns.IsConsensusDataEqual([]byte("X")))
}

func TestConsensusState_IsJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetJobDone("1", bn.SrBlock, false)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))

	cns.SetJobDone("1", bn.SrCommitment, true)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))

	cns.SetJobDone("2", bn.SrBlock, true)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))
}

func TestConsensusState_IsJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetJobDone("1", bn.SrBlock, true)

	assert.True(t, cns.IsJobDone("1", bn.SrBlock))
}

func TestConsensusState_IsSelfJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))

	cns.SetJobDone(cns.SelfPubKey(), bn.SrCommitment, true)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))

	cns.SetJobDone(cns.SelfPubKey()+"X", bn.SrBlock, true)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))
}

func TestConsensusState_IsSelfJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)

	assert.True(t, cns.IsSelfJobDone(bn.SrBlock))
}

func TestConsensusState_IsCurrentSubroundFinishedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetStatus(bn.SrBlock, spos.SsNotFinished)
	assert.False(t, cns.IsCurrentSubroundFinished(bn.SrBlock))

	cns.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.False(t, cns.IsCurrentSubroundFinished(bn.SrBlock))

}

func TestConsensusState_IsCurrentSubroundFinishedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.SetStatus(bn.SrBlock, spos.SsFinished)
	assert.True(t, cns.IsCurrentSubroundFinished(bn.SrBlock))
}

func TestConsensusState_IsNodeSelfShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.False(t, cns.IsNodeSelf(cns.SelfPubKey()+"X"))
}

func TestConsensusState_IsNodeSelfShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	assert.True(t, cns.IsNodeSelf(cns.SelfPubKey()))
}

func TestConsensusState_IsBlockBodyAlreadyReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.BlockBody = nil

	assert.False(t, cns.IsBlockBodyAlreadyReceived())
}

func TestConsensusState_IsBlockBodyAlreadyReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.BlockBody = make(block.Body, 0)

	assert.True(t, cns.IsBlockBodyAlreadyReceived())
}

func TestConsensusState_IsHeaderAlreadyReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Header = nil

	assert.False(t, cns.IsHeaderAlreadyReceived())
}

func TestConsensusState_IsHeaderAlreadyReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Header = &block.Header{}

	assert.True(t, cns.IsHeaderAlreadyReceived())
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenConsensusDataNotSet(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = nil

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenSelfJobIsDone(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	cns.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanDoSubroundJobShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	cns.SetStatus(bn.SrBlock, spos.SsNotFinished)

	assert.True(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedFromItself(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte(cns.SelfPubKey()),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedForOtherRound(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 1, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenJobIsDone(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	cns.SetJobDone("1", bn.SrBlock, true)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	cns.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_CanProcessReceivedMessageShouldReturnTrue(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	cnsDta := &consensus.Message{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.True(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestConsensusState_GenerateBitmapShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	bitmapExpected := make([]byte, cns.ConsensusGroupSize()/8+1)
	selfIndexInConsensusGroup, _ := cns.SelfConsensusGroupIndex()
	bitmapExpected[selfIndexInConsensusGroup/8] |= 1 << (uint16(selfIndexInConsensusGroup) % 8)

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)
	bitmap := cns.GenerateBitmap(bn.SrBlock)

	assert.Equal(t, bitmapExpected, bitmap)
}

func TestConsensusState_SetAndGetProcessingBlockShouldWork(t *testing.T) {
	t.Parallel()
	cns := internalInitConsensusState()
	cns.SetProcessingBlock(true)

	assert.Equal(t, true, cns.ProcessingBlock())
}
