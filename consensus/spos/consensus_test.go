package spos_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func TestNewRoundStatus(t *testing.T) {

	rs := spos.NewRoundStatus()

	assert.NotNil(t, rs)

	rs.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	assert.Equal(t, spos.SsFinished, rs.Status(spos.SrCommitmentHash))

	rs.SetStatus(spos.SrBitmap, spos.SsExtended)
	assert.Equal(t, spos.SsExtended, rs.Status(spos.SrBitmap))

	rs.ResetRoundStatus()
	assert.Equal(t, spos.SsNotFinished, rs.Status(spos.SrBitmap))
}

func TestNewThreshold(t *testing.T) {
	rt := spos.NewRoundThreshold()

	rt.SetThreshold(spos.SrBlock, 1)
	rt.SetThreshold(spos.SrCommitmentHash, 2)
	rt.SetThreshold(spos.SrBitmap, 3)
	rt.SetThreshold(spos.SrCommitment, 4)
	rt.SetThreshold(spos.SrSignature, 5)

	assert.Equal(t, 3, rt.Threshold(spos.SrBitmap))
}

func TestNewConsensus(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	rt := spos.NewRoundThreshold()

	rt.SetThreshold(spos.SrBlock, 1)
	rt.SetThreshold(spos.SrCommitmentHash, 2)
	rt.SetThreshold(spos.SrBitmap, 3)
	rt.SetThreshold(spos.SrCommitment, 4)
	rt.SetThreshold(spos.SrSignature, 5)

	rs := spos.NewRoundStatus()

	rs.SetStatus(spos.SrBlock, spos.SsNotFinished)
	rs.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)
	rs.SetStatus(spos.SrBitmap, spos.SsNotFinished)
	rs.SetStatus(spos.SrCommitment, spos.SsNotFinished)
	rs.SetStatus(spos.SrSignature, spos.SsNotFinished)

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

	cns := spos.NewConsensus(true,
		nil,
		vld,
		rt,
		rs,
		chr)

	assert.NotNil(t, cns)
}

func InitConsensus() *spos.Consensus {
	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		"1")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

	rt := spos.NewRoundThreshold()

	rt.SetThreshold(spos.SrBlock, 1)
	rt.SetThreshold(spos.SrCommitmentHash, 7)
	rt.SetThreshold(spos.SrBitmap, 7)
	rt.SetThreshold(spos.SrCommitment, 7)
	rt.SetThreshold(spos.SrSignature, 7)

	rs := spos.NewRoundStatus()

	rs.SetStatus(spos.SrBlock, spos.SsFinished)
	rs.SetStatus(spos.SrCommitmentHash, spos.SsFinished)
	rs.SetStatus(spos.SrBitmap, spos.SsFinished)
	rs.SetStatus(spos.SrCommitment, spos.SsFinished)
	rs.SetStatus(spos.SrSignature, spos.SsFinished)

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

	cns := spos.NewConsensus(true,
		nil,
		vld,
		rt,
		rs,
		chr)

	return cns
}

func TestConsensus_IsNodeLeaderInCurrentRound(t *testing.T) {

	vld := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetJobDone(vld.ConsensusGroup()[i], spos.SrSignature, false)
	}

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

	cns := spos.NewConsensus(true,
		nil,
		vld,
		nil,
		nil,
		nil)

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))

	cns.Chr = chr
	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensus_GetLeader(t *testing.T) {

	vld1 := spos.NewRoundConsensus(
		nil,
		"")

	for i := 0; i < len(vld1.ConsensusGroup()); i++ {
		vld1.SetJobDone(vld1.ConsensusGroup()[i], spos.SrBlock, false)
		vld1.SetJobDone(vld1.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld1.SetJobDone(vld1.ConsensusGroup()[i], spos.SrBitmap, false)
		vld1.SetJobDone(vld1.ConsensusGroup()[i], spos.SrCommitment, false)
		vld1.SetJobDone(vld1.ConsensusGroup()[i], spos.SrSignature, false)
	}

	vld2 := spos.NewRoundConsensus(
		[]string{},
		"")

	for i := 0; i < len(vld2.ConsensusGroup()); i++ {
		vld2.ResetRoundState()
	}

	vld3 := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"1")

	for i := 0; i < len(vld3.ConsensusGroup()); i++ {
		vld3.SetJobDone(vld3.ConsensusGroup()[i], spos.SrBlock, false)
		vld3.SetJobDone(vld3.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld3.SetJobDone(vld3.ConsensusGroup()[i], spos.SrBitmap, false)
		vld3.SetJobDone(vld3.ConsensusGroup()[i], spos.SrCommitment, false)
		vld3.SetJobDone(vld3.ConsensusGroup()[i], spos.SrSignature, false)
	}

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd1 := chronology.NewRound(genesisTime,
		currentTime.Add(-RoundTimeDuration),
		RoundTimeDuration)

	rnd2 := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr1 := chronology.NewChronology(true,
		true,
		nil,
		genesisTime,
		&ntp.LocalTime{})

	chr2 := chronology.NewChronology(true,
		true,
		rnd1,
		genesisTime,
		&ntp.LocalTime{})

	chr3 := chronology.NewChronology(true,
		true,
		rnd2,
		genesisTime,
		&ntp.LocalTime{})

	cns1 := spos.NewConsensus(true,
		nil,
		nil,
		nil,
		nil,
		nil)

	cns2 := spos.NewConsensus(true,
		nil,
		vld1,
		nil,
		nil,
		nil)

	cns3 := spos.NewConsensus(true,
		nil,
		vld2,
		nil,
		nil,
		nil)

	cns4 := spos.NewConsensus(true,
		nil,
		vld3,
		nil,
		nil,
		nil)

	leader, err := cns1.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "chronology is null", err.Error())
	assert.Equal(t, "", leader)

	cns1.Chr = chr1
	leader, err = cns1.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "round is null", err.Error())
	assert.Equal(t, "", leader)

	cns1.Chr = chr2
	leader, err = cns1.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "round index is negative", err.Error())
	assert.Equal(t, "", leader)

	cns2.Chr = chr3
	leader, err = cns2.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "consensusGroup is null", err.Error())
	assert.Equal(t, "", leader)

	cns3.Chr = chr3
	leader, err = cns3.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "consensusGroup is empty", err.Error())
	assert.Equal(t, "", leader)

	cns4.Chr = chr3
	leader, err = cns4.GetLeader()
	assert.Nil(t, err)
	assert.Equal(t, "1", leader)
}

func TestConsensus_Log(t *testing.T) {

	cns := spos.NewConsensus(true,
		nil,
		nil,
		nil,
		nil,
		nil)

	assert.NotNil(t, cns)
}
