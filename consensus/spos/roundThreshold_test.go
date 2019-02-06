package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestNewThreshold(t *testing.T) {
	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, 2)
	rthr.SetThreshold(bn.SrBitmap, 3)
	rthr.SetThreshold(bn.SrCommitment, 4)
	rthr.SetThreshold(bn.SrSignature, 5)

	assert.Equal(t, 3, rthr.Threshold(bn.SrBitmap))
}
