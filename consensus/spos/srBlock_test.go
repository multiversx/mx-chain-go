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

func RcvWithoutSuccessAndAbord(rcvMsg *[]byte, chr *chronology.Chronology) bool {

	fmt.Printf("message was consumed with success, but this round should be aborded\n")
	chr.SetSelfSubround(chronology.SR_ABORDED)
	return false
}

func RcvWithoutSuccess(rcvMsg *[]byte, chr *chronology.Chronology) bool {
	fmt.Printf("message was NOT consumed with success\n")
	return false
}

func TestNewSRBlock(t *testing.T) {
	sr := NewSRBlock(true, int64(100*ROUND_TIME_DURATION/100), nil, nil, nil)
	assert.NotNil(t, sr)
}

func TestSRBlock_DoWork(t *testing.T) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, ROUND_TIME_DURATION)

	chr := chronology.NewChronology(true, true, &rnd, genesisTime, &ntp.LocalTime{ClockOffset: 0})

	vld := NewValidators([]string{"1", "2", "3"}, "2")
	th := NewThreshold(1, 2*len(vld.ConsensusGroup)/3, 2*len(vld.ConsensusGroup)/3, 2*len(vld.ConsensusGroup)/3, 2*len(vld.ConsensusGroup)/3)
	rs := NewRoundStatus(SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED)

	cns := NewConsensus(true, vld, th, rs, &chr)

	sr := NewSRBlock(true, int64(100*ROUND_TIME_DURATION/100), &cns, nil, nil)
	chr.AddSubrounder(&sr)

	fmt.Printf("1: Test case when send message is done but consensus is not done -> R_None\n")
	sr.OnSendMessage = SndWithSuccess
	r := sr.doBlock(&chr)
	assert.Equal(t, r, R_None)

	fmt.Printf("2: Test case when send message and consensus is done -> R_True\n")
	rv := sr.cns.ValidationMap[sr.cns.Self]
	rv.Block = true
	sr.cns.ValidationMap[sr.cns.Self] = rv
	r = sr.doBlock(&chr)
	assert.Equal(t, r, R_True)

	fmt.Printf("3: Test case when time has expired -> R_True\n")
	sr.OnSendMessage = SndWithoutSuccess
	chr.SetClockOffset(time.Duration(sr.endTime + 1))
	r = sr.doBlock(&chr)
	assert.Equal(t, r, R_True)

	fmt.Printf("4: Test case when receive message is done but consensus is not done -> R_None\n")
	rv = sr.cns.ValidationMap[sr.cns.Self]
	rv.Block = false
	sr.cns.ValidationMap[sr.cns.Self] = rv
	chr.SetClockOffset(0)
	sr.cns.ChRcvMsg <- []byte("Message has come")
	sr.OnReceivedMessage = RcvWithSuccess
	r = sr.doBlock(&chr)
	assert.Equal(t, r, R_None)

	fmt.Printf("5: Test case when receive message and consensus is done -> R_True\n")
	rv = sr.cns.ValidationMap[sr.cns.Self]
	rv.Block = true
	sr.cns.ValidationMap[sr.cns.Self] = rv
	sr.cns.ChRcvMsg <- []byte("Message has come")
	r2 := sr.DoWork(&chr)
	assert.Equal(t, r2, true)

	fmt.Printf("6: Test case when receive message is done but round should be aborded -> R_None\n")
	sr.cns.ChRcvMsg <- []byte("Message has come")
	sr.OnReceivedMessage = RcvWithoutSuccessAndAbord
	r2 = sr.DoWork(&chr)
	assert.Equal(t, r2, false)
}

func TestSRBlock_Current(t *testing.T) {
	sr := NewSRBlock(true, int64(100*ROUND_TIME_DURATION/100), nil, nil, nil)
	assert.Equal(t, sr.Current(), chronology.Subround(SR_BLOCK))
}

func TestSRBlock_Next(t *testing.T) {
	sr := NewSRBlock(true, int64(100*ROUND_TIME_DURATION/100), nil, nil, nil)
	assert.Equal(t, sr.Next(), chronology.Subround(SR_COMITMENT_HASH))
}

func TestSRBlock_EndTime(t *testing.T) {
	sr := NewSRBlock(true, int64(100*ROUND_TIME_DURATION/100), nil, nil, nil)
	assert.Equal(t, sr.EndTime(), int64(100*ROUND_TIME_DURATION/100))
}

func TestSRBlock_Name(t *testing.T) {
	sr := NewSRBlock(true, int64(100*ROUND_TIME_DURATION/100), nil, nil, nil)
	assert.Equal(t, sr.Name(), "<BLOCK>")
}

func TestSRBlock_Log(t *testing.T) {
	sr := NewSRBlock(true, int64(100*ROUND_TIME_DURATION/100), nil, nil, nil)
	sr.Log("Test")
}
