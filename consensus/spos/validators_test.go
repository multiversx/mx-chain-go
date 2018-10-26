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

	assert.Equal(t, 3, len(vld.ConsensusGroup()))
	assert.Equal(t, "3", vld.ConsensusGroup()[2])
	assert.Equal(t, "2", vld.Self())
}

func TestValidators_ResetValidationMap(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	vld.SetValidationMap("1", true, spos.SrBlock)
	assert.Equal(t, true, vld.ValidationMap("1", spos.SrBlock))

	vld.ResetValidationMap()
	assert.Equal(t, false, vld.ValidationMap("1", spos.SrBlock))
}

func TestValidators_IsNodeInBitmapGroup(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	assert.Equal(t, false, vld.IsNodeInBitmapGroup(vld.Self()))
	vld.SetValidationMap(vld.Self(), true, spos.SrBitmap)
	assert.Equal(t, true, vld.IsNodeInBitmapGroup(vld.Self()))
}

func TestValidators_IsNodeInValidationGroup(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	assert.Equal(t, false, vld.IsNodeInValidationGroup("4"))
	assert.Equal(t, true, vld.IsNodeInValidationGroup(vld.Self()))
}

func TestValidators_IsBlockReceived(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	ok := vld.IsBlockReceived(1)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrBlock)
	assert.Equal(t, true, vld.ValidationMap("1", spos.SrBlock))

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

	ok := vld.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrCommitmentHash)
	assert.Equal(t, true, vld.ValidationMap("1", spos.SrCommitmentHash))

	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("2", true, spos.SrCommitmentHash)
	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)

	vld.SetValidationMap("3", true, spos.SrCommitmentHash)
	ok = vld.IsCommitmentHashReceived(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInCommitmentHash(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	ok := vld.IsBitmapInCommitmentHash(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrBitmap)
	vld.SetValidationMap("3", true, spos.SrBitmap)
	assert.Equal(t, true, vld.ValidationMap("3", spos.SrBitmap))

	ok = vld.IsBitmapInCommitmentHash(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("2", true, spos.SrCommitmentHash)
	assert.Equal(t, true, vld.ValidationMap("2", spos.SrCommitmentHash))

	ok = vld.IsBitmapInCommitmentHash(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("3", true, spos.SrCommitmentHash)
	ok = vld.IsBitmapInCommitmentHash(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrCommitmentHash)
	ok = vld.IsBitmapInCommitmentHash(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInCommitment(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	ok := vld.IsBitmapInCommitment(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrBitmap)
	vld.SetValidationMap("3", true, spos.SrBitmap)
	assert.Equal(t, true, vld.ValidationMap("3", spos.SrBitmap))

	ok = vld.IsBitmapInCommitment(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("2", true, spos.SrCommitment)
	assert.Equal(t, true, vld.ValidationMap("2", spos.SrCommitment))

	ok = vld.IsBitmapInCommitment(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("3", true, spos.SrCommitment)
	ok = vld.IsBitmapInCommitment(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrCommitment)
	ok = vld.IsBitmapInCommitment(2)
	assert.Equal(t, true, ok)
}

func TestValidators_IsBitmapInSignature(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	ok := vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrBitmap)
	vld.SetValidationMap("3", true, spos.SrBitmap)
	assert.Equal(t, true, vld.ValidationMap("3", spos.SrBitmap))

	ok = vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("2", true, spos.SrSignature)
	assert.Equal(t, true, vld.ValidationMap("2", spos.SrSignature))

	ok = vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("3", true, spos.SrSignature)
	ok = vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)

	vld.SetValidationMap("1", true, spos.SrSignature)
	ok = vld.IsBitmapInSignature(2)
	assert.Equal(t, true, ok)
}

func TestValidators_GetBlocksCount(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	vld.SetValidationMap("1", true, spos.SrBlock)
	assert.Equal(t, 1, vld.GetBlocksCount())
}

func TestValidators_GetCommitmentHashesCount(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	vld.SetValidationMap("1", true, spos.SrCommitmentHash)
	assert.Equal(t, 1, vld.GetCommitmentHashesCount())
}

func TestValidators_GetBitmapsCount(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	vld.SetValidationMap("1", true, spos.SrBitmap)
	vld.SetValidationMap("2", true, spos.SrBitmap)
	assert.Equal(t, 2, vld.GetBitmapsCount())
}

func TestValidators_GetCommitmentsCount(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	vld.SetValidationMap("1", true, spos.SrCommitment)
	vld.SetValidationMap("2", true, spos.SrCommitment)
	vld.SetValidationMap("3", true, spos.SrCommitment)
	assert.Equal(t, 3, vld.GetCommitmentsCount())
}

func TestValidators_GetSignaturesCount(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	assert.Equal(t, 0, vld.GetCommitmentsCount())
}
