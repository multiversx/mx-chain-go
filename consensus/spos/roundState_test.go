package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestRoundState_NewRoundStateShouldWork(t *testing.T) {
	t.Parallel()

	rstate := spos.NewRoundState()
	assert.NotNil(t, rstate)
}

func TestRoundState_SetJobDoneShouldWork(t *testing.T) {
	t.Parallel()

	rstate := spos.NewRoundState()

	rstate.SetJobDone(1, true)

	assert.True(t, rstate.JobDone(1))
}

func TestRoundState_ResetJobDoneShouldWork(t *testing.T) {
	t.Parallel()

	rstate := spos.NewRoundState()

	rstate.SetJobDone(1, true)
	rstate.SetJobDone(2, true)
	rstate.SetJobDone(3, true)
	rstate.SetJobDone(4, true)
	rstate.SetJobDone(5, true)

	rstate.ResetJobsDone()

	assert.False(t, rstate.JobDone(1))
	assert.False(t, rstate.JobDone(2))
	assert.False(t, rstate.JobDone(3))
	assert.False(t, rstate.JobDone(4))
	assert.False(t, rstate.JobDone(5))
}
