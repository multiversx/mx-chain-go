package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestNewValidators(t *testing.T) {

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

func TestValidators_ResetValidationMap(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	vld.SetJobDone("1", spos.SrBlock, true)
	assert.Equal(t, true, vld.GetJobDone("1", spos.SrBlock))

	vld.ResetRoundState()
	assert.Equal(t, false, vld.GetJobDone("1", spos.SrBlock))
}

func TestValidators_IsNodeInBitmapGroup(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, false, vld.IsValidatorInBitmap(vld.SelfPubKey()))
	vld.SetJobDone(vld.SelfPubKey(), spos.SrBitmap, true)
	assert.Equal(t, true, vld.IsValidatorInBitmap(vld.SelfPubKey()))
}

func TestValidators_IsNodeInValidationGroup(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, false, vld.IsNodeInConsensusGroup("4"))
	assert.Equal(t, true, vld.IsNodeInConsensusGroup(vld.SelfPubKey()))
}

func TestValidators_IsBlockReceived(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.IsBlockReceived(1)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrBlock, true)
	assert.Equal(t, true, vld.GetJobDone("1", spos.SrBlock))

	ok = vld.IsBlockReceived(1)
	assert.Equal(t, true, ok)

	ok = vld.IsBlockReceived(2)
	assert.Equal(t, false, ok)
}

func TestValidators_IsCommitmentHashReceived(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrCommitmentHash, true)
	assert.Equal(t, true, vld.GetJobDone("1", spos.SrCommitmentHash))

	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("2", spos.SrCommitmentHash, true)
	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)

	vld.SetJobDone("3", spos.SrCommitmentHash, true)
	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInCommitmentHash(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrBitmap, true)
	vld.SetJobDone("3", spos.SrBitmap, true)
	assert.Equal(t, true, vld.GetJobDone("3", spos.SrBitmap))

	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("2", spos.SrCommitmentHash, true)
	assert.Equal(t, true, vld.GetJobDone("2", spos.SrCommitmentHash))

	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("3", spos.SrCommitmentHash, true)
	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrCommitmentHash, true)
	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInCommitment(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrBitmap, true)
	vld.SetJobDone("3", spos.SrBitmap, true)
	assert.Equal(t, true, vld.GetJobDone("3", spos.SrBitmap))

	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("2", spos.SrCommitment, true)
	assert.Equal(t, true, vld.GetJobDone("2", spos.SrCommitment))

	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("3", spos.SrCommitment, true)
	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrCommitment, true)
	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInSignature(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrBitmap, true)
	vld.SetJobDone("3", spos.SrBitmap, true)
	assert.Equal(t, true, vld.GetJobDone("3", spos.SrBitmap))

	ok = vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("2", spos.SrSignature, true)
	assert.Equal(t, true, vld.GetJobDone("2", spos.SrSignature))

	ok = vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("3", spos.SrSignature, true)
	ok = vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetJobDone("1", spos.SrSignature, true)
	ok = vld.SignaturesCollected(2)
	assert.Equal(t, true, ok)
}

func TestValidators_ComputeSize(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	vld.SetJobDone("1", spos.SrBlock, true)
	assert.Equal(t, 1, vld.ComputeSize(spos.SrBlock))
}
