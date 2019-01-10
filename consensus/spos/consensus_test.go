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

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
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

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	cns := spos.NewConsensus(
		nil,
		rCns,
		rt,
		rs,
		chr,
	)

	assert.NotNil(t, cns)
}

func InitConsensus() *spos.Consensus {
	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		"1")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
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

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	cns := spos.NewConsensus(
		nil,
		rCns,
		rt,
		rs,
		chr,
	)

	return cns
}

func TestConsensus_IsNodeLeaderInCurrentRound(t *testing.T) {

	rCns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(rCns.ConsensusGroup()); i++ {
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBlock, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns.SetJobDone(rCns.ConsensusGroup()[i], spos.SrSignature, false)
	}

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	cns := spos.NewConsensus(
		nil,
		rCns,
		nil,
		nil,
		nil,
	)

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))

	cns.Chr = chr
	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensus_GetLeader(t *testing.T) {

	rCns1 := spos.NewRoundConsensus(
		nil,
		"")

	for i := 0; i < len(rCns1.ConsensusGroup()); i++ {
		rCns1.SetJobDone(rCns1.ConsensusGroup()[i], spos.SrBlock, false)
		rCns1.SetJobDone(rCns1.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns1.SetJobDone(rCns1.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns1.SetJobDone(rCns1.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns1.SetJobDone(rCns1.ConsensusGroup()[i], spos.SrSignature, false)
	}

	rCns2 := spos.NewRoundConsensus(
		[]string{},
		"")

	for i := 0; i < len(rCns2.ConsensusGroup()); i++ {
		rCns2.ResetRoundState()
	}

	rCns3 := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		"1")

	for i := 0; i < len(rCns3.ConsensusGroup()); i++ {
		rCns3.SetJobDone(rCns3.ConsensusGroup()[i], spos.SrBlock, false)
		rCns3.SetJobDone(rCns3.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		rCns3.SetJobDone(rCns3.ConsensusGroup()[i], spos.SrBitmap, false)
		rCns3.SetJobDone(rCns3.ConsensusGroup()[i], spos.SrCommitment, false)
		rCns3.SetJobDone(rCns3.ConsensusGroup()[i], spos.SrSignature, false)
	}

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd1 := chronology.NewRound(genesisTime,
		currentTime.Add(-RoundTimeDuration),
		RoundTimeDuration)

	rnd2 := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr1 := chronology.NewChronology(
		true,
		nil,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	chr2 := chronology.NewChronology(
		true,
		rnd1,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	chr3 := chronology.NewChronology(
		true,
		rnd2,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	cns1 := spos.NewConsensus(
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	cns2 := spos.NewConsensus(
		nil,
		rCns1,
		nil,
		nil,
		nil,
	)

	cns3 := spos.NewConsensus(
		nil,
		rCns2,
		nil,
		nil,
		nil,
	)

	cns4 := spos.NewConsensus(
		nil,
		rCns3,
		nil,
		nil,
		nil,
	)

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

	cns := spos.NewConsensus(
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	assert.NotNil(t, cns)
}
