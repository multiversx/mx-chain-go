package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestNewRoundConsensus(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {

		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrSignature, false)
	}

	assert.Equal(t, 3, len(rcns.ConsensusGroup()))
	assert.Equal(t, "3", rcns.ConsensusGroup()[2])
	assert.Equal(t, "2", rcns.SelfPubKey())
}

func TestRoundConsensus_ResetValidationMap(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrSignature, false)
	}

	rcns.SetJobDone("1", bn.SrBlock, true)
	jobDone, _ := rcns.GetJobDone("1", bn.SrBlock)
	assert.Equal(t, true, jobDone)

	rcns.ResetRoundState()
	jobDone, err := rcns.GetJobDone("1", bn.SrBlock)
	assert.Equal(t, false, jobDone)
	assert.Nil(t, err)
}

func TestRoundConsensus_IsNodeInValidationGroup(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrSignature, false)
	}

	assert.Equal(t, false, rcns.IsNodeInConsensusGroup("4"))
	assert.Equal(t, true, rcns.IsNodeInConsensusGroup(rcns.SelfPubKey()))
}

func TestRoundConsensus_ComputeSize(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrSignature, false)
	}

	rcns.SetJobDone("1", bn.SrBlock, true)
	assert.Equal(t, 1, rcns.ComputeSize(bn.SrBlock))
}

func TestRoundConsensus_ConsensusGroupIndexFound(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key3")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key3")

	assert.Equal(t, 2, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_ConsensusGroupIndexNotFound(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key4")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.ConsensusGroupIndex("key4")

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrSelfNotFoundInConsensus, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupInConsesus(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key2")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.IndexSelfConsensusGroup()

	assert.Equal(t, 1, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupNotFound(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rcns := spos.NewRoundConsensus(pubKeys, 3, "key4")
	rcns.SetConsensusGroup(pubKeys)
	index, err := rcns.IndexSelfConsensusGroup()

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrSelfNotFoundInConsensus, err)
}

func TestRoundConsensus_SetConsensusGroupShouldChangeTheConsensusGroup(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	rcns.SetConsensusGroup([]string{"4", "5", "6"})

	assert.Equal(t, "4", rcns.ConsensusGroup()[0])
	assert.Equal(t, "5", rcns.ConsensusGroup()[1])
	assert.Equal(t, "6", rcns.ConsensusGroup()[2])
}

func TestRoundConsensus_GetJobDoneShouldReturnsFalseWhenValidatorIsNotInTheConsensusGroup(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	rcns.SetJobDone("3", bn.SrBlock, true)
	rcns.SetConsensusGroup([]string{"1", "2"})
	isJobDone, _ := rcns.GetJobDone("3", bn.SrBlock)
	assert.False(t, isJobDone)
}

func TestRoundConsensus_SetJobDoneShouldNotBeSetWhenValidatorIsNotInTheConsensusGroup(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	rcns.SetJobDone("4", bn.SrBlock, true)
	isJobDone, _ := rcns.GetJobDone("4", bn.SrBlock)
	assert.False(t, isJobDone)
}

func TestSetSelfJobDone_ShouldWork(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	rcns.SetSelfJobDone(bn.SrBlock, true)

	jobDone, _ := rcns.GetJobDone("2", bn.SrBlock)
	assert.True(t, jobDone)
}

func TestGetSelfJobDone_ShouldReturnFalse(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		if rcns.ConsensusGroup()[i] == rcns.SelfPubKey() {
			continue
		}

		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, true)
	}

	jobDone, _ := rcns.GetSelfJobDone(bn.SrBlock)
	assert.False(t, jobDone)
}

func TestGetSelfJobDone_ShouldReturnTrue(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	rcns.SetJobDone("2", bn.SrBlock, true)

	jobDone, _ := rcns.GetSelfJobDone(bn.SrBlock)
	assert.True(t, jobDone)
}
