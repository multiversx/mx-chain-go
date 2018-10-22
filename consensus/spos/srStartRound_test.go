package spos_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"

	"github.com/stretchr/testify/assert"
)

func InitSRStartRound() (*chronology.Chronology, *spos.SRStartRound) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := spos.NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := spos.NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := spos.NewRoundStatus(spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished)

	cns := spos.NewConsensus(true, vld, th, rs, chr)
	//Cns.block = block.NewBlock(-1, "", "", "", "", "")

	sr := spos.NewSRStartRound(true, int64(100*spos.RoundTimeDuration/100), cns, OnStartRound)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRStartRound(t *testing.T) {
	sr := spos.NewSRStartRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.NotNil(t, sr)
}

func TestStartRound_DoWork1(t *testing.T) {

	chr, sr := InitSRStartRound()

	fmt.Printf("1: Test case when consensus group is empty -> RNone\n")

	sr.Cns.Validators.ConsensusGroup = nil

	r := sr.DoStartRound(chr)

	assert.Equal(t, spos.RNone, r)
}

func TestStartRound_DoWork2(t *testing.T) {

	chr, sr := InitSRStartRound()

	fmt.Printf("2: Test case when I am the leader -> true\n")

	sr.Cns.Self = "1"

	r := sr.DoWork(chr)

	assert.Equal(t, true, r)
}

func TestStartRound_DoWork3(t *testing.T) {

	chr, sr := InitSRStartRound()

	fmt.Printf("3: Test case when I am NOT the leader -> true\n")

	r := sr.DoWork(chr)

	assert.Equal(t, true, r)
}

func TestStartRound_DoWork4(t *testing.T) {

	chr, sr := InitSRStartRound()

	fmt.Printf("4: Test case when subround is canceled -> false\n")

	chr.SetSelfSubround(chronology.Subround(chronology.SrCanceled))

	r := sr.DoWork(chr)

	assert.Equal(t, false, r)
}

func TestSRStartRound_Current(t *testing.T) {
	sr := spos.NewSRStartRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrStartRound), sr.Current())
}

func TestSRStartRound_Next(t *testing.T) {
	sr := spos.NewSRStartRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrBlock), sr.Next())
}

func TestSRStartRound_EndTime(t *testing.T) {
	sr := spos.NewSRStartRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, int64(100*spos.RoundTimeDuration/100), sr.EndTime())
}

func TestSRStartRound_Name(t *testing.T) {
	sr := spos.NewSRStartRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, "<START_ROUND>", sr.Name())
}

func TestSRStartRound_Log(t *testing.T) {
	sr := spos.NewSRStartRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	sr.Log("Test SRStartRound")
}
