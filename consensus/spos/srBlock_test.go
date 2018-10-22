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

func InitSRBlock() (*chronology.Chronology, *spos.SRBlock) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := spos.NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := spos.NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := spos.NewRoundStatus(spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished)

	cns := spos.NewConsensus(true, vld, th, rs, chr)

	sr := spos.NewSRBlock(true, int64(100*spos.RoundTimeDuration/100), cns, nil)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRBlock(t *testing.T) {
	sr := spos.NewSRBlock(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.NotNil(t, sr)
}

func TestSRBlock_DoWork1(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("1: Test case when send message is done but consensus is not done -> RNone\n")

	sr.OnSendBlock = SndWithSuccess

	r := sr.DoBlock(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.RNone, r)
}

func TestSRBlock_DoWork2(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("2: Test case when send message and consensus is done -> RTrue\n")

	sr.OnSendBlock = SndWithSuccess

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	r := sr.DoBlock(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.RTrue, r)
}

func TestSRBlock_DoWork3(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("3: Test case when time has expired -> RTrue\n")

	sr.OnSendBlock = SndWithoutSuccess

	chr.SetClockOffset(time.Duration(sr.EndTime() + 1))

	r := sr.DoBlock(chr)

	assert.Equal(t, spos.SsExtended, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.RTrue, r)
}

func TestSRBlock_DoWork4(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("4: Test case when receive message is done but consensus is not done -> RNone\n")

	sr.OnSendBlock = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(true)

	r := sr.DoBlock(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.RNone, r)
}

func TestSRBlock_DoWork5(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("5: Test case when receive message and consensus is done -> true\n")

	sr.OnSendBlock = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(true)

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, true, r)
}

func TestSRBlock_DoWork6(t *testing.T) {

	chr, sr := InitSRBlock()

	fmt.Printf("6: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnSendBlock = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(false)
	sr.Cns.Chr.SetSelfSubround(chronology.SrCanceled)

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSRBlock_Current(t *testing.T) {
	sr := spos.NewSRBlock(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrBlock), sr.Current())
}

func TestSRBlock_Next(t *testing.T) {
	sr := spos.NewSRBlock(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrComitmentHash), sr.Next())
}

func TestSRBlock_EndTime(t *testing.T) {
	sr := spos.NewSRBlock(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, int64(100*spos.RoundTimeDuration/100), sr.EndTime())
}

func TestSRBlock_Name(t *testing.T) {
	sr := spos.NewSRBlock(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, "<BLOCK>", sr.Name())
}

func TestSRBlock_Log(t *testing.T) {
	sr := spos.NewSRBlock(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	sr.Log("Test SRBlock")
}
