package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestRoundThreshold_NewThresholdShouldWork(t *testing.T) {
	t.Parallel()

	rthr := spos.NewRoundThreshold()

	assert.NotNil(t, rthr)
}

func TestRoundThreshold_SetThresholdShouldWork(t *testing.T) {
	t.Parallel()

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, 2)
	rthr.SetThreshold(bn.SrBitmap, 3)
	rthr.SetThreshold(bn.SrCommitment, 4)
	rthr.SetThreshold(bn.SrSignature, 5)

	assert.Equal(t, 3, rthr.Threshold(bn.SrBitmap))
}
