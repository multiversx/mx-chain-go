package chronology

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

const (
	SR_START_ROUND Subround = iota
	SR_BLOCK
	SR_COMITMENT_HASH
	SR_BITMAP
	SR_COMITMENT
	SR_SIGNATURE
	SR_END_ROUND
)

const ROUND_TIME_DURATION = time.Duration(10 * time.Millisecond)

var GENESIS_TIME = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)

// #################### <START_ROUND> ####################

type SRStartRound struct {
	Hits int
}

func (sr *SRStartRound) DoWork(chr *Chronology) bool {
	sr.Hits++
	fmt.Printf("DoStartRound with %d hits\n", sr.Hits)
	return true
}

func (sr *SRStartRound) Current() Subround {
	return SR_START_ROUND
}

func (sr *SRStartRound) Next() Subround {
	return SR_BLOCK
}

func (sr *SRStartRound) EndTime() int64 {
	return int64(5 * ROUND_TIME_DURATION / 100)
}

func (sr *SRStartRound) Name() string {
	return "<START_ROUND>"
}

// #################### <BLOCK> ####################

type SRBlock struct {
	Hits int
}

func (sr *SRBlock) DoWork(chr *Chronology) bool {
	sr.Hits++
	fmt.Printf("DoBlock with %d hits\n", sr.Hits)
	return true
}

func (sr *SRBlock) Current() Subround {
	return SR_BLOCK
}

func (sr *SRBlock) Next() Subround {
	return SR_COMITMENT_HASH
}

func (sr *SRBlock) EndTime() int64 {
	return int64(25 * ROUND_TIME_DURATION / 100)
}

func (sr *SRBlock) Name() string {
	return "<BLOCK>"
}

// #################### <COMITMENT_HASH> ####################

type SRComitmentHash struct {
	Hits int
}

func (sr *SRComitmentHash) DoWork(chr *Chronology) bool {
	sr.Hits++
	fmt.Printf("DoComitmentHash with %d hits\n", sr.Hits)
	return true
}

func (sr *SRComitmentHash) Current() Subround {
	return SR_COMITMENT_HASH
}

func (sr *SRComitmentHash) Next() Subround {
	return SR_BITMAP
}

func (sr *SRComitmentHash) EndTime() int64 {
	return int64(40 * ROUND_TIME_DURATION / 100)
}

func (sr *SRComitmentHash) Name() string {
	return "<COMITMENT_HASH>"
}

// #################### <BITMAP> ####################

type SRBitmap struct {
	Hits int
}

func (sr *SRBitmap) DoWork(chr *Chronology) bool {
	sr.Hits++
	fmt.Printf("DoBitmap with %d hits\n", sr.Hits)
	return true
}

func (sr *SRBitmap) Current() Subround {
	return SR_BITMAP
}

func (sr *SRBitmap) Next() Subround {
	return SR_COMITMENT
}

func (sr *SRBitmap) EndTime() int64 {
	return int64(55 * ROUND_TIME_DURATION / 100)
}

func (sr *SRBitmap) Name() string {
	return "<BITMAP>"
}

// #################### <COMITMENT> ####################

type SRComitment struct {
	Hits int
}

func (sr *SRComitment) DoWork(chr *Chronology) bool {
	sr.Hits++
	fmt.Printf("DoComitment with %d hits\n", sr.Hits)
	return true
}

func (sr *SRComitment) Current() Subround {
	return SR_COMITMENT
}

func (sr *SRComitment) Next() Subround {
	return SR_SIGNATURE
}

func (sr *SRComitment) EndTime() int64 {
	return int64(70 * ROUND_TIME_DURATION / 100)
}

func (sr *SRComitment) Name() string {
	return "<COMITMENT>"
}

// #################### <SIGNATURE> ####################

type SRSignature struct {
	Hits int
}

func (sr *SRSignature) DoWork(chr *Chronology) bool {
	sr.Hits++
	fmt.Printf("DoSignature with %d hits\n", sr.Hits)
	return true
}

func (sr *SRSignature) Current() Subround {
	return SR_SIGNATURE
}

func (sr *SRSignature) Next() Subround {
	return SR_END_ROUND
}

func (sr *SRSignature) EndTime() int64 {
	return int64(85 * ROUND_TIME_DURATION / 100)
}

func (sr *SRSignature) Name() string {
	return "<SIGNATURE>"
}

// #################### <END_ROUND> ####################

type SREndRound struct {
	Hits int
}

func (sr *SREndRound) DoWork(chr *Chronology) bool {
	sr.Hits++
	fmt.Printf("DoEndRound with %d hits\n", sr.Hits)
	return true
}

func (sr *SREndRound) Current() Subround {
	return SR_END_ROUND
}

func (sr *SREndRound) Next() Subround {
	return SR_START_ROUND
}

func (sr *SREndRound) EndTime() int64 {
	return int64(100 * ROUND_TIME_DURATION / 100)
}

func (sr *SREndRound) Name() string {
	return "<END_ROUND>"
}

func TestStartRound(t *testing.T) {

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := NewRound(genesisTime, currentTime, ROUND_TIME_DURATION)
	syncTime := &ntp.LocalTime{ClockOffset: ROUND_TIME_DURATION + 1}

	chr := NewChronology(true, true, &rnd, genesisTime, syncTime)

	chr.AddSubrounder(&SRStartRound{})
	chr.AddSubrounder(&SRBlock{})
	chr.AddSubrounder(&SRComitmentHash{})
	chr.AddSubrounder(&SRBitmap{})
	chr.AddSubrounder(&SRComitment{})
	chr.AddSubrounder(&SRSignature{})
	chr.AddSubrounder(&SREndRound{})

	for {
		chr.startRound()
		if len(chr.subrounders) > 0 {
			if chr.selfSubround == chr.subrounders[len(chr.subrounders)-1].Next() {
				break
			}
		}
	}
}

func TestRoundState(t *testing.T) {
	currentTime := time.Now()

	rnd := NewRound(currentTime, currentTime, ROUND_TIME_DURATION)
	chr := Chronology{round: &rnd}

	state := chr.GetSubroundFromDateTime(currentTime.Add(-1 * time.Hour))
	assert.Equal(t, SR_BEFORE_ROUND, state)

	state = chr.GetSubroundFromDateTime(currentTime.Add(1 * time.Hour))
	assert.Equal(t, SR_AFTER_ROUND, state)

	state = chr.GetSubroundFromDateTime(currentTime)
	assert.Equal(t, SR_UNKNOWN, state)

	chr.AddSubrounder(&SRStartRound{})

	state = chr.GetSubroundFromDateTime(currentTime)
	assert.NotEqual(t, SR_UNKNOWN, state)
}

func TestLoadSubrounder(t *testing.T) {
	chr := Chronology{}

	sr := chr.loadSubrounder(SR_BEFORE_ROUND)
	assert.Nil(t, sr)

	sr = chr.loadSubrounder(SR_AFTER_ROUND)
	assert.Nil(t, sr)

	chr.AddSubrounder(&SRStartRound{})

	sr = chr.loadSubrounder(SR_START_ROUND)
	assert.NotNil(t, sr)

	assert.Equal(t, sr.Name(), chr.subrounders[0].Name())
}

func TestGetters(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := NewRound(genesisTime, currentTime, ROUND_TIME_DURATION)
	syncTime := &ntp.LocalTime{ClockOffset: 0}

	chr := NewChronology(true, true, &rnd, genesisTime, syncTime)

	assert.Equal(t, 0, chr.GetRoundIndex())
	assert.Equal(t, SR_BEFORE_ROUND, chr.GetSelfSubround())

	chr.SetSelfSubround(SR_START_ROUND)
	assert.Equal(t, SR_START_ROUND, chr.GetSelfSubround())

	assert.Equal(t, SR_BEFORE_ROUND, chr.GetTimeSubround())
	assert.Equal(t, time.Duration(0), chr.GetClockOffset())
	assert.NotNil(t, chr.GetSyncTimer())
	assert.Equal(t, time.Duration(0), chr.GetSyncTimer().GetClockOffset())

	chr.SetClockOffset(time.Duration(5))
	assert.Equal(t, time.Duration(5), chr.GetClockOffset())

	fmt.Printf("%v\n%v", chr.GetCurrentTime(), chr.GetFormatedCurrentTime())
}
