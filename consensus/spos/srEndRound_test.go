package spos_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"

	//"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func InitSREndRound() (*chronology.Chronology, *spos.SREndRound) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, spos.RoundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{})

	vld := spos.NewValidators(nil, nil, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := spos.NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := spos.NewRoundStatus(spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished, spos.SsNotFinished)

	cns := spos.NewConsensus(true, vld, th, rs, chr)
	//Cns.block = block.NewBlock(-1, "", "", "", "", "")
	//Cns.blockChain = blockchain.NewBlockChain(nil, Chr.SyncTime(), true)

	sr := spos.NewSREndRound(true, int64(100*spos.RoundTimeDuration/100), cns, OnEndRound)
	chr.AddSubround(sr)

	return chr, sr
}

func TestNewSREndRound(t *testing.T) {
	sr := spos.NewSREndRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.NotNil(t, sr)
}

func TestSREndRound_DoWork1(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("1: Test case when consensus is not done -> RNone\n")

	sr.Cns.SetReceivedMessage(true)
	r := sr.DoEndRound(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Comitment)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, spos.RNone, r)
}

func TestSREndRound_DoWork2(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("2: Test case when consensus is semi-done (only subround BLOCK, COMITMENT_HASH, BITMAP and COMITMENT is done) -> RNone\n")

	sr.Cns.ValidationMap[sr.Cns.Self].Block = true

	for i := 0; i < len(sr.Cns.ConsensusGroup); i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].ComitmentHash = true
	}

	for i := 0; i < sr.Cns.Threshold.Bitmap; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
	}

	for i := 0; i < sr.Cns.Threshold.Comitment; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Comitment = true
	}

	sr.Cns.SetReceivedMessage(true)
	r := sr.DoEndRound(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Comitment)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, spos.RNone, r)
}

func TestSREndRound_DoWork3(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("3: Test case when consensus is done and I am the leader -> RTrue\n")

	sr.Cns.Self = "1"

	sr.Cns.RoundStatus.Block = spos.SsFinished
	sr.Cns.RoundStatus.ComitmentHash = spos.SsFinished
	sr.Cns.RoundStatus.Bitmap = spos.SsFinished
	sr.Cns.RoundStatus.Comitment = spos.SsFinished

	for i := 0; i < sr.Cns.Threshold.Signature; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Signature = true
	}

	sr.Cns.SetReceivedMessage(true)
	r := sr.DoEndRound(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, spos.RTrue, r)
}

func TestSREndRound_DoWork4(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("3: Test case when consensus is done and I am NOT the leader -> RTrue\n")

	sr.Cns.RoundStatus.Block = spos.SsFinished
	sr.Cns.RoundStatus.ComitmentHash = spos.SsFinished
	sr.Cns.RoundStatus.Bitmap = spos.SsFinished
	sr.Cns.RoundStatus.Comitment = spos.SsFinished

	for i := 0; i < sr.Cns.Threshold.Signature; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Signature = true
	}

	sr.Cns.SetReceivedMessage(true)
	r := sr.DoEndRound(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, spos.RTrue, r)
}

func TestSREndRound_DoWork5(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("4: Test case when time has expired -> RTrue\n")

	chr.SetClockOffset(time.Duration(sr.EndTime() + 1))

	sr.Cns.SetReceivedMessage(true)
	r := sr.DoEndRound(chr)

	assert.Equal(t, r, spos.RTrue)
}

func TestSREndRound_DoWork6(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("5: Test case when receive message is done but consensus is not done -> RNone\n")

	sr.Cns.SetReceivedMessage(true)
	r := sr.DoEndRound(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Block)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.ComitmentHash)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Bitmap)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Comitment)
	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, spos.RNone, r)
}

func TestSREndRound_DoWork7(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("6: Test case when receive message and consensus is done and I am the leader -> true\n")

	sr.Cns.Self = "1"

	sr.Cns.SetReceivedMessage(true)

	sr.Cns.RoundStatus.Block = spos.SsFinished
	sr.Cns.RoundStatus.ComitmentHash = spos.SsFinished
	sr.Cns.RoundStatus.Bitmap = spos.SsFinished
	sr.Cns.RoundStatus.Comitment = spos.SsFinished

	for i := 0; i < sr.Cns.Threshold.Signature; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Signature = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, true, r)
}

func TestSREndRound_DoWork8(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("6: Test case when receive message and consensus is done and I am NOT the leader -> true\n")

	sr.Cns.SetReceivedMessage(true)

	sr.Cns.RoundStatus.Block = spos.SsFinished
	sr.Cns.RoundStatus.ComitmentHash = spos.SsFinished
	sr.Cns.RoundStatus.Bitmap = spos.SsFinished
	sr.Cns.RoundStatus.Comitment = spos.SsFinished

	for i := 0; i < sr.Cns.Threshold.Signature; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Signature = true
	}

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, true, r)
}

func TestSREndRound_DoWork9(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("7: Test case when receive message is not done and round should be canceled -> false\n")

	sr.Cns.SetReceivedMessage(false)
	sr.Cns.Chr.SetSelfSubround(chronology.SrCanceled)

	sr.Cns.RoundStatus.Block = spos.SsFinished
	sr.Cns.RoundStatus.ComitmentHash = spos.SsFinished
	sr.Cns.RoundStatus.Bitmap = spos.SsFinished
	sr.Cns.RoundStatus.Comitment = spos.SsFinished

	for i := 0; i < sr.Cns.Threshold.Signature; i++ {
		sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Bitmap = true
		if sr.Cns.ConsensusGroup[i] != sr.Cns.Self {
			sr.Cns.ValidationMap[sr.Cns.ConsensusGroup[i]].Signature = true
		}
	}

	r := sr.DoWork(chr)

	assert.Equal(t, spos.SsNotFinished, sr.Cns.RoundStatus.Signature)
	assert.Equal(t, chronology.SrCanceled, chr.SelfSubround())
	assert.Equal(t, false, r)
}

func TestSREndRound_Current(t *testing.T) {
	sr := spos.NewSREndRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrEndRound), sr.Current())
}

func TestSREndRound_Next(t *testing.T) {
	sr := spos.NewSREndRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, chronology.Subround(spos.SrStartRound), sr.Next())
}

func TestSREndRound_EndTime(t *testing.T) {
	sr := spos.NewSREndRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, int64(100*spos.RoundTimeDuration/100), sr.EndTime())
}

func TestSREndRound_Name(t *testing.T) {
	sr := spos.NewSREndRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	assert.Equal(t, "<END_ROUND>", sr.Name())
}

func TestSREndRound_Log(t *testing.T) {
	sr := spos.NewSREndRound(true, int64(100*spos.RoundTimeDuration/100), nil, nil)
	sr.Log("Test SREndRound")
}
