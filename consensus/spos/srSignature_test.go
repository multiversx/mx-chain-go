package spos

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

func InitSRSignature() (*chronology.Chronology, *SRSignature) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := NewRoundStatus(SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished)

	cns := NewConsensus(true, vld, th, rs, chr)

	sr := NewSRSignature(true, int64(100*roundTimeDuration/100), cns, nil, nil)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRSignature(t *testing.T) {
	sr := NewSRSignature(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.NotNil(t, sr)
}

func TestSRSignature_DoWork1(t *testing.T) {

	chr, sr := InitSRSignature()

	fmt.Printf("1: Test case when send message is done but consensus is not done -> rNone\n")

	sr.OnSendSignature = SndWithSuccess

	r := sr.doSignature(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Comitment)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rNone, r)
}

func TestSRSignature_DoWork2(t *testing.T) {

	chr, sr := InitSRSignature()

	fmt.Printf("2: Test case when send message is done but consensus is semi-done (only subround BLOCK, COMITMENT_HASH, BITMAP and COMITMENT is done) -> rNone\n")

	sr.OnSendSignature = SndWithSuccess

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.cns.Threshold.Bitmap; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
	}

	for i := 0; i < sr.cns.Threshold.Comitment; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Comitment = true
	}

	r := sr.doSignature(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Comitment)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rNone, r)
}

func TestSRSignature_DoWork3(t *testing.T) {

	chr, sr := InitSRSignature()

	fmt.Printf("3: Test case when send message and consensus is done -> rTrue\n")

	sr.OnSendSignature = SndWithSuccess

	sr.cns.RoundStatus.Block = SsFinished
	sr.cns.RoundStatus.ComitmentHash = SsFinished
	sr.cns.RoundStatus.Bitmap = SsFinished
	sr.cns.RoundStatus.Comitment = SsFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Signature = true
	}

	r := sr.doSignature(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rTrue, r)
}

func TestSRSignature_DoWork4(t *testing.T) {

	chr, sr := InitSRSignature()

	fmt.Printf("4: Test case when time has expired -> rTrue\n")

	sr.OnSendSignature = SndWithoutSuccess

	chr.SetClockOffset(time.Duration(sr.endTime + 1))

	r := sr.doSignature(chr)

	assert.Equal(t, SsExtended, sr.cns.RoundStatus.Signature)
	assert.Equal(t, r, rTrue)
}

func TestSRSignature_DoWork5(t *testing.T) {

	chr, sr := InitSRSignature()

	fmt.Printf("5: Test case when receive message is done but consensus is not done -> rNone\n")

	sr.OnSendSignature = SndWithoutSuccess
	sr.OnReceivedSignature = RcvWithSuccess

	sr.cns.SetReceivedMessage(true)

	r := sr.doSignature(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Comitment)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rNone, r)
}

func TestSRSignature_DoWork6(t *testing.T) {

	chr, sr := InitSRSignature()

	fmt.Printf("6: Test case when receive message and consensus is done -> true\n")

	sr.OnSendSignature = SndWithoutSuccess
	sr.OnReceivedSignature = RcvWithSuccess

	sr.cns.SetReceivedMessage(true)

	sr.cns.RoundStatus.Block = SsFinished
	sr.cns.RoundStatus.ComitmentHash = SsFinished
	sr.cns.RoundStatus.Bitmap = SsFinished
	sr.cns.RoundStatus.Comitment = SsFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Signature = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, true, r)
}

func TestSRSignature_DoWork7(t *testing.T) {

	chr, sr := InitSRSignature()

	fmt.Printf("7: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnSendSignature = SndWithoutSuccess
	sr.OnReceivedSignature = RcvWithoutSuccessAndCancel

	sr.cns.SetReceivedMessage(false)
	sr.cns.Chr.SetSelfSubround(chronology.SrCanceled)

	sr.cns.RoundStatus.Block = SsFinished
	sr.cns.RoundStatus.ComitmentHash = SsFinished
	sr.cns.RoundStatus.Bitmap = SsFinished
	sr.cns.RoundStatus.Comitment = SsFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Signature = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSRSignature_Current(t *testing.T) {
	sr := NewSRSignature(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrSignature), sr.Current())
}

func TestSRSignature_Next(t *testing.T) {
	sr := NewSRSignature(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrEndRound), sr.Next())
}

func TestSRSignature_EndTime(t *testing.T) {
	sr := NewSRSignature(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, int64(100*roundTimeDuration/100), sr.EndTime())
}

func TestSRSignature_Name(t *testing.T) {
	sr := NewSRSignature(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, "<SIGNATURE>", sr.Name())
}

func TestSRSignature_Log(t *testing.T) {
	sr := NewSRSignature(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	sr.Log("Test SRSignature")
}
