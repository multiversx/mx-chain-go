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

func InitSRImpl() (*chronology.Chronology, *spos.Consensus) {

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr := chronology.NewChronology(true,
		true,
		rnd,
		genesisTime,
		&ntp.LocalTime{})

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		"1")

	pbft := 2*len(vld.ConsensusGroup())/3 + 1

	th := spos.NewThreshold(1,
		pbft,
		pbft,
		pbft,
		pbft)

	rs := spos.NewRoundStatus(spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished)

	cns := spos.NewConsensus(true,
		vld,
		th,
		rs,
		chr)

	return chr, cns
}

func TestSRImpl_DoWork1(t *testing.T) {

	fmt.Printf("1: Test the case when consensus is done -> true\n")

	chr, _ := InitSRImpl()

	sr := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithSuccess)

	chr.AddSubround(sr)
	r := sr.DoWork(chr)
	assert.Equal(t, true, r)
}

func TestSRImpl_DoWork2(t *testing.T) {

	fmt.Printf("2: Test the case when consensus is not done -> false\n")

	chr, _ := InitSRImpl()

	sr := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr)
	chr.SetSelfSubround(chronology.SrCanceled)
	r := sr.DoWork(chr)
	assert.Equal(t, false, r)
}

func TestSRImpl_DoWork3(t *testing.T) {

	fmt.Printf("3: Test the case when time has expired -> true\n")

	chr, _ := InitSRImpl()

	sr1 := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr1)

	sr2 := spos.NewSRImpl(chronology.Subround(spos.SrCommitmentHash),
		chronology.Subround(spos.SrBitmap),
		int64(40*RoundTimeDuration/100),
		"<COMMITMENT_HASH>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr2)

	chr.SetClockOffset(time.Duration(sr1.EndTime() + 1))
	r := sr1.DoWork(chr)
	assert.Equal(t, true, r)
}

func TestSRImpl_DoWork4(t *testing.T) {

	fmt.Printf("4: Test the case when job is done but consensus is not done and than the time will be " +
		"expired -> true\n")

	chr, _ := InitSRImpl()

	sr1 := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr1)

	sr2 := spos.NewSRImpl(chronology.Subround(spos.SrCommitmentHash),
		chronology.Subround(spos.SrBitmap),
		int64(40*RoundTimeDuration/100),
		"<COMMITMENT_HASH>",
		DoSubroundJob,
		DoExtendSubround,
		DoCheckConsensusWithoutSuccess)

	chr.AddSubround(sr2)
	chr.SetClockOffset(time.Duration(sr1.EndTime() - int64(5*time.Millisecond)))
	r := sr1.DoWork(chr)
	assert.Equal(t, true, r)
}

func TestNewSRImpl(t *testing.T) {

	sr := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.NotNil(t, sr)
}

func TestSRImpl_Current(t *testing.T) {

	sr := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, chronology.Subround(spos.SrBlock), sr.Current())
}

func TestSRImpl_Next(t *testing.T) {

	sr := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, chronology.Subround(spos.SrCommitmentHash), sr.Next())
}

func TestSRImpl_EndTime(t *testing.T) {

	sr := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, int64(25*RoundTimeDuration/100), sr.EndTime())
}

func TestSRImpl_Name(t *testing.T) {

	sr := spos.NewSRImpl(chronology.Subround(spos.SrBlock),
		chronology.Subround(spos.SrCommitmentHash),
		int64(25*RoundTimeDuration/100),
		"<BLOCK>",
		nil,
		nil,
		nil)

	assert.Equal(t, "<BLOCK>", sr.Name())
}
