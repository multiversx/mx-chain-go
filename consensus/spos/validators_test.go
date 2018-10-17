package spos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidators(t *testing.T) {

	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	assert.Equal(t, 3, len(vld.ConsensusGroup))
	assert.Equal(t, "3", vld.ConsensusGroup[2])
	assert.Equal(t, "2", vld.Self)

	vld.ValidationMap["1"].Block = true

	assert.Equal(t, true, vld.ValidationMap["1"].Block)

	vld.ResetValidationMap()

	assert.Equal(t, false, vld.ValidationMap["1"].Block)
}

func TestValidators_IsNodeInBitmapGroup(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	assert.Equal(t, false, vld.IsNodeInBitmapGroup(vld.Self))

	vld.ValidationMap[vld.Self].Bitmap = true

	assert.Equal(t, true, vld.IsNodeInBitmapGroup(vld.Self))
}

func TestValidators_IsNodeInValidationGroup(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	assert.Equal(t, false, vld.IsNodeInValidationGroup("4"))
	assert.Equal(t, true, vld.IsNodeInValidationGroup(vld.Self))
}

func TestValidators_IsBlockReceived(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	ok, n := vld.IsBlockReceived(1)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].Block = true

	ok, n = vld.IsBlockReceived(1)
	assert.Equal(t, true, ok)
	assert.Equal(t, 1, n)

	ok, n = vld.IsBlockReceived(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 1, n)
}

func TestValidators_IsComitmentHashReceived(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	ok, n := vld.IsComitmentHashReceived(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].ComitmentHash = true

	ok, n = vld.IsComitmentHashReceived(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 1, n)

	vld.ValidationMap["2"].ComitmentHash = true

	ok, n = vld.IsComitmentHashReceived(2)
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, n)

	vld.ValidationMap["3"].ComitmentHash = true

	ok, n = vld.IsComitmentHashReceived(2)
	assert.Equal(t, true, ok)
	assert.Equal(t, 3, n)
}

func TestValidators_IsBitmapInComitmentHash(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	ok, n := vld.IsBitmapInComitmentHash(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].Bitmap = true
	vld.ValidationMap["3"].Bitmap = true

	ok, n = vld.IsBitmapInComitmentHash(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["2"].ComitmentHash = true

	ok, n = vld.IsBitmapInComitmentHash(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["3"].ComitmentHash = true

	ok, n = vld.IsBitmapInComitmentHash(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].ComitmentHash = true

	ok, n = vld.IsBitmapInComitmentHash(2)
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, n)
}

func TestValidators_IsBitmapInComitment(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	ok, n := vld.IsBitmapInComitment(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].Bitmap = true
	vld.ValidationMap["3"].Bitmap = true

	ok, n = vld.IsBitmapInComitment(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["2"].Comitment = true

	ok, n = vld.IsBitmapInComitment(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["3"].Comitment = true

	ok, n = vld.IsBitmapInComitment(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].Comitment = true

	ok, n = vld.IsBitmapInComitment(2)
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, n)
}

func TestValidators_IsBitmapInSignature(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	ok, n := vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].Bitmap = true
	vld.ValidationMap["3"].Bitmap = true

	ok, n = vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["2"].Signature = true

	ok, n = vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["3"].Signature = true

	ok, n = vld.IsBitmapInSignature(2)
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, n)

	vld.ValidationMap["1"].Signature = true

	ok, n = vld.IsBitmapInSignature(2)
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, n)
}

func TestValidators_GetComitmentHashesCount(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	vld.ValidationMap["1"].ComitmentHash = true
	assert.Equal(t, 1, vld.GetComitmentHashesCount())
}

func TestValidators_GetComitmentsCount(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	vld.ValidationMap["1"].Comitment = true
	vld.ValidationMap["2"].Comitment = true
	vld.ValidationMap["3"].Comitment = true

	assert.Equal(t, 3, vld.GetComitmentsCount())
}

func TestValidators_GetSignaturesCount(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	assert.Equal(t, 0, vld.GetComitmentsCount())
}
