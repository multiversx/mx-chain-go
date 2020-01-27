package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestRoundThreshold_NewThresholdShouldWork(t *testing.T) {
	t.Parallel()

	rthr := spos.NewRoundThreshold()

	assert.NotNil(t, rthr)
}
