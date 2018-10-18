package spos

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func InitSRStartRound() (*chronology.Chronology, *SRStartRound) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := NewRoundStatus(SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished)

	cns := NewConsensus(true, vld, th, rs, chr)
	//cns.block = block.NewBlock(-1, "", "", "", "", "")

	sr := NewSRStartRound(true, int64(100*roundTimeDuration/100), cns, nil, nil)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRStartRound(t *testing.T) {
	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.NotNil(t, sr)
}

func TestStartRound_DoWork1(t *testing.T) {

	chr, sr := InitSRStartRound()

	fmt.Printf("1: Test case when consensus group is empty -> rNone\n")

	sr.cns.Validators.ConsensusGroup = nil

	r := sr.doStartRound(chr)

	assert.Equal(t, rNone, r)
}

func TestStartRound_DoWork2(t *testing.T) {

	chr, sr := InitSRStartRound()

	fmt.Printf("2: Test case when I am the leader -> true\n")

	sr.cns.Self = "1"

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
	sr := NewSRStartRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrStartRound), sr.Current())
}

func TestSRStartRound_Next(t *testing.T) {
	sr := NewSRStartRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrBlock), sr.Next())
}

func TestSRStartRound_EndTime(t *testing.T) {
	sr := NewSRStartRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, int64(100*roundTimeDuration/100), sr.EndTime())
}

func TestSRStartRound_Name(t *testing.T) {
	sr := NewSRStartRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, "<START_ROUND>", sr.Name())
}

func TestSRStartRound_Log(t *testing.T) {
	sr := NewSRStartRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	sr.Log("Test SRStartRound")
}
