package spos_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/stretchr/testify/assert"
)

// RoundTimeDuration defines the time duration in milliseconds of each round
const RoundTimeDuration = time.Duration(4000 * time.Millisecond)

func TestNewSpos(t *testing.T) {

	rcns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		3,
		"2")

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrSignature, false)
	}

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, 2)
	rthr.SetThreshold(bn.SrBitmap, 3)
	rthr.SetThreshold(bn.SrCommitment, 4)
	rthr.SetThreshold(bn.SrSignature, 5)

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrBlock, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitment, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrSignature, spos.SsNotFinished)

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

	sps := spos.NewSpos(
		nil,
		rcns,
		rthr,
		rstatus,
		chr,
	)

	assert.NotNil(t, sps)
}

func TestSpos_IsNodeLeaderInCurrentRound(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}
	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	for i := 0; i < len(rcns.ConsensusGroup()); i++ {
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBlock, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns.SetJobDone(rcns.ConsensusGroup()[i], bn.SrSignature, false)
	}

	sps := spos.NewSpos(
		nil,
		rcns,
		nil,
		nil,
		nil,
	)

	assert.Equal(t, true, sps.IsNodeLeaderInCurrentRound("1"))
}

func TestSpos_GetLeader(t *testing.T) {

	rcns1 := spos.NewRoundConsensus(
		nil,
		0,
		"")

	rcns1.SetConsensusGroup(nil)

	for i := 0; i < len(rcns1.ConsensusGroup()); i++ {
		rcns1.SetJobDone(rcns1.ConsensusGroup()[i], bn.SrBlock, false)
		rcns1.SetJobDone(rcns1.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns1.SetJobDone(rcns1.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns1.SetJobDone(rcns1.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns1.SetJobDone(rcns1.ConsensusGroup()[i], bn.SrSignature, false)
	}

	rcns2 := spos.NewRoundConsensus(
		[]string{},
		0,
		"")

	rcns2.SetConsensusGroup([]string{})

	for i := 0; i < len(rcns2.ConsensusGroup()); i++ {
		rcns2.ResetRoundState()
	}

	rcns3 := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		3,
		"1")

	rcns3.SetConsensusGroup([]string{"1", "2", "3"})

	for i := 0; i < len(rcns3.ConsensusGroup()); i++ {
		rcns3.SetJobDone(rcns3.ConsensusGroup()[i], bn.SrBlock, false)
		rcns3.SetJobDone(rcns3.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		rcns3.SetJobDone(rcns3.ConsensusGroup()[i], bn.SrBitmap, false)
		rcns3.SetJobDone(rcns3.ConsensusGroup()[i], bn.SrCommitment, false)
		rcns3.SetJobDone(rcns3.ConsensusGroup()[i], bn.SrSignature, false)
	}

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd2 := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr3 := chronology.NewChronology(
		true,
		rnd2,
		genesisTime,
		ntp.NewSyncTime(RoundTimeDuration, nil))

	sps2 := spos.NewSpos(
		nil,
		rcns1,
		nil,
		nil,
		nil,
	)

	sps3 := spos.NewSpos(
		nil,
		rcns2,
		nil,
		nil,
		nil,
	)

	sps4 := spos.NewSpos(
		nil,
		rcns3,
		nil,
		nil,
		nil,
	)

	sps2.Chr = chr3
	leader, err := sps2.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "consensusGroup is null", err.Error())
	assert.Equal(t, "", leader)

	sps3.Chr = chr3
	leader, err = sps3.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "consensusGroup is empty", err.Error())
	assert.Equal(t, "", leader)

	sps4.Chr = chr3
	leader, err = sps4.GetLeader()
	assert.Nil(t, err)
	assert.Equal(t, "1", leader)
}

func TestSpos_IsSelfLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	rcns := spos.NewRoundConsensus(
		[]string{"1", "2", "3"},
		3,
		"2")

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		&mock.SyncTimerMock{})

	sps := spos.NewSpos(
		nil,
		rcns,
		nil,
		nil,
		chr,
	)

	assert.False(t, sps.IsSelfLeaderInCurrentRound())
}

func TestSpos_IsSelfLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime,
		currentTime,
		RoundTimeDuration)

	chr := chronology.NewChronology(
		true,
		rnd,
		genesisTime,
		&mock.SyncTimerMock{})

	sps := spos.NewSpos(
		nil,
		rcns,
		nil,
		nil,
		chr,
	)

	rnd.UpdateRound(genesisTime, genesisTime.Add(RoundTimeDuration))
	assert.False(t, sps.IsSelfLeaderInCurrentRound())
}
