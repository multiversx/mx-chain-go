package spos_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
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

	rstatus.SetStatus(bls.SrSignature, spos.SsFinished)
	assert.Equal(t, spos.SsFinished, rstatus.Status(bls.SrSignature))
}

func TestRoundStatus_ResetRoundStatusShouldWork(t *testing.T) {
	t.Parallel()

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bls.SrStartRound, spos.SsFinished)
	rstatus.SetStatus(bls.SrBlock, spos.SsFinished)
	rstatus.SetStatus(bls.SrSignature, spos.SsFinished)
	rstatus.SetStatus(bls.SrEndRound, spos.SsFinished)

	rstatus.ResetRoundStatus()

	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bls.SrStartRound))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bls.SrBlock))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bls.SrSignature))
	assert.Equal(t, spos.SsNotFinished, rstatus.Status(bls.SrEndRound))
}
