package spos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidators(t *testing.T) {

	vld := NewValidators([]string{"1", "2", "3"}, "2")
	vld.ResetValidationMap()

	assert.Equal(t, len(vld.ConsensusGroup), 3)
	assert.Equal(t, vld.ConsensusGroup[2], "3")
	assert.Equal(t, vld.Self, "2")

	v := vld.ValidationMap["1"]
	v.Block = true
	vld.ValidationMap["1"] = v

	assert.Equal(t, vld.ValidationMap["1"].Block, true)
}
