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

// RoundTimeDuration defines the time duration in milliseconds of each round
const RoundTimeDuration = time.Duration(4000 * time.Millisecond)

func DoSubroundJob() bool {
	fmt.Printf("do job\n")
	time.Sleep(5 * time.Millisecond)
	return true
}

func DoExtendSubround() {
	fmt.Printf("do extend subround\n")
}

func DoCheckConsensusWithSuccess(subround spos.Subround) bool {
	fmt.Printf("do check consensus with success in subround %d\n", subround)
	return true
}

func DoCheckConsensusWithoutSuccess(subround spos.Subround) bool {
	fmt.Printf("do check consensus without success in subround %d\n", subround)
	return false
}

func TestNewRoundStatus(t *testing.T) {

	rs := spos.NewRoundStatus(spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished)

	assert.Equal(t, spos.SsNotFinished, rs.CommitmentHash())

	rs.SetCommitmentHash(spos.SsFinished)
	assert.Equal(t, spos.SsFinished, rs.CommitmentHash())

	rs.SetBitmap(spos.SsExtended)
	assert.Equal(t, spos.SsExtended, rs.Bitmap())

	rs.ResetRoundStatus()
	assert.Equal(t, spos.SsNotFinished, rs.Bitmap())
}

func TestNewThreshold(t *testing.T) {
	th := spos.NewThreshold(1, 2, 3, 4, 5)
	assert.Equal(t, 3, th.Bitmap())
}

func TestNewConsensus(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	th := spos.NewThreshold(1,
		2,
		3,
		4,
		5)

	rs := spos.NewRoundStatus(spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished,
		spos.SsNotFinished)

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
		vld,
		th,
		rs,
		chr)

	assert.NotNil(t, cns)
}

func InitConsensus() *spos.Consensus {
	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		"1")

	th := spos.NewThreshold(1,
		7,
		7,
		7,
		7)

	rs := spos.NewRoundStatus(spos.SsFinished,
		spos.SsFinished,
		spos.SsFinished,
		spos.SsFinished,
		spos.SsFinished)

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
		vld,
		th,
		rs,
		chr)

	return cns
}

func TestConsensus_CheckConsensus(t *testing.T) {
	cns := InitConsensus()

	ok := cns.CheckConsensus(spos.SrStartRound)
	assert.Equal(t, true, ok)

	ok = cns.CheckConsensus(spos.SrEndRound)
	assert.Equal(t, true, ok)

	ok = cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, false, ok)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, true, ok)
}

func TestConsensus_CheckBlockConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.RoundStatus.SetBlock(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrBlock)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.Block())

	cns.SetValidationMap("2", true, spos.SrBlock)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBlock)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.Block())
}

func TestConsensus_CheckCommitmentHashConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.RoundStatus.SetCommitmentHash(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.CommitmentHash())

	for i := 0; i < cns.Threshold.CommitmentHash(); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrCommitmentHash)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.CommitmentHash())

	cns.Validators.SetSelf("2")

	cns.RoundStatus.SetCommitmentHash(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.CommitmentHash())

	for i := 0; i < cns.Threshold.Bitmap(); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrBitmap)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.CommitmentHash())

	for i := 0; i < cns.Threshold.Bitmap(); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], false, spos.SrBitmap)
	}

	for i := 0; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrCommitmentHash)
	}

	cns.RoundStatus.SetCommitmentHash(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.CommitmentHash())
}

func TestConsensus_CheckBitmapConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.RoundStatus.SetBitmap(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.Bitmap())

	for i := 1; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrCommitmentHash)
	}

	for i := 0; i < cns.Threshold.Bitmap(); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrBitmap)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.Bitmap())

	cns.SetValidationMap(cns.Validators.ConsensusGroup()[0], true, spos.SrCommitmentHash)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.Bitmap())

	for i := 1; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrBitmap)
	}

	cns.SetValidationMap(cns.Validators.Self(), false, spos.SrBitmap)

	cns.RoundStatus.SetBitmap(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.Bitmap())
}

func TestConsensus_CheckCommitmentConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.RoundStatus.SetCommitment(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrCommitment)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.Commitment())

	for i := 0; i < cns.Threshold.Bitmap(); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrBitmap)
	}

	for i := 1; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrCommitment)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitment)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.Commitment())

	cns.SetValidationMap(cns.Validators.ConsensusGroup()[0], true, spos.SrCommitment)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitment)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.Commitment())
}

func TestConsensus_CheckSignatureConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.RoundStatus.SetSignature(spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.Signature())

	for i := 0; i < cns.Threshold.Bitmap(); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrBitmap)
	}

	for i := 1; i < cns.Threshold.Signature(); i++ {
		cns.SetValidationMap(cns.Validators.ConsensusGroup()[i], true, spos.SrSignature)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.RoundStatus.Signature())

	cns.SetValidationMap(cns.Validators.ConsensusGroup()[0], true, spos.SrSignature)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.RoundStatus.Signature())
}

func TestConsensus_IsNodeLeaderInCurrentRound(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

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
		vld,
		nil,
		nil,
		nil)

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))

	cns.Chr = chr
	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensus_GetLeader(t *testing.T) {

	vld1 := spos.NewValidators(nil,
		nil,
		nil,
		"")

	vld2 := spos.NewValidators(nil,
		nil,
		[]string{},
		"")

	vld3 := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"1")

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
		nil)

	cns2 := spos.NewConsensus(true,
		vld1,
		nil,
		nil,
		nil)

	cns3 := spos.NewConsensus(true,
		vld2,
		nil,
		nil,
		nil)

	cns4 := spos.NewConsensus(true,
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
		nil)

	cns.Log("Test Consesnus")
}

func TestGettersAndSetters(t *testing.T) {

	cns := spos.NewConsensus(true,
		nil,
		nil,
		nil,
		nil)

	cns.SetShouldCheckConsensus(true)
	assert.Equal(t, true, cns.ShouldCheckConsensus())
}
