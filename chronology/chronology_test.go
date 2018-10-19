package chronology_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

const (
	srStartRound chronology.Subround = iota
	srBlock
	srComitmentHash
	srBitmap
	srComitment
	srSignature
	srEndRound
)

const roundTimeDuration = time.Duration(10 * time.Millisecond)

// #################### <START_ROUND> ####################

type SRStartRound struct {
	Hits int
}

func (sr *SRStartRound) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoStartRound with %d hits\n", sr.Hits)
	return true
}

func (sr *SRStartRound) Current() chronology.Subround {
	return srStartRound
}

func (sr *SRStartRound) Next() chronology.Subround {
	return srBlock
}

func (sr *SRStartRound) EndTime() int64 {
	return int64(5 * roundTimeDuration / 100)
}

func (sr *SRStartRound) Name() string {
	return "<START_ROUND>"
}

// #################### <BLOCK> ####################

type SRBlock struct {
	Hits int
}

func (sr *SRBlock) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoBlock with %d hits\n", sr.Hits)
	return true
}

func (sr *SRBlock) Current() chronology.Subround {
	return srBlock
}

func (sr *SRBlock) Next() chronology.Subround {
	return srComitmentHash
}

func (sr *SRBlock) EndTime() int64 {
	return int64(25 * roundTimeDuration / 100)
}

func (sr *SRBlock) Name() string {
	return "<BLOCK>"
}

// #################### <COMITMENT_HASH> ####################

type SRComitmentHash struct {
	Hits int
}

func (sr *SRComitmentHash) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoComitmentHash with %d hits\n", sr.Hits)
	return true
}

func (sr *SRComitmentHash) Current() chronology.Subround {
	return srComitmentHash
}

func (sr *SRComitmentHash) Next() chronology.Subround {
	return srBitmap
}

func (sr *SRComitmentHash) EndTime() int64 {
	return int64(40 * roundTimeDuration / 100)
}

func (sr *SRComitmentHash) Name() string {
	return "<COMITMENT_HASH>"
}

// #################### <BITMAP> ####################

type SRBitmap struct {
	Hits int
}

func (sr *SRBitmap) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoBitmap with %d hits\n", sr.Hits)
	return true
}

func (sr *SRBitmap) Current() chronology.Subround {
	return srBitmap
}

func (sr *SRBitmap) Next() chronology.Subround {
	return srComitment
}

func (sr *SRBitmap) EndTime() int64 {
	return int64(55 * roundTimeDuration / 100)
}

func (sr *SRBitmap) Name() string {
	return "<BITMAP>"
}

// #################### <COMITMENT> ####################

type SRComitment struct {
	Hits int
}

func (sr *SRComitment) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoComitment with %d hits\n", sr.Hits)
	return true
}

func (sr *SRComitment) Current() chronology.Subround {
	return srComitment
}

func (sr *SRComitment) Next() chronology.Subround {
	return srSignature
}

func (sr *SRComitment) EndTime() int64 {
	return int64(70 * roundTimeDuration / 100)
}

func (sr *SRComitment) Name() string {
	return "<COMITMENT>"
}

// #################### <SIGNATURE> ####################

type SRSignature struct {
	Hits int
}

func (sr *SRSignature) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoSignature with %d hits\n", sr.Hits)
	return true
}

func (sr *SRSignature) Current() chronology.Subround {
	return srSignature
}

func (sr *SRSignature) Next() chronology.Subround {
	return srEndRound
}

func (sr *SRSignature) EndTime() int64 {
	return int64(85 * roundTimeDuration / 100)
}

func (sr *SRSignature) Name() string {
	return "<SIGNATURE>"
}

// #################### <END_ROUND> ####################

type SREndRound struct {
	Hits int
}

func (sr *SREndRound) DoWork(chr *chronology.Chronology) bool {
	sr.Hits++
	fmt.Printf("DoEndRound with %d hits\n", sr.Hits)
	return true
}

func (sr *SREndRound) Current() chronology.Subround {
	return srEndRound
}

func (sr *SREndRound) Next() chronology.Subround {
	return srStartRound
}

func (sr *SREndRound) EndTime() int64 {
	return int64(100 * roundTimeDuration / 100)
}

func (sr *SREndRound) Name() string {
	return "<END_ROUND>"
}

func TestStartRound(t *testing.T) {

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)
	syncTime := &ntp.LocalTime{}
	syncTime.SetClockOffset(roundTimeDuration + 1)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, syncTime)

	chr.AddSubround(&SRStartRound{})
	chr.AddSubround(&SRBlock{})
	chr.AddSubround(&SRComitmentHash{})
	chr.AddSubround(&SRBitmap{})
	chr.AddSubround(&SRComitment{})
	chr.AddSubround(&SRSignature{})
	chr.AddSubround(&SREndRound{})

	for {
		chr.StartRound()
		if len(chr.SubroundHandlers()) > 0 {
			if chr.SelfSubround() == chr.SubroundHandlers()[len(chr.SubroundHandlers())-1].Next() {
				break
			}
		}
	}
}

func TestRoundState(t *testing.T) {
	currentTime := time.Now()

	rnd := chronology.NewRound(currentTime, currentTime, roundTimeDuration)
	chr := chronology.NewChronology(true, true, rnd, currentTime, &ntp.LocalTime{})

	state := chr.GetSubroundFromDateTime(currentTime.Add(-1 * time.Hour))
	assert.Equal(t, chronology.SrBeforeRound, state)

	state = chr.GetSubroundFromDateTime(currentTime.Add(1 * time.Hour))
	assert.Equal(t, chronology.SrAfterRound, state)

	state = chr.GetSubroundFromDateTime(currentTime)
	assert.Equal(t, chronology.SrUnknown, state)

	chr.AddSubround(&SRStartRound{})

	state = chr.GetSubroundFromDateTime(currentTime)
	assert.NotEqual(t, chronology.SrUnknown, state)
}

func TestLoadSubrounder(t *testing.T) {
	chr := chronology.Chronology{}

	sr := chr.LoadSubroundHandler(chronology.SrBeforeRound)
	assert.Nil(t, sr)

	sr = chr.LoadSubroundHandler(chronology.SrAfterRound)
	assert.Nil(t, sr)

	chr.AddSubround(&SRStartRound{})

	sr = chr.LoadSubroundHandler(srStartRound)
	assert.NotNil(t, sr)

	assert.Equal(t, sr.Name(), chr.SubroundHandlers()[0].Name())
}

func TestGettersAndSetters(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)
	syncTime := &ntp.LocalTime{}

	chr := chronology.NewChronology(true, true, rnd, genesisTime, syncTime)

	assert.Equal(t, 0, chr.Round().Index())
	assert.Equal(t, chronology.SrBeforeRound, chr.SelfSubround())

	chr.SetSelfSubround(srStartRound)
	assert.Equal(t, srStartRound, chr.SelfSubround())

	assert.Equal(t, chronology.SrBeforeRound, chr.TimeSubround())
	assert.Equal(t, time.Duration(0), chr.ClockOffset())
	assert.NotNil(t, chr.SyncTime())
	assert.Equal(t, time.Duration(0), chr.SyncTime().ClockOffset())

	chr.SetClockOffset(time.Duration(5))
	assert.Equal(t, time.Duration(5), chr.ClockOffset())

	fmt.Printf("%v\n%v\n%v", chr.SyncTime().CurrentTime(chr.ClockOffset()), chr.SyncTime().FormatedCurrentTime(chr.ClockOffset()), chr.SubroundHandlers())
}
