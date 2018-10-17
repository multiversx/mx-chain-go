package spos

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

func SndWithSuccess(chronology.Subround) bool {
	fmt.Printf("message was sent with success\n")
	return true
}

func SndWithoutSuccess(chronology.Subround) bool {
	fmt.Printf("message was NOT sent with success\n")
	return false
}

func RcvWithSuccess(rcvMsg *[]byte, chr *chronology.Chronology) bool {
	fmt.Printf("message was consumed with success\n")
	return true
}

func RcvWithoutSuccess(rcvMsg *[]byte, chr *chronology.Chronology) bool {
	fmt.Printf("message was NOT consumed with success\n")
	return false
}

func RcvWithoutSuccessAndCancel(rcvMsg *[]byte, chr *chronology.Chronology) bool {

	fmt.Printf("message was NOT consumed with success and this round should be canceled\n")
	chr.SetSelfSubround(chronology.SrCanceled)
	return false
}

func TestNewRoundStatus(t *testing.T) {
	rs := NewRoundStatus(SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished)

	assert.Equal(t, SsNotFinished, rs.ComitmentHash)

	rs.ComitmentHash = SsFinished

	assert.Equal(t, SsFinished, rs.ComitmentHash)

	rs.Bitmap = SsExtended

	assert.Equal(t, SsExtended, rs.Bitmap)

	rs.ResetRoundStatus()

	assert.Equal(t, SsNotFinished, rs.Bitmap)
}

func TestNewThreshold(t *testing.T) {
	th := NewThreshold(1, 2, 3, 4, 5)
	assert.Equal(t, 3, th.Bitmap)
}

func TestNewConsensus(t *testing.T) {

	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")
	th := NewThreshold(1, 2, 3, 4, 5)
	rs := NewRoundStatus(SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished)

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	cns1 := NewConsensus(true, nil, th, rs, chr)
	assert.Equal(t, 0, cap(cns1.ChRcvMsg))

	cns2 := NewConsensus(true, vld, th, rs, chr)
	assert.Equal(t, 3, cap(cns2.ChRcvMsg))
}

func TestConsensus_CheckConsensus(t *testing.T) {

}

func TestConsensus_IsNodeLeaderInCurrentRound(t *testing.T) {
	vld := NewValidators(nil, nil, []string{"1", "2", "3"}, "2")

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	cns := NewConsensus(true, vld, nil, nil, nil)

	assert.Equal(t, false, cns.IsNodeLeaderInCurrentRound("1"))

	cns.Chr = chr

	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
}

func TestConsensus_GetLeader(t *testing.T) {
	vld1 := NewValidators(nil, nil, nil, "")
	vld2 := NewValidators(nil, nil, []string{}, "")
	vld3 := NewValidators(nil, nil, []string{"1", "2", "3"}, "1")

	genesisTime := time.Now()
	currentTime := genesisTime

	rnd1 := chronology.NewRound(genesisTime, currentTime.Add(-roundTimeDuration), roundTimeDuration)
	rnd2 := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr1 := chronology.NewChronology(true, true, nil, genesisTime, &ntp.LocalTime{})
	chr2 := chronology.NewChronology(true, true, rnd1, genesisTime, &ntp.LocalTime{})
	chr3 := chronology.NewChronology(true, true, rnd2, genesisTime, &ntp.LocalTime{})

	cns1 := NewConsensus(true, nil, nil, nil, nil)
	cns2 := NewConsensus(true, vld1, nil, nil, nil)
	cns3 := NewConsensus(true, vld2, nil, nil, nil)
	cns4 := NewConsensus(true, vld3, nil, nil, nil)

	leader, err := cns1.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "Chronology is null", err.Error())
	assert.Equal(t, "", leader)

	cns1.Chr = chr1
	leader, err = cns1.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "Round is null", err.Error())
	assert.Equal(t, "", leader)

	cns1.Chr = chr2
	leader, err = cns1.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "Round index is negative", err.Error())
	assert.Equal(t, "", leader)

	cns2.Chr = chr3
	leader, err = cns2.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "ConsensusGroup is null", err.Error())
	assert.Equal(t, "", leader)

	cns3.Chr = chr3
	leader, err = cns3.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "ConsensusGroup is empty", err.Error())
	assert.Equal(t, "", leader)

	cns4.Chr = chr3
	leader, err = cns4.GetLeader()
	assert.Nil(t, err)
	assert.Equal(t, "1", leader)
}

func TestConsensus_Log(t *testing.T) {
	cns := NewConsensus(true, nil, nil, nil, nil)
	cns.Log("Test Consesnus")
}
