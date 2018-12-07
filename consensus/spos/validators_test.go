package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestNewValidators(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, 3, len(vld.ConsensusGroup()))
	assert.Equal(t, "3", vld.ConsensusGroup()[2])
	assert.Equal(t, "2", vld.SelfId())
}

func TestValidators_ResetValidationMap(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	vld.SetValidation("1", spos.SrBlock, true)
	assert.Equal(t, true, vld.GetValidation("1", spos.SrBlock))

	vld.ResetValidation()
	assert.Equal(t, false, vld.GetValidation("1", spos.SrBlock))
}

func TestValidators_IsNodeInBitmapGroup(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, false, vld.IsNodeInBitmapGroup(vld.SelfId()))
	vld.SetValidation(vld.SelfId(), spos.SrBitmap, true)
	assert.Equal(t, true, vld.IsNodeInBitmapGroup(vld.SelfId()))
}

func TestValidators_IsNodeInValidationGroup(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	assert.Equal(t, false, vld.IsNodeInValidationGroup("4"))
	assert.Equal(t, true, vld.IsNodeInValidationGroup(vld.SelfId()))
}

func TestValidators_IsBlockReceived(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.IsBlockReceived(1)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrBlock, true)
	assert.Equal(t, true, vld.GetValidation("1", spos.SrBlock))

	ok = vld.IsBlockReceived(1)
	assert.Equal(t, true, ok)

	ok = vld.IsBlockReceived(2)
	assert.Equal(t, false, ok)
}

func TestValidators_IsCommitmentHashReceived(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrCommitmentHash, true)
	assert.Equal(t, true, vld.GetValidation("1", spos.SrCommitmentHash))

	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("2", spos.SrCommitmentHash, true)
	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)

	vld.SetValidation("3", spos.SrCommitmentHash, true)
	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInCommitmentHash(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrBitmap, true)
	vld.SetValidation("3", spos.SrBitmap, true)
	assert.Equal(t, true, vld.GetValidation("3", spos.SrBitmap))

	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("2", spos.SrCommitmentHash, true)
	assert.Equal(t, true, vld.GetValidation("2", spos.SrCommitmentHash))

	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("3", spos.SrCommitmentHash, true)
	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrCommitmentHash, true)
	ok = vld.CommitmentHashesCollected(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInCommitment(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrBitmap, true)
	vld.SetValidation("3", spos.SrBitmap, true)
	assert.Equal(t, true, vld.GetValidation("3", spos.SrBitmap))

	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("2", spos.SrCommitment, true)
	assert.Equal(t, true, vld.GetValidation("2", spos.SrCommitment))

	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("3", spos.SrCommitment, true)
	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrCommitment, true)
	ok = vld.CommitmentsCollected(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInSignature(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	ok := vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrBitmap, true)
	vld.SetValidation("3", spos.SrBitmap, true)
	assert.Equal(t, true, vld.GetValidation("3", spos.SrBitmap))

	ok = vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("2", spos.SrSignature, true)
	assert.Equal(t, true, vld.GetValidation("2", spos.SrSignature))

	ok = vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("3", spos.SrSignature, true)
	ok = vld.SignaturesCollected(2)
	assert.Equal(t, false, ok)

	vld.SetValidation("1", spos.SrSignature, true)
	ok = vld.SignaturesCollected(2)
	assert.Equal(t, true, ok)
}

func TestValidators_ComputeSize(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	vld.SetValidation("1", spos.SrBlock, true)
	assert.Equal(t, 1, vld.ComputeSize(spos.SrBlock))
}
