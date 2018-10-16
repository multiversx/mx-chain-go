package spos

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func InitSREndRound() (*chronology.Chronology, *SREndRound) {
	genesisTime := time.Now()
	currentTime := genesisTime

	rnd := chronology.NewRound(genesisTime, currentTime, roundTimeDuration)

	chr := chronology.NewChronology(true, true, rnd, genesisTime, &ntp.LocalTime{ClockOffset: 0})

	vld := NewValidators([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, "2")
	pbft := 2*len(vld.ConsensusGroup)/3 + 1
	th := NewThreshold(1, pbft, pbft, pbft, pbft)
	rs := NewRoundStatus(ssNotFinished, ssNotFinished, ssNotFinished, ssNotFinished, ssNotFinished)

	cns := NewConsensus(true, vld, th, rs, chr)
	cns.Block = block.NewBlock(-1, "", "", "", "", "")
	cns.BlockChain = blockchain.NewBlockChain(nil, chr.GetSyncTimer(), true)

	sr := NewSREndRound(true, int64(100*roundTimeDuration/100), cns, nil, nil)
	chr.AddSubrounder(sr)

	return chr, sr
}

func TestNewSREndRound(t *testing.T) {
	sr := NewSREndRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.NotNil(t, sr)
}

func TestSREndRound_DoWork1(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("1: Test case when consensus is not done -> rNone\n")

	bActionDone := true
	r := sr.doEndRound(chr, &bActionDone)

	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Comitment)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rNone, r)
}

func TestSREndRound_DoWork2(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("2: Test case when consensus is semi-done (only subround BLOCK, COMITMENT_HASH, BITMAP and COMITMENT is done) -> rNone\n")

	rv := sr.cns.ValidationMap[sr.cns.Self]
	rv.Block = true
	sr.cns.ValidationMap[sr.cns.Self] = rv

	for i := 0; i < len(sr.cns.ConsensusGroup); i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.ComitmentHash = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	for i := 0; i < sr.cns.Threshold.Bitmap; i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.Bitmap = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	for i := 0; i < sr.cns.Threshold.Comitment; i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.Comitment = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	bActionDone := true
	r := sr.doEndRound(chr, &bActionDone)

	assert.Equal(t, ssFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, ssFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, ssFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, ssFinished, sr.cns.RoundStatus.Comitment)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rNone, r)
}

func TestSREndRound_DoWork3(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("3: Test case when consensus is done and I am the leader -> rTrue\n")

	sr.cns.Self = "1"

	sr.cns.RoundStatus.Block = ssFinished
	sr.cns.RoundStatus.ComitmentHash = ssFinished
	sr.cns.RoundStatus.Bitmap = ssFinished
	sr.cns.RoundStatus.Comitment = ssFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.Comitment = true
		rv.Signature = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	bActionDone := true
	r := sr.doEndRound(chr, &bActionDone)

	assert.Equal(t, ssFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rTrue, r)
}

func TestSREndRound_DoWork4(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("3: Test case when consensus is done and I am NOT the leader -> rTrue\n")

	sr.cns.RoundStatus.Block = ssFinished
	sr.cns.RoundStatus.ComitmentHash = ssFinished
	sr.cns.RoundStatus.Bitmap = ssFinished
	sr.cns.RoundStatus.Comitment = ssFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.Comitment = true
		rv.Signature = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	bActionDone := true
	r := sr.doEndRound(chr, &bActionDone)

	assert.Equal(t, ssFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rTrue, r)
}

func TestSREndRound_DoWork5(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("4: Test case when time has expired -> rTrue\n")

	chr.SetClockOffset(time.Duration(sr.endTime + 1))

	bActionDone := true
	r := sr.doEndRound(chr, &bActionDone)

	assert.Equal(t, r, rTrue)
}

func TestSREndRound_DoWork6(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("5: Test case when receive message is done but consensus is not done -> rNone\n")

	sr.OnReceivedMessage = RcvWithSuccess

	sr.cns.ChRcvMsg <- []byte("Message has come")

	bActionDone := false
	r := sr.doEndRound(chr, &bActionDone)

	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Block)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.ComitmentHash)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Bitmap)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Comitment)
	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, rNone, r)
}

func TestSREndRound_DoWork7(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("6: Test case when receive message and consensus is done and I am the leader -> true\n")

	sr.cns.Self = "1"

	sr.OnReceivedMessage = RcvWithSuccess

	sr.cns.ChRcvMsg <- []byte("Message has come")

	sr.cns.RoundStatus.Block = ssFinished
	sr.cns.RoundStatus.ComitmentHash = ssFinished
	sr.cns.RoundStatus.Bitmap = ssFinished
	sr.cns.RoundStatus.Comitment = ssFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.Comitment = true
		rv.Signature = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	r := sr.DoWork(chr)

	assert.Equal(t, ssFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, true, r)
}

func TestSREndRound_DoWork8(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("6: Test case when receive message and consensus is done and I am NOT the leader -> true\n")

	sr.OnReceivedMessage = RcvWithSuccess

	sr.cns.ChRcvMsg <- []byte("Message has come")

	sr.cns.RoundStatus.Block = ssFinished
	sr.cns.RoundStatus.ComitmentHash = ssFinished
	sr.cns.RoundStatus.Bitmap = ssFinished
	sr.cns.RoundStatus.Comitment = ssFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.Comitment = true
		rv.Signature = true
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	r := sr.DoWork(chr)

	assert.Equal(t, ssFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, true, r)
}

func TestSREndRound_DoWork9(t *testing.T) {

	chr, sr := InitSREndRound()

	fmt.Printf("7: Test case when receive message is not done and round should be canceled -> false\n")

	sr.OnReceivedMessage = RcvWithoutSuccessAndCancel

	sr.cns.ChRcvMsg <- []byte("Message has come")

	sr.cns.RoundStatus.Block = ssFinished
	sr.cns.RoundStatus.ComitmentHash = ssFinished
	sr.cns.RoundStatus.Bitmap = ssFinished
	sr.cns.RoundStatus.Comitment = ssFinished

	for i := 0; i < sr.cns.Threshold.Signature; i++ {
		rv := sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]]
		rv.Comitment = true
		if sr.cns.ConsensusGroup[i] != sr.cns.Self {
			rv.Signature = true
		}
		sr.cns.ValidationMap[sr.cns.ConsensusGroup[i]] = rv
	}

	r := sr.DoWork(chr)

	assert.Equal(t, ssNotFinished, sr.cns.RoundStatus.Signature)
	assert.Equal(t, chronology.SrCanceled, chr.GetSelfSubround())
	assert.Equal(t, false, r)
}

func TestSREndRound_Current(t *testing.T) {
	sr := NewSREndRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(srEndRound), sr.Current())
}

func TestSREndRound_Next(t *testing.T) {
	sr := NewSREndRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, chronology.Subround(srStartRound), sr.Next())
}

func TestSREndRound_EndTime(t *testing.T) {
	sr := NewSREndRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, int64(100*roundTimeDuration/100), sr.EndTime())
}

func TestSREndRound_Name(t *testing.T) {
	sr := NewSREndRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	assert.Equal(t, "<END_ROUND>", sr.Name())
}

func TestSREndRound_Log(t *testing.T) {
	sr := NewSREndRound(true, int64(100*roundTimeDuration/100), nil, nil, nil)
	sr.Log("Test SREndRound")
}
