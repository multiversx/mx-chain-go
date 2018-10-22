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

func InitSRComitmentHash() (*chronology.Chronology, *spos.SRComitmentHash) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := spos.NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := spos.NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := spos.NewRoundStatus(spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished)

	cns := spos.NewConsensus(true, vld, th, rs, chr)

	sr := spos.NewSRComitmentHash(true, int64(100*spos.RoundTimeDuration/100), cns, nil)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSRComitmentHash(t *testing.T) {
	sr := spos.NewSRComitmentHash(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.NotNil(t, sr)
}

func TestSRComitmentHash_DoWork1(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("1: Test case when send message is done but consensus is not done -> RNone\n")

	sr.OnSendComitmentHash = SndWithSuccess

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.RNone, r)
}

func TestSRComitmentHash_DoWork2(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("2: Test case when send message is done but consensus is semi-done (only subround BLOCK is done) -> RNone\n")

	sr.OnSendComitmentHash = SndWithSuccess

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.RNone, r)
}

func TestSRComitmentHash_DoWork3(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("3: Test case when send message is done, but consensus (full) is not done (I AM NOT LEADER) -> RNone\n")

	sr.OnSendComitmentHash = SndWithSuccess

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	for i := 0; i < sr.Cns.Threshold.ComitmentHash; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.RNone, r)
}

func TestSRComitmentHash_DoWork4(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("4: Test case when send message is done and consensus (pbft) is enough (I AM LEADER) -> RTrue\n")

	sr.OnSendComitmentHash = SndWithSuccess

	sr.Cns.Self = "1"

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	for i := 0; i < sr.Cns.Threshold.ComitmentHash; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.RTrue, r)
}

func TestSRComitmentHash_DoWork5(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("5: Test case when send message is done and consensus (pbft) is enough because of the bitmap (I AM NOT LEADER) -> RTrue\n")

	sr.OnSendComitmentHash = SndWithSuccess

	sr.Cns.RoundStatus.Block = spos.SsFinished

	for i := 0; i < sr.Cns.Threshold.Bitmap; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
	}

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.RTrue, r)
}

func TestSRComitmentHash_DoWork6(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("6: Test case when send message and consensus (full) is done (I AM NOT LEADER) -> RTrue\n")

	sr.OnSendComitmentHash = SndWithSuccess

	sr.Cns.RoundStatus.Block = spos.SsFinished

	for i := 0; i < len(sr.Cns.ConsensusGroup); i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.RTrue, r)
}

func TestSRComitmentHash_DoWork7(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("7: Test case when time has expired with consensus (pbft) not done -> RTrue\n")

	sr.OnSendComitmentHash = SndWithSuccess

	chr.SetClockOffset(time.Duration(sr.EndTime() + 1))

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsExtended, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, r, spos.RTrue)
}

func TestSRComitmentHash_DoWork8(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("8: Test case when time has expired with consensus (pbft) done -> RTrue\n")

	sr.OnSendComitmentHash = SndWithoutSuccess

	chr.SetClockOffset(time.Duration(sr.EndTime() + 1))

	for i := 0; i < sr.Cns.Threshold.ComitmentHash; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsExtended, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, r, spos.RTrue)
}

func TestSRComitmentHash_DoWork9(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("9: Test case when receive message is done but consensus is not done -> RNone\n")

	sr.OnSendComitmentHash = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(true)

	r := sr.DoComitmentHash(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.RNone, r)
}

func TestSRComitmentHash_DoWork10(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("10: Test case when receive message and consensus is done -> true\n")

	sr.OnSendComitmentHash = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(true)

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	for i := 0; i < len(sr.Cns.ConsensusGroup); i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, true, r)
}

func TestSRComitmentHash_DoWork11(t *testing.T) {

	chr, sr := InitSRComitmentHash()

	fmt.Printf("11: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnSendComitmentHash = SndWithoutSuccess
	sr.Cns.SetReceivedMessage(false)
	sr.Cns.Chr.SetSelfSubround(chronology.SrCanceled)

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	for i := 0; i < len(sr.Cns.ConsensusGroup); i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSRComitmentHash_Current(t *testing.T) {
	sr := spos.NewSRComitmentHash(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrComitmentHash), sr.Current())
}

func TestSRComitmentHash_Next(t *testing.T) {
	sr := spos.NewSRComitmentHash(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrBitmap), sr.Next())
}

func TestSRComitmentHash_EndTime(t *testing.T) {
	sr := spos.NewSRComitmentHash(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, int64(100*spos.RoundTimeDuration/100), sr.EndTime())
}

func TestSRComitmentHash_Name(t *testing.T) {
	sr := spos.NewSRComitmentHash(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, "<COMITMENT_HASH>", sr.Name())
}

func TestSRComitmentHash_Log(t *testing.T) {
	sr := spos.NewSRComitmentHash(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	sr.Log("Test SRComitmentHash")
}
