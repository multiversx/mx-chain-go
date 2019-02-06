package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestNewRoundStatus(t *testing.T) {

	rstatus := spos.NewRoundStatus()

	assert.NotNil(t, rstatus)

	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.Equal(t, spos.SsFinished, rstatus.Status(bn.SrCommitmentHash))

	rstatus.SetStatus(bn.SrBitmap, spos.SsExtended)
	assert.Equal(t, spos.SsExtended, rstatus.Status(bn.SrBitmap))

	rstatus.ResetRoundStatus()
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrBitmap))
}
