package spos_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func DoSubroundJob() bool {
	fmt.Printf("do job\n")
	time.Sleep(5 * time.Millisecond)
	return true
}

func DoExtendSubround() {
	fmt.Printf("do extend subround\n")
}

func DoCheckConsensusWithSuccess() bool {
	fmt.Printf("do check consensus with success in subround \n")
	return true
}

func DoCheckConsensusWithoutSuccess() bool {
	fmt.Printf("do check consensus without success in subround \n")
	return false
}

func InitSubround() (*chronology.Chronology, *spos.Spos) {

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil),
	)

	rcns := spos.NewRoundConsensus(
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		9,
		"1")

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		rcns.ResetRoundState()
	}

	pbft := 2*len(rcns.ConsensusGroup())/3 + 1

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, pbft)
	rthr.SetThreshold(bn.SrBitmap, pbft)
	rthr.SetThreshold(bn.SrCommitment, pbft)
	rthr.SetThreshold(bn.SrSignature, pbft)

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrBlock, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitment, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrSignature, spos.SsNotFinished)

	sps := spos.NewSpos(
		nil,
		rcns,
		rthr,
		rstatus,
		chr,
	)

	return chr, sps
}

func TestSubround_DoWork1(t *testing.T) {

	fmt.Printf("1: Test the case when consensus is done -> true\n")

	chr, _ := InitSubround()

	sr := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithSuccess)

	chr.AddSubround(sr)

	computeSubRoundId := func() chronology.SubroundId {
		return chr.GetSubroundFromDateTime(chr.SyncTimer().CurrentTime(chr.ClockOffset()))
	}

	isCancelled := func() bool {
		return chr.SelfSubround() == chronology.SubroundId(-1)
	}

	r := sr.DoWork(computeSubRoundId, isCancelled)
	assert.Equal(t, true, r)
}

func TestSubround_DoWork2(t *testing.T) {

	fmt.Printf("2: Test the case when consensus is not done -> false\n")

	chr, _ := InitSubround()

	sr := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr)
	chr.SetSelfSubround(-1)
	computeSubRoundId := func() chronology.SubroundId {
		return chr.GetSubroundFromDateTime(chr.SyncTimer().CurrentTime(chr.ClockOffset()))
	}

	isCancelled := func() bool {
		return chr.SelfSubround() == chronology.SubroundId(-1)
	}

	r := sr.DoWork(computeSubRoundId, isCancelled)
	assert.Equal(t, false, r)
}

func TestSubround_DoWork3(t *testing.T) {

	fmt.Printf("3: Test the case when time has expired -> true\n")

	chr, _ := InitSubround()

	sr1 := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr1)

	sr2 := spos.NewSubround(chronology.SubroundId(bn.SrCommitmentHash),
		chronology.SubroundId(bn.SrBitmap),
		int64(40*RoundTimeDuration/100),
		"<COMMITMENT_HASH>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr2)

	chr.SetClockOffset(time.Duration(sr1.EndTime() + 1))
	computeSubRoundId := func() chronology.SubroundId {
		return chr.GetSubroundFromDateTime(chr.SyncTimer().CurrentTime(chr.ClockOffset()))
	}

	isCancelled := func() bool {
		return chr.SelfSubround() == chronology.SubroundId(-1)
	}

	r := sr1.DoWork(computeSubRoundId, isCancelled)
	assert.Equal(t, true, r)
}

func TestSubround_DoWork4(t *testing.T) {

	fmt.Printf("4: Test the case when job is done but consensus is not done and than the time will be " +
		"expired -> true\n")

	chr, _ := InitSubround()

	sr1 := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr1)

	sr2 := spos.NewSubround(chronology.SubroundId(bn.SrCommitmentHash),
		chronology.SubroundId(bn.SrBitmap),
		int64(40*RoundTimeDuration/100),
		"<COMMITMENT_HASH>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr2)
	chr.SetClockOffset(time.Duration(sr1.EndTime() - int64(5*time.Millisecond)))
	computeSubRoundId := func() chronology.SubroundId {
		return chr.GetSubroundFromDateTime(chr.SyncTimer().CurrentTime(chr.ClockOffset()))
	}

	isCancelled := func() bool {
		return chr.SelfSubround() == chronology.SubroundId(-1)
	}

	r := sr1.DoWork(computeSubRoundId, isCancelled)
	assert.Equal(t, true, r)
}

func TestNewSubround(t *testing.T) {

	sr := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.NotNil(t, sr)
}

func TestSubround_Current(t *testing.T) {

	sr := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, chronology.SubroundId(bn.SrBlock), sr.Current())
}

func TestSubround_Next(t *testing.T) {

	sr := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, chronology.SubroundId(bn.SrCommitmentHash), sr.Next())
}

func TestSubround_EndTime(t *testing.T) {

	sr := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, int64(25*RoundTimeDuration/100), sr.EndTime())
}

func TestSubround_Name(t *testing.T) {

	sr := spos.NewSubround(chronology.SubroundId(bn.SrBlock),
		chronology.SubroundId(bn.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, "<BLOCK>", sr.Name())
}
