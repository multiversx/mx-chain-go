package spos

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func TestStartRound(t *testing.T) {

	sr := NewSRStartRound(true, int64(5*ROUND_TIME_DURATION/100), nil)

	vld := NewValidators([]string{"1", "2", "3"}, "2")
	th := NewThreshold(1, 2*len(vld.ConsensusGroup)/3, 2*len(vld.ConsensusGroup)/3, 2*len(vld.ConsensusGroup)/3, 2*len(vld.ConsensusGroup)/3)
	rs := NewRoundStatus(SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED)

	cns := NewConsensus(true, vld, th, rs)

	bl := block.NewBlock(-1, "", "", "", "", "")
	cns.Block = &bl
	sr.cns = &cns

	assert.Equal(t, sr.EndTime(), int64(5*ROUND_TIME_DURATION/100))
	assert.Equal(t, sr.Current(), chronology.Subround(SR_START_ROUND))
	assert.Equal(t, sr.Next(), chronology.Subround(SR_BLOCK))
	assert.Equal(t, sr.Name(), "<START_ROUND>")

	genesisTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
	currentTime := time.Now()
	//	genesisTime := time.Now()
	//	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, ROUND_TIME_DURATION)
	syncTime := &ntp.LocalTime{0}

	chr := chronology.NewChronology(true, true, &rnd, genesisTime, syncTime)
	sr.cns.chr = &chr

	sr.DoWork()

	chr.SetSelfSubRound(chronology.SR_ABORDED)

	sr.DoWork()
}
