package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func initRoundConsensus() spos.RoundConsensus {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		len(eligibleList),
		"2")

	rcns.SetConsensusGroup(eligibleList)

	rcns.ResetRoundState()

	return rcns
}

func TestRoundConsensus_NewRoundConsensusShouldWork(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.NotNil(t, rcns)
	assert.Equal(t, 3, len(rcns.ConsensusGroup()))
	assert.Equal(t, "3", rcns.ConsensusGroup()[2])
	assert.Equal(t, "2", rcns.SelfPubKey())
}

func TestRoundConsensus_ConsensusGroupIndexFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key3")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key3")

	assert.Equal(t, 2, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_ConsensusGroupIndexNotFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key4")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key4")

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrNotFoundInConsensus, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupInConsesus(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key2")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.SelfConsensusGroupIndex()

	assert.Equal(t, 1, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupNotFound(t *testing.T) {
	t.Parallel()

	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key4")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.SelfConsensusGroupIndex()

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrNotFoundInConsensus, err)
}

func TestRoundConsensus_SetEligibleListShouldChangeTheEligibleList(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetEligibleList([]string{"4", "5", "6"})

	assert.Equal(t, "4", rcns.EligibleList()[0])
	assert.Equal(t, "5", rcns.EligibleList()[1])
	assert.Equal(t, "6", rcns.EligibleList()[2])
}

func TestRoundConsensus_SetConsensusGroupShouldChangeTheConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetConsensusGroup([]string{"4", "5", "6"})

	assert.Equal(t, "4", rcns.ConsensusGroup()[0])
	assert.Equal(t, "5", rcns.ConsensusGroup()[1])
	assert.Equal(t, "6", rcns.ConsensusGroup()[2])
}

func TestRoundConsensus_SetConsensusGroupSizeShouldChangeTheConsensusGroupSize(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, len(rcns.ConsensusGroup()), rcns.ConsensusGroupSize())
	rcns.SetConsensusGroupSize(99999)
	assert.Equal(t, 99999, rcns.ConsensusGroupSize())
}

func TestRoundConsensus_SetSelfPubKeyShouldChangeTheSelfPubKey(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetSelfPubKey("X")
	assert.Equal(t, "X", rcns.SelfPubKey())
}

func TestRoundConsensus_GetJobDoneShouldReturnsFalseWhenValidatorIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetJobDone("3", bn.SrBlock, true)
	rcns.SetConsensusGroup([]string{"1", "2"})
	isJobDone, _ := rcns.JobDone("3", bn.SrBlock)
	assert.False(t, isJobDone)
}

func TestRoundConsensus_SetJobDoneShouldNotBeSetWhenValidatorIsNotInTheConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetJobDone("4", bn.SrBlock, true)
	isJobDone, _ := rcns.JobDone("4", bn.SrBlock)
	assert.False(t, isJobDone)
}

func TestRoundConsensus_GetSelfJobDoneShouldReturnFalse(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		if rcns.ConsensusGroup()[i] == rcns.SelfPubKey() {
			continue
		}

		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, true)
	}

	jobDone, _ := rcns.SelfJobDone(bn.SrBlock)
	assert.False(t, jobDone)
}

func TestRoundConsensus_GetSelfJobDoneShouldReturnTrue(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetJobDone("2", bn.SrBlock, true)

	jobDone, _ := rcns.SelfJobDone(bn.SrBlock)
	assert.True(t, jobDone)
}

func TestRoundConsensus_SetSelfJobDoneShouldWork(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetSelfJobDone(bn.SrBlock, true)

	jobDone, _ := rcns.JobDone("2", bn.SrBlock)
	assert.True(t, jobDone)
}

func TestRoundConsensus_IsNodeInConsensusGroup(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, false, rcns.IsNodeInConsensusGroup("4"))
	assert.Equal(t, true, rcns.IsNodeInConsensusGroup(rcns.SelfPubKey()))
}

func TestRoundConsensus_IsNodeInEligibleList(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	assert.Equal(t, false, rcns.IsNodeInEligibleList("4"))
	assert.Equal(t, true, rcns.IsNodeInEligibleList(rcns.SelfPubKey()))
}

func TestRoundConsensus_ComputeSize(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetJobDone("1", bn.SrBlock, true)
	assert.Equal(t, 1, rcns.ComputeSize(bn.SrBlock))
}

func TestRoundConsensus_ResetValidationMap(t *testing.T) {
	t.Parallel()

	rcns := *initRoundConsensus()

	rcns.SetJobDone("1", bn.SrBlock, true)
	jobDone, _ := rcns.JobDone("1", bn.SrBlock)
	assert.Equal(t, true, jobDone)

	rcns.ConsensusGroup()[1] = "X"

	rcns.ResetRoundState()

	jobDone, err := rcns.JobDone("1", bn.SrBlock)
	assert.Equal(t, false, jobDone)
	assert.Nil(t, err)
}
