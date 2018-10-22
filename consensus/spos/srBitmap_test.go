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

func InitSRBitmap() (*chronology.Chronology, *spos.SRBitmap) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := spos.NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := spos.NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := spos.NewRoundStatus(spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished)

	cns := spos.NewConsensus(true, vld, th, rs, chr)

	sr := spos.NewSRBitmap(true, int64(100*spos.RoundTimeDuration/100), cns, nil)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRBitmap(t *testing.T) {
	sr := spos.NewSRBitmap(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.NotNil(t, sr)
}

func TestSRBitmap_DoWork1(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("1: Test case when send message is done but consensus is not done -> RNone\n")

	sr.OnSendBitmap = SndWithSuccess

	r := sr.DoBitmap(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, spos.RNone, r)
}

func TestSRBitmap_DoWork2(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("2: Test case when send message and consensus is done -> RTrue\n")

	sr.OnSendBitmap = SndWithSuccess

	sr.Cns.RoundStatus.Block = spos.SsFinished

	for i := 0; i < len(sr.Cns.ConsensusGroup); i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.Cns.Threshold.Bitmap; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.DoBitmap(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, spos.RTrue, r)
}

func TestSRBitmap_DoWork3(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("3: Test case when time has expired -> RTrue\n")

	sr.OnSendBitmap = SndWithoutSuccess

	chr.SetClockOffset(time.Duration(sr.EndTime() + 1))

	r := sr.DoBitmap(chr)

	assert.Equal(t, spos.SsExtended, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, spos.RTrue, r)
}

func TestSRBitmap_DoWork4(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("4: Test case when receive message is done but consensus is not done -> RNone\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(true)

	r := sr.DoBitmap(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, spos.RNone, r)
}

func TestSRBitmap_DoWork5(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("5: Test case when receive message and consensus is done, and I WAS selected in leader's bitmap -> true\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(true)

	sr.Cns.RoundStatus.Block = spos.SsFinished

	for i := 0; i < len(sr.Cns.ConsensusGroup); i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.Cns.Threshold.Bitmap; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, true, r)
}

func TestSRBitmap_DoWork6(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("6: Test case when receive message and consensus is done, and I WAS NOT selected in leader's bitmap -> true\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(true)

	sr.Cns.RoundStatus.Block = spos.SsFinished

	for i := 0; i < len(sr.Cns.ConsensusGroup); i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.Cns.Threshold.Bitmap+1; i++ {
		if sr.Cns.ConsensusGroup[i] == sr.Cns.Self {
			continue
		}
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, true, r)
}

func TestSRBitmap_DoWork7(t *testing.T) {

	chr, sr := InitSRBitmap()

	fmt.Printf("7: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnSendBitmap = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(false)
	sr.Cns.Chr.SetSelfSubround(chronology.SrCanceled)

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSRBitmap_Current(t *testing.T) {
	sr := spos.NewSRBitmap(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrBitmap), sr.Current())
}

func TestSRBitmap_Next(t *testing.T) {
	sr := spos.NewSRBitmap(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrComitment), sr.Next())
}

func TestSRBitmap_EndTime(t *testing.T) {
	sr := spos.NewSRBitmap(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, int64(100*spos.RoundTimeDuration/100), sr.EndTime())
}

func TestSRBitmap_Name(t *testing.T) {
	sr := spos.NewSRBitmap(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, "<BITMAP>", sr.Name())
}

func TestSRBitmap_Log(t *testing.T) {
	sr := spos.NewSRBitmap(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	sr.Log("Test SRBitmap")
}
