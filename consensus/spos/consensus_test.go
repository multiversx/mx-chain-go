package spos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsensus(t *testing.T) {

	cns := NewConsensus(true, Validators{}, Threshold{}, RoundStatus{})
	cns.ResetRoundStatus()

	assert.Equal(t, cns.RoundStatus.Block, SS_NOTFINISHED)
}
