package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestRoundStatus_NewRoundStatusShouldWork(t *testing.T) {
	t.Parallel()

	rstatus := spos.NewRoundStatus()
	assert.NotNil(t, rstatus)
}

func TestRoundStatus_SetRoundStatusShouldWork(t *testing.T) {
	t.Parallel()

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.Equal(t, spos.SsFinished, rstatus.Status(bn.SrCommitmentHash))
}

func TestRoundStatus_ResetRoundStatusShouldWork(t *testing.T) {
	t.Parallel()

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrStartRound, spos.SsFinished)
	rstatus.SetStatus(bn.SrBlock, spos.SsFinished)
	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	rstatus.SetStatus(bn.SrBitmap, spos.SsFinished)
	rstatus.SetStatus(bn.SrCommitment, spos.SsFinished)
	rstatus.SetStatus(bn.SrSignature, spos.SsFinished)
	rstatus.SetStatus(bn.SrEndRound, spos.SsFinished)

	rstatus.ResetRoundStatus()

	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrStartRound))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrBlock))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrCommitmentHash))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrBitmap))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrCommitment))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrSignature))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bn.SrEndRound))
}
