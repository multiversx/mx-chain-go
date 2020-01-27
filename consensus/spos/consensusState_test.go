package spos_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
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

	nodesCoordinator := &mock.NodesCoordinatorMock{}
	err := errors.New("error")
	nodesCoordinator.ComputeValidatorsGroupCalled = func(
		randomness []byte,
		round uint64,
		shardId uint32,
	) ([]sharding.Validator, error) {
		return nil, err
	}

	_, _, err2 := cns.GetNextConsensusGroup([]byte(""), 0, 0, nodesCoordinator)
	assert.Equal(t, err, err2)
}

func TestConsensusState_GetNextConsensusGroupShouldWork(t *testing.T) {
	t.Parallel()

	cns := internalInitConsensusState()

	nodesCoordinator := &mock.NodesCoordinatorMock{}

	nextConsensusGroup, rewardAddresses, err := cns.GetNextConsensusGroup(nil, 0, 0, nodesCoordinator)
	assert.Nil(t, err)
	assert.NotNil(t, nextConsensusGroup)
	assert.NotNil(t, rewardAddresses)
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

func TestConsensusState_SetAndGetProcessingBlockShouldWork(t *testing.T) {
	t.Parallel()
	cns := internalInitConsensusState()
	cns.SetProcessingBlock(true)

	assert.Equal(t, true, cns.ProcessingBlock())
}
