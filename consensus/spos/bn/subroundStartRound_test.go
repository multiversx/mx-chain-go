package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestWorker_StartRound(t *testing.T) {
	wrk := initWorker()

	r := wrk.DoStartRoundJob()
	assert.True(t, r)
}

func TestWorker_CheckStartRoundConsensus(t *testing.T) {
	wrk := initWorker()

	ok := wrk.CheckStartRoundConsensus()
	assert.True(t, ok)
}

func TestWorker_DoExtendStartRoundShouldSetStartRoundFinished(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendStartRound()

	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrStartRound))
}
