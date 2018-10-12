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

func (sr *SRStartRound) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoStartRound with %d hits\n", sr.Hits)
	return true
}

func (s *SRStartRound) Current() Subround {
	return SR_START_ROUND
}

func (s *SRStartRound) Next() Subround {
	return SR_BLOCK
}

func (s *SRStartRound) EndTime() int64 {
	return int64(5 * ROUND_TIME_DURATION / 100)
}

func (s *SRStartRound) Name() string {
	return "<START_ROUND>"
}

// #################### <BLOCK> ####################

type SRBlock struct {
	Hits int
}

func (sr *SRBlock) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoBlock with %d hits\n", sr.Hits)
	return true
}

func (s *SRBlock) Current() Subround {
	return SR_BLOCK
}

func (s *SRBlock) Next() Subround {
	return SR_COMITMENT_HASH
}

func (s *SRBlock) EndTime() int64 {
	return int64(25 * ROUND_TIME_DURATION / 100)
}

func (s *SRBlock) Name() string {
	return "<BLOCK>"
}

// #################### <COMITMENT_HASH> ####################

type SRComitmentHash struct {
	Hits int
}

func (sr *SRComitmentHash) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoComitmentHash with %d hits\n", sr.Hits)
	return true
}

func (s *SRComitmentHash) Current() Subround {
	return SR_COMITMENT_HASH
}

func (s *SRComitmentHash) Next() Subround {
	return SR_BITMAP
}

func (s *SRComitmentHash) EndTime() int64 {
	return int64(40 * ROUND_TIME_DURATION / 100)
}

func (s *SRComitmentHash) Name() string {
	return "<COMITMENT_HASH>"
}

// #################### <BITMAP> ####################

type SRBitmap struct {
	Hits int
}

func (sr *SRBitmap) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoBitmap with %d hits\n", sr.Hits)
	return true
}

func (s *SRBitmap) Current() Subround {
	return SR_BITMAP
}

func (s *SRBitmap) Next() Subround {
	return SR_COMITMENT
}

func (s *SRBitmap) EndTime() int64 {
	return int64(55 * ROUND_TIME_DURATION / 100)
}

func (s *SRBitmap) Name() string {
	return "<BITMAP>"
}

// #################### <COMITMENT> ####################

type SRComitment struct {
	Hits int
}

func (sr *SRComitment) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoComitment with %d hits\n", sr.Hits)
	return true
}

func (s *SRComitment) Current() Subround {
	return SR_COMITMENT
}

func (s *SRComitment) Next() Subround {
	return SR_SIGNATURE
}

func (s *SRComitment) EndTime() int64 {
	return int64(70 * ROUND_TIME_DURATION / 100)
}

func (s *SRComitment) Name() string {
	return "<COMITMENT>"
}

// #################### <SIGNATURE> ####################

type SRSignature struct {
	Hits int
}

func (sr *SRSignature) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoSignature with %d hits\n", sr.Hits)
	return true
}

func (s *SRSignature) Current() Subround {
	return SR_SIGNATURE
}

func (s *SRSignature) Next() Subround {
	return SR_END_ROUND
}

func (s *SRSignature) EndTime() int64 {
	return int64(85 * ROUND_TIME_DURATION / 100)
}

func (s *SRSignature) Name() string {
	return "<SIGNATURE>"
}

// #################### <END_ROUND> ####################

type SREndRound struct {
	Hits int
}

func (sr *SREndRound) DoWork() bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoEndRound with %d hits\n", sr.Hits)
	return true
}

func (s *SREndRound) Current() Subround {
	return SR_END_ROUND
}

func (s *SREndRound) Next() Subround {
	return SR_START_ROUND
}

func (s *SREndRound) EndTime() int64 {
	return int64(100 * ROUND_TIME_DURATION / 100)
}

func (s *SREndRound) Name() string {
	return "<END_ROUND>"
}

func TestStartRound(t *testing.T) {

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := NewRound(genesisTime, currentTime, ROUND_TIME_DURATION)
	syncTime := &ntp.LocalTime{ROUND_TIME_DURATION + 1}

	chr := NewChronology(true, true, &rnd, genesisTime, syncTime)

	//chr.AddSubRounder(&spos.SRStartRound{true, 5 * ROUND_TIME_DURATION / 100})
	chr.AddSubRounder(&SRStartRound{})
	chr.AddSubRounder(&SRBlock{})
	chr.AddSubRounder(&SRComitmentHash{})
	chr.AddSubRounder(&SRBitmap{})
	chr.AddSubRounder(&SRComitment{})
	chr.AddSubRounder(&SRSignature{})
	chr.AddSubRounder(&SREndRound{})

	for {
		chr.startRound()
		if len(chr.subRounders) > 0 {
			if chr.selfSubRound == chr.subRounders[len(chr.subRounders)-1].Next() {
				break
			}
		}
	}
}

func TestRoundState(t *testing.T) {
	currentTime := time.Now()

	rnd := NewRound(currentTime, currentTime, ROUND_TIME_DURATION)
	chr := Chronology{round: &rnd}

	state := chr.GetSubRoundFromDateTime(currentTime.Add(-1 * time.Hour))
	assert.Equal(t, SR_BEFORE_ROUND, state)

	state = chr.GetSubRoundFromDateTime(currentTime.Add(1 * time.Hour))
	assert.Equal(t, SR_AFTER_ROUND, state)

	state = chr.GetSubRoundFromDateTime(currentTime)
	assert.Equal(t, SR_UNKNOWN, state)

	chr.AddSubRounder(&SRStartRound{})

	state = chr.GetSubRoundFromDateTime(currentTime)
	assert.NotEqual(t, SR_UNKNOWN, state)
}

func TestLoadSubrounder(t *testing.T) {
	chr := Chronology{}

	sr := chr.LoadSubRounder(SR_BEFORE_ROUND)
	assert.Nil(t, sr)

	sr = chr.LoadSubRounder(SR_AFTER_ROUND)
	assert.Nil(t, sr)

	chr.AddSubRounder(&SRStartRound{})

	sr = chr.LoadSubRounder(SR_START_ROUND)
	assert.NotNil(t, sr)

	assert.Equal(t, sr.Name(), chr.subRounders[0].Name())
}

func TestGetters(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := NewRound(genesisTime, currentTime, ROUND_TIME_DURATION)
	syncTime := &ntp.LocalTime{0}

	chr := NewChronology(true, true, &rnd, genesisTime, syncTime)

	assert.Equal(t, 0, chr.GetRoundIndex())
	assert.Equal(t, SR_BEFORE_ROUND, chr.GetSelfSubRound())

	chr.SetSelfSubRound(SR_START_ROUND)
	assert.Equal(t, SR_START_ROUND, chr.GetSelfSubRound())

	assert.Equal(t, SR_BEFORE_ROUND, chr.GetTimeSubRound())
	assert.Equal(t, time.Duration(0), chr.GetClockOffset())
	assert.NotNil(t, chr.GetSyncTimer())
	assert.Equal(t, time.Duration(0), chr.GetSyncTimer().GetClockOffset())

	fmt.Printf("%v\n%v", chr.GetCurrentTime(), chr.GetFormatedCurrentTime())
}
