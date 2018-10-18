package spos

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

func InitSRBlock() (*chronology.Chronology, *SRBlock) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := NewRoundStatus(SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished)

	cns := NewConsensus(true, vld, th, rs, chr)

	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), cns, nil, nil)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRBlock(t *testing.T) {
	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.NotNil(t, sr)
}

func TestSRBlock_DoWork1(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("1: Test case when send message is done but consensus is not done -> rNone\n")

	sr.OnSendBlock = SndWithSuccess

	r := sr.doBlock(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, rNone, r)
}

func TestSRBlock_DoWork2(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("2: Test case when send message and consensus is done -> rTrue\n")

	sr.OnSendBlock = SndWithSuccess

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	r := sr.doBlock(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, rTrue, r)
}

func TestSRBlock_DoWork3(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("3: Test case when time has expired -> rTrue\n")

	sr.OnSendBlock = SndWithoutSuccess

	chr.SetClockOffset(time.Duration(sr.endTime + 1))

	r := sr.doBlock(chr)

	assert.Equal(t, SsExtended, sr.cns.RoundStatus.Block)
	assert.Equal(t, rTrue, r)
}

func TestSRBlock_DoWork4(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("4: Test case when receive message is done but consensus is not done -> rNone\n")

	sr.OnSendBlock = SndWithoutSuccess
	sr.OnReceivedBlock = RcvWithSuccess

	sr.cns.SetReceivedMessage(true)

	r := sr.doBlock(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, rNone, r)
}

func TestSRBlock_DoWork5(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("5: Test case when receive message and consensus is done -> true\n")

	sr.OnSendBlock = SndWithoutSuccess
	sr.OnReceivedBlock = RcvWithSuccess

	sr.cns.SetReceivedMessage(true)

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	r := sr.DoWork(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, true, r)
}

func TestSRBlock_DoWork6(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("6: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnSendBlock = SndWithoutSuccess
	sr.OnReceivedBlock = RcvWithoutSuccessAndCancel

	sr.cns.SetReceivedMessage(false)
	sr.cns.Chr.SetSelfSubround(chronology.SrCanceled)

	r := sr.DoWork(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSRBlock_Current(t *testing.T) {
	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrBlock), sr.Current())
}

func TestSRBlock_Next(t *testing.T) {
	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrComitmentHash), sr.Next())
}

func TestSRBlock_EndTime(t *testing.T) {
	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, int64(100*roundTimeDuration/100), sr.EndTime())
}

func TestSRBlock_Name(t *testing.T) {
	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, "<BLOCK>", sr.Name())
}

func TestSRBlock_Log(t *testing.T) {
	sr := NewSRBlock(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	sr.Log("Test SRBlock")
}
