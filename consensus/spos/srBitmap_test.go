package spos

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

func InitSRBitmap() (*chronology.Chronology, *SRBitmap) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := NewRoundStatus(SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished)

	cns := NewConsensus(true, vld, th, rs, chr)

	sr := NewSRBitmap(true, int64(100*roundTimeDuration/100), cns, nil, nil)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRBitmap(t *testing.T) {
	sr := NewSRBitmap(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.NotNil(t, sr)
}

func TestSRBitmap_DoWork1(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("1: Test case when send message is done but consensus is not done -> rNone\n")

	sr.OnSendBitmap = SndWithSuccess

	r := sr.doBitmap(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, rNone, r)
}

func TestSRBitmap_DoWork2(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("2: Test case when send message and consensus is done -> rTrue\n")

	sr.OnSendBitmap = SndWithSuccess

	sr.cns.RoundStatus.Block = SsFinished

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.cns.Threshold.Bitmap; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.doBitmap(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, rTrue, r)
}

func TestSRBitmap_DoWork3(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("3: Test case when time has expired -> rTrue\n")

	sr.OnSendBitmap = SndWithoutSuccess

	chr.SetClockOffset(time.Duration(sr.endTime + 1))

	r := sr.doBitmap(chr)

	assert.Equal(t, SsExtended, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, rTrue, r)
}

func TestSRBitmap_DoWork4(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("4: Test case when receive message is done but consensus is not done -> rNone\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.OnReceivedBitmap = RcvWithSuccess

	sr.cns.SetReceivedMessage(true)

	r := sr.doBitmap(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, rNone, r)
}

func TestSRBitmap_DoWork5(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("5: Test case when receive message and consensus is done, and I WAS selected in leader's bitmap -> true\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.OnReceivedBitmap = RcvWithSuccess

	sr.cns.SetReceivedMessage(true)

	sr.cns.RoundStatus.Block = SsFinished

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.cns.Threshold.Bitmap; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, true, r)
}

func TestSRBitmap_DoWork6(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("6: Test case when receive message and consensus is done, and I WAS NOT selected in leader's bitmap -> true\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.OnReceivedBitmap = RcvWithSuccess

	sr.cns.SetReceivedMessage(true)

	sr.cns.RoundStatus.Block = SsFinished

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.cns.Threshold.Bitmap+1; i++ {
		if sr.cns.ConsensusGroup[i] == sr.cns.Self {
			continue
		}
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, true, r)
}

func TestSRBitmap_DoWork7(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("7: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.OnReceivedBitmap = RcvWithoutSuccessAndCancel

	sr.cns.SetReceivedMessage(false)
	sr.cns.Chr.SetSelfSubround(chronology.SrCanceled)

	r := sr.DoWork(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSRBitmap_Current(t *testing.T) {
	sr := NewSRBitmap(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrBitmap), sr.Current())
}

func TestSRBitmap_Next(t *testing.T) {
	sr := NewSRBitmap(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrComitment), sr.Next())
}

func TestSRBitmap_EndTime(t *testing.T) {
	sr := NewSRBitmap(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, int64(100*roundTimeDuration/100), sr.EndTime())
}

func TestSRBitmap_Name(t *testing.T) {
	sr := NewSRBitmap(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, "<BITMAP>", sr.Name())
}

func TestSRBitmap_Log(t *testing.T) {
	sr := NewSRBitmap(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	sr.Log("Test SRBitmap")
}
