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

func SndWithSuccess() bool {
	fmt.Printf("message was sent with success\n")
	return true
}

func SndWithoutSuccess() bool {
	fmt.Printf("message was NOT sent with success\n")
	return false
}

func OnStartRound() {
	fmt.Printf("do init round\n")
}

func OnEndRound() {
	fmt.Printf("do finish round\n")
}

func TestNewRoundStatus(t *testing.T) {
	rs := spos.NewRoundStatus(spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished)

	assert.Equal(t, spos.SsNotFinished, rs.ComitmentHash)

	rs.ComitmentHash = spos.SsFinished

	assert.Equal(t, spos.SsFinished, rs.ComitmentHash)

	rs.Bitmap = spos.SsExtended

	assert.Equal(t, spos.SsExtended, rs.Bitmap)

	rs.ResetRoundStatus()

	assert.Equal(t, spos.SsNotFinished, rs.Bitmap)
}

func TestNewThreshold(t *testing.T) {
	th := spos.NewThreshold(1, 2, 3, 4, 5)
	assert.Equal(t, 3, th.Bitmap)
}

func TestNewConsensus(t *testing.T) {

	vld := spos.NewValidators(nil, nil, []string{"1", "2", "3"}, "2")
	th := spos.NewThreshold(1, 2, 3, 4, 5)
	rs := spos.NewRoundStatus(spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished)

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	cns := spos.NewConsensus(true, vld, th, rs, chr)
	assert.NotNil(t, cns)
}

func TestConsensus_CheckConsensus(t *testing.T) {
	rs := spos.NewRoundStatus(spos.SsFinished, spos.SsFinished, spos.SsFinished, spos.SsFinished, spos.SsFinished)
	cns := spos.NewConsensus(true, nil, nil, rs, nil)

	ok, n := cns.CheckConsensus(chronology.Subround(spos.SrStartRound), chronology.Subround(spos.SrEndRound))
	assert.Equal(t, false, ok)
	assert.Equal(t, -1, n)

	ok, n = cns.CheckConsensus(chronology.Subround(spos.SrBlock), chronology.Subround(spos.SrSignature))
	assert.Equal(t, true, ok)
	assert.Equal(t, 0, n)
}

func TestConsensus_IsNodeLeaderInCurrentRound(t *testing.T) {
	vld := spos.NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	cns := spos.NewConsensus(true, vld, nil, nil, nil)

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))

	cns.Chr = chr

	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensus_GetLeader(t *testing.T) {
	vld1 := spos.NewValidators(nil, nil, nil, "")
	vld2 := spos.NewValidators(nil, nil, []string{}, "")
	vld3 := spos.NewValidators(nil, nil, []string{"1", "2", "3"}, "1")

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd1 := chronology.NewRound(genesisTime, currentTime.Add(-spos.RoundTimeDuration), spos.RoundTimeDuration)
	rnd2 := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr1 := chronology.NewChronology(true, true, nil, genesisTime, &ntp.LocalTime{})
	chr2 := chronology.NewChronology(true, true, rnd1, genesisTime, &ntp.LocalTime{})
	chr3 := chronology.NewChronology(true, true, rnd2, genesisTime, &ntp.LocalTime{})

	cns1 := spos.NewConsensus(true, nil, nil, nil, nil)
	cns2 := spos.NewConsensus(true, vld1, nil, nil, nil)
	cns3 := spos.NewConsensus(true, vld2, nil, nil, nil)
	cns4 := spos.NewConsensus(true, vld3, nil, nil, nil)

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
	cns := spos.NewConsensus(true, nil, nil, nil, nil)
	cns.Log("Test Consesnus")
}

func TestGettersAndSetters(t *testing.T) {
	cns := spos.NewConsensus(true, nil, nil, nil, nil)

	cns.SetSentMessage(true)
	assert.Equal(t, true, cns.SentMessage())

	cns.SetReceivedMessage(true)
	assert.Equal(t, true, cns.ReceivedMessage())
}
