package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestWorker_DoAdvanceJobShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrEndRound, spos.SsFinished)
	assert.False(t, wrk.DoAdvanceJob())
}

func TestWorker_DoAdvanceJobShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.DoAdvanceJob())
}
