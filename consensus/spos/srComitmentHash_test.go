package spos

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/stretchr/testify/assert"
)

func InitSRComitmentHash() (*chronology.Chronology, *SRComitmentHash) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := NewRoundStatus(SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished, SsNotFinished)

	cns := NewConsensus(true, vld, th, rs, chr)

	sr := NewSRComitmentHash(true, int64(100*roundTimeDuration/100), cns, nil, nil)
	chr.AddSubrounder(sr)

	return chr, sr
}

func TestNewSRComitmentHash(t *testing.T) {
	sr := NewSRComitmentHash(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.NotNil(t, sr)
}

func TestSRComitmentHash_DoWork1(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("1: Test case when send message is done but consensus is not done -> rNone\n")

	sr.OnSendMessage = SndWithSuccess

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, rNone, r)
}

func TestSRComitmentHash_DoWork2(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("2: Test case when send message is done but consensus is semi-done (only subround BLOCK is done) -> rNone\n")

	sr.OnSendMessage = SndWithSuccess

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, rNone, r)
}

func TestSRComitmentHash_DoWork3(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("3: Test case when send message is done, but consensus (full) is not done (I AM NOT LEADER) -> rNone\n")

	sr.OnSendMessage = SndWithSuccess

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	for i := 0; i < sr.cns.Threshold.ComitmentHash; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, rNone, r)
}

func TestSRComitmentHash_DoWork4(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("4: Test case when send message is done and consensus (pbft) is enough (I AM LEADER) -> rTrue\n")

	sr.OnSendMessage = SndWithSuccess

	sr.cns.Self = "1"

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	for i := 0; i < sr.cns.Threshold.ComitmentHash; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, rTrue, r)
}

func TestSRComitmentHash_DoWork5(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("5: Test case when send message is done and consensus (pbft) is enough because of the bitmap (I AM NOT LEADER) -> rTrue\n")

	sr.OnSendMessage = SndWithSuccess

	sr.cns.RoundStatus.Block = SsFinished

	for i := 0; i < sr.cns.Threshold.Bitmap; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, rTrue, r)
}

func TestSRComitmentHash_DoWork6(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("6: Test case when send message and consensus (full) is done (I AM NOT LEADER) -> rTrue\n")

	sr.OnSendMessage = SndWithSuccess

	sr.cns.RoundStatus.Block = SsFinished

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, rTrue, r)
}

func TestSRComitmentHash_DoWork7(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("7: Test case when time has expired with consensus (pbft) not done -> rTrue\n")

	sr.OnSendMessage = SndWithSuccess

	chr.SetClockOffset(time.Duration(sr.endTime + 1))

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsExtended, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, r, rTrue)
}

func TestSRComitmentHash_DoWork8(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("8: Test case when time has expired with consensus (pbft) done -> rTrue\n")

	sr.OnSendMessage = SndWithoutSuccess

	chr.SetClockOffset(time.Duration(sr.endTime + 1))

	for i := 0; i < sr.cns.Threshold.ComitmentHash; i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsExtended, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, r, rTrue)
}

func TestSRComitmentHash_DoWork9(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("9: Test case when receive message is done but consensus is not done -> rNone\n")

	sr.OnSendMessage = SndWithSuccess
	sr.OnReceivedMessage = RcvWithSuccess

	sr.cns.ChRcvMsg <- []byte("Message has come")

	r := sr.doComitmentHash(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, rNone, r)
}

func TestSRComitmentHash_DoWork10(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("10: Test case when receive message and consensus is done -> true\n")

	sr.OnSendMessage = SndWithoutSuccess
	sr.OnReceivedMessage = RcvWithSuccess

	sr.cns.ChRcvMsg <- []byte("Message has come")

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, SsFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, true, r)
}

func TestSRComitmentHash_DoWork11(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("11: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnSendMessage = SndWithoutSuccess
	sr.OnReceivedMessage = RcvWithoutSuccessAndCancel

	sr.cns.ChRcvMsg <- []byte("Message has come")

	sr.cns.ValidationMap[sr.cns.Self].Block = true

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, SsNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSRComitmentHash_Current(t *testing.T) {
	sr := NewSRComitmentHash(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrComitmentHash), sr.Current())
}

func TestSRComitmentHash_Next(t *testing.T) {
	sr := NewSRComitmentHash(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(SrBitmap), sr.Next())
}

func TestSRComitmentHash_EndTime(t *testing.T) {
	sr := NewSRComitmentHash(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, int64(100*roundTimeDuration/100), sr.EndTime())
}

func TestSRComitmentHash_Name(t *testing.T) {
	sr := NewSRComitmentHash(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, "<COMITMENT_HASH>", sr.Name())
}

func TestSRComitmentHash_Log(t *testing.T) {
	sr := NewSRComitmentHash(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	sr.Log("Test SRComitmentHash")
}
