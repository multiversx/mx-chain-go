package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestNewRoundConsensus(t *testing.T) {

	roundConsensus := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(roundConsensus.ConsensusGroup()); i++ {

		roundConsensus.SetJobDone(roundConsensus.ConsensusGroup()[i], spos.SrBlock, false)
		roundConsensus.SetJobDone(roundConsensus.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		roundConsensus.SetJobDone(roundConsensus.ConsensusGroup()[i], spos.SrBitmap, false)
		roundConsensus.SetJobDone(roundConsensus.ConsensusGroup()[i], spos.SrCommitment, false)
		roundConsensus.SetJobDone(roundConsensus.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, 3, len(roundConsensus.ConsensusGroup()))
	assert.Equal(t, "3", roundConsensus.ConsensusGroup()[2])
	assert.Equal(t, "2", roundConsensus.SelfPubKey())
}

func TestRoundConsensus_ResetValidationMap(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	rCns.SetJobDone("1", spos.SrBlock, true)
	jobDone, _ := rCns.GetJobDone("1", spos.SrBlock)
	assert.Equal(t, true, jobDone)

	rCns.ResetRoundState()
	jobDone, err := rCns.GetJobDone("1", spos.SrBlock)
	assert.Equal(t, false, jobDone)
	assert.Nil(t, err)
}

func TestRoundConsensus_IsNodeInBitmapGroup(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, false, rCns.IsValidatorInBitmap(rCns.SelfPubKey()))
	rCns.SetJobDone(rCns.SelfPubKey(), spos.SrBitmap, true)
	assert.Equal(t, true, rCns.IsValidatorInBitmap(rCns.SelfPubKey()))
}

func TestRoundConsensus_IsNodeInValidationGroup(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, false, rCns.IsNodeInConsensusGroup("4"))
	assert.Equal(t, true, rCns.IsNodeInConsensusGroup(rCns.SelfPubKey()))
}

func TestRoundConsensus_IsBlockReceived(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := rCns.IsBlockReceived(1)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrBlock, true)
	isJobDone, _ := rCns.GetJobDone("1", spos.SrBlock)

	assert.Equal(t, true, isJobDone)

	ok = rCns.IsBlockReceived(1)
	assert.Equal(t, true, ok)

	ok = rCns.IsBlockReceived(2)
	assert.Equal(t, false, ok)
}

func TestRoundConsensus_IsCommitmentHashReceived(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := rCns.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrCommitmentHash, true)
	isJobDone, _ := rCns.GetJobDone("1", spos.SrCommitmentHash)
	assert.Equal(t, true, isJobDone)

	ok = rCns.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("2", spos.SrCommitmentHash, true)
	ok = rCns.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)

	rCns.SetJobDone("3", spos.SrCommitmentHash, true)
	ok = rCns.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)
}

func TestRoundConsensus_IsBitmapInCommitmentHash(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := rCns.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrBitmap, true)
	rCns.SetJobDone("3", spos.SrBitmap, true)
	isJobDone, _ := rCns.GetJobDone("3", spos.SrBitmap)
	assert.Equal(t, true, isJobDone)

	ok = rCns.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("2", spos.SrCommitmentHash, true)
	isJobDone, _ = rCns.GetJobDone("2", spos.SrCommitmentHash)
	assert.Equal(t, true, isJobDone)

	ok = rCns.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("3", spos.SrCommitmentHash, true)
	ok = rCns.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrCommitmentHash, true)
	ok = rCns.CommitmentHashesCollected(2)
	assert.Equal(t, true, ok)
}

func TestRoundConsensus_IsBitmapInCommitment(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := rCns.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrBitmap, true)
	rCns.SetJobDone("3", spos.SrBitmap, true)
	isJobDone, _ := rCns.GetJobDone("3", spos.SrBitmap)
	assert.Equal(t, true, isJobDone)

	ok = rCns.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("2", spos.SrCommitment, true)
	isJobDone, _ = rCns.GetJobDone("2", spos.SrCommitment)
	assert.Equal(t, true, isJobDone)

	ok = rCns.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("3", spos.SrCommitment, true)
	ok = rCns.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrCommitment, true)
	ok = rCns.CommitmentsCollected(2)
	assert.Equal(t, true, ok)
}

func TestRoundConsensus_IsBitmapInSignature(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := rCns.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrBitmap, true)
	rCns.SetJobDone("3", spos.SrBitmap, true)
	isJobDone, _ := rCns.GetJobDone("3", spos.SrBitmap)
	assert.Equal(t, true, isJobDone)

	ok = rCns.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("2", spos.SrSignature, true)
	isJobDone, _ = rCns.GetJobDone("2", spos.SrSignature)
	assert.Equal(t, true, isJobDone)

	ok = rCns.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("3", spos.SrSignature, true)
	ok = rCns.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	rCns.SetJobDone("1", spos.SrSignature, true)
	ok = rCns.SignaturesCollected(2)
	assert.Equal(t, true, ok)
}

func TestRoundConsensus_ComputeSize(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	rCns.SetJobDone("1", spos.SrBlock, true)
	assert.Equal(t, 1, rCns.ComputeSize(spos.SrBlock))
}

func TestRoundConsensus_ConsensusGroupIndexFound(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rCns := spos.NewRoundConsensus(pubKeys, "key3")
	index, err := rCns.ConsensusGroupIndex("key3")

	assert.Equal(t, 2, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_ConsensusGroupIndexNotFound(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rCns := spos.NewRoundConsensus(pubKeys, "key4")
	index, err := rCns.ConsensusGroupIndex("key4")

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrSelfNotFoundInConsensus, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupInConsesus(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rCns := spos.NewRoundConsensus(pubKeys, "key2")
	index, err := rCns.IndexSelfConsensusGroup()

	assert.Equal(t, 1, index)
	assert.Nil(t, err)
}

func TestRoundConsensus_IndexSelfConsensusGroupNotFound(t *testing.T) {
	pubKeys := []string{"key1", "key2", "key3"}

	rCns := spos.NewRoundConsensus(pubKeys, "key4")
	index, err := rCns.IndexSelfConsensusGroup()

	assert.Zero(t, index)
	assert.Equal(t, spos.ErrSelfNotFoundInConsensus, err)
}
