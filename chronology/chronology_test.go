package chronology

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

type SubRound int

const (
	SR_START_ROUND SubRound = iota
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
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoStartRound with %d hits\n", sr.Hits)
	return true
}

func (s *SRStartRound) Next() int {
	return int(SR_BLOCK)
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

func (sr *SRBlock) DoWork(chr *Chronology) bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoBlock with %d hits\n", sr.Hits)
	return true
}

func (s *SRBlock) Next() int {
	return int(SR_COMITMENT_HASH)
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

func (sr *SRComitmentHash) DoWork(chr *Chronology) bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoComitmentHash with %d hits\n", sr.Hits)
	return true
}

func (s *SRComitmentHash) Next() int {
	return int(SR_BITMAP)
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

func (sr *SRBitmap) DoWork(chr *Chronology) bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoBitmap with %d hits\n", sr.Hits)
	return true
}

func (s *SRBitmap) Next() int {
	return int(SR_COMITMENT)
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

func (sr *SRComitment) DoWork(chr *Chronology) bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoComitment with %d hits\n", sr.Hits)
	return true
}

func (s *SRComitment) Next() int {
	return int(SR_SIGNATURE)
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

func (sr *SRSignature) DoWork(chr *Chronology) bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoSignature with %d hits\n", sr.Hits)
	return true
}

func (s *SRSignature) Next() int {
	return int(SR_END_ROUND)
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

func (sr *SREndRound) DoWork(chr *Chronology) bool {
	time.Sleep(time.Millisecond * 1)
	sr.Hits++
	fmt.Printf("DoEndRound with %d hits\n", sr.Hits)
	return true
}

func (s *SREndRound) Next() int {
	return int(SR_START_ROUND)
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

	chr.AddSubRounder(&SRStartRound{})
	chr.AddSubRounder(&SRBlock{})
	chr.AddSubRounder(&SRComitmentHash{})
	chr.AddSubRounder(&SRBitmap{})
	chr.AddSubRounder(&SRComitment{})
	chr.AddSubRounder(&SRSignature{})
	chr.AddSubRounder(&SREndRound{})

	for {
		chr.startRound()
		if chr.selfRoundState == int(SR_START_ROUND) {
			break
		}
	}
}

func TestRoundState(t *testing.T) {
	currentTime := time.Now()

	rnd := NewRound(currentTime, currentTime, ROUND_TIME_DURATION)
	chr := Chronology{round: &rnd}

	state := chr.GetRoundStateFromDateTime(currentTime.Add(-1 * time.Hour))
	assert.Equal(t, RS_BEFORE_ROUND, state)

	state = chr.GetRoundStateFromDateTime(currentTime.Add(1 * time.Hour))
	assert.Equal(t, RS_AFTER_ROUND, state)

	state = chr.GetRoundStateFromDateTime(currentTime)
	assert.Equal(t, RS_UNKNOWN, state)

	chr.AddSubRounder(&SRStartRound{})

	state = chr.GetRoundStateFromDateTime(currentTime)
	assert.NotEqual(t, RS_UNKNOWN, state)
}

func TestLoadSubrounder(t *testing.T) {
	chr := Chronology{}

	sr := chr.LoadSubRounder(int(RS_BEFORE_ROUND))
	assert.Nil(t, sr)

	sr = chr.LoadSubRounder(int(RS_AFTER_ROUND))
	assert.Nil(t, sr)

	chr.AddSubRounder(&SRStartRound{})

	sr = chr.LoadSubRounder(int(SR_START_ROUND))
	assert.NotNil(t, sr)

	assert.Equal(t, sr.Name(), chr.subRounders[0].Name())
}
