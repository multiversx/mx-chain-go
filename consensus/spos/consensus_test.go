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

func DoCheckConsensusWithSuccess(subroundId chronology.SubroundId) bool {
	fmt.Printf("do check consensus with success in subround %s\n", GetSubroundName(subroundId))
	return true
}

func DoCheckConsensusWithoutSuccess(subroundId chronology.SubroundId) bool {
	fmt.Printf("do check consensus without success in subround %s\n", GetSubroundName(subroundId))
	return false
}

func GetSubroundName(subroundId chronology.SubroundId) string {
	switch subroundId {
	case spos.SrStartRound:
		return "<START_ROUND>"
	case spos.SrBlock:
		return "<BLOCK>"
	case spos.SrCommitmentHash:
		return "<COMMITMENT_HASH>"
	case spos.SrBitmap:
		return "<BITMAP>"
	case spos.SrCommitment:
		return "<COMMITMENT>"
	case spos.SrSignature:
		return "<SIGNATURE>"
	case spos.SrEndRound:
		return "<END_ROUND>"
	default:
		return "Undifined subround"
	}
}

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

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
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
	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
		"1")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
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

	cns.SetStatus(spos.SrBlock, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrBlock)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrBlock))

	cns.SetValidation("2", spos.SrBlock, true)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBlock)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrBlock))
}

func TestConsensus_CheckCommitmentHashConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitmentHash))

	for i := 0; i < cns.Threshold(spos.SrCommitmentHash); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitmentHash))

	cns.Validators.SetSelfId("2")

	cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitmentHash))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitmentHash))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrBitmap, false)
	}

	for i := 0; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	cns.SetStatus(spos.SrCommitmentHash, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitmentHash)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitmentHash))
}

func TestConsensus_CheckBitmapConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrBitmap))

	for i := 1; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrCommitmentHash, true)
	}

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrBitmap))

	cns.SetValidation(cns.Validators.ConsensusGroup()[0], spos.SrCommitmentHash, true)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrBitmap))

	for i := 1; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	cns.SetValidation(cns.Validators.SelfId(), spos.SrBitmap, false)

	cns.SetStatus(spos.SrBitmap, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrBitmap)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrBitmap))
}

func TestConsensus_CheckCommitmentConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrCommitment, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrCommitment)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitment))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	for i := 1; i < len(cns.Validators.ConsensusGroup()); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrCommitment, true)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitment)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrCommitment))

	cns.SetValidation(cns.Validators.ConsensusGroup()[0], spos.SrCommitment, true)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrCommitment)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrCommitment))
}

func TestConsensus_CheckSignatureConsensus(t *testing.T) {
	cns := InitConsensus()

	cns.SetStatus(spos.SrSignature, spos.SsNotFinished)

	cns.SetShouldCheckConsensus(true)
	ok := cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrSignature))

	for i := 0; i < cns.Threshold(spos.SrBitmap); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrBitmap, true)
	}

	for i := 1; i < cns.Threshold(spos.SrSignature); i++ {
		cns.SetValidation(cns.Validators.ConsensusGroup()[i], spos.SrSignature, true)
	}

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, false, ok)
	assert.Equal(t, spos.SsNotFinished, cns.Status(spos.SrSignature))

	cns.SetValidation(cns.Validators.ConsensusGroup()[0], spos.SrSignature, true)

	cns.SetShouldCheckConsensus(true)
	ok = cns.CheckConsensus(spos.SrSignature)
	assert.Equal(t, true, ok)
	assert.Equal(t, spos.SsFinished, cns.Status(spos.SrSignature))
}

func TestConsensus_IsNodeLeaderInCurrentRound(t *testing.T) {

	vld := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"2")

	for i := 0; i < len(vld.ConsensusGroup()); i++ {
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBlock, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrBitmap, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrCommitment, false)
		vld.SetValidation(vld.ConsensusGroup()[i], spos.SrSignature, false)
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

	vld1 := spos.NewValidators(nil,
		nil,
		nil,
		"")

	for i := 0; i < len(vld1.ConsensusGroup()); i++ {
		vld1.SetValidation(vld1.ConsensusGroup()[i], spos.SrBlock, false)
		vld1.SetValidation(vld1.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld1.SetValidation(vld1.ConsensusGroup()[i], spos.SrBitmap, false)
		vld1.SetValidation(vld1.ConsensusGroup()[i], spos.SrCommitment, false)
		vld1.SetValidation(vld1.ConsensusGroup()[i], spos.SrSignature, false)
	}

	vld2 := spos.NewValidators(nil,
		nil,
		[]string{},
		"")

	for i := 0; i < len(vld2.ConsensusGroup()); i++ {
		vld2.SetValidation(vld2.ConsensusGroup()[i], spos.SrBlock, false)
		vld2.SetValidation(vld2.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld2.SetValidation(vld2.ConsensusGroup()[i], spos.SrBitmap, false)
		vld2.SetValidation(vld2.ConsensusGroup()[i], spos.SrCommitment, false)
		vld2.SetValidation(vld2.ConsensusGroup()[i], spos.SrSignature, false)
	}

	vld3 := spos.NewValidators(nil,
		nil,
		[]string{"1", "2", "3"},
		"1")

	for i := 0; i < len(vld3.ConsensusGroup()); i++ {
		vld3.SetValidation(vld3.ConsensusGroup()[i], spos.SrBlock, false)
		vld3.SetValidation(vld3.ConsensusGroup()[i], spos.SrCommitmentHash, false)
		vld3.SetValidation(vld3.ConsensusGroup()[i], spos.SrBitmap, false)
		vld3.SetValidation(vld3.ConsensusGroup()[i], spos.SrCommitment, false)
		vld3.SetValidation(vld3.ConsensusGroup()[i], spos.SrSignature, false)
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

	cns.Log("Test Consesnus")
}

func TestGettersAndSetters(t *testing.T) {

	cns := spos.NewConsensus(true,
		nil,
		nil,
		nil,
		nil,
		nil)

	cns.SetShouldCheckConsensus(true)
	assert.Equal(t, true, cns.ShouldCheckConsensus())
}
