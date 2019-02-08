package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func initConsensusState() *spos.ConsensusState {
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

	rthr := spos.NewRoundThreshold()

	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, 3)
	rthr.SetThreshold(bn.SrBitmap, 3)
	rthr.SetThreshold(bn.SrCommitment, 3)
	rthr.SetThreshold(bn.SrSignature, 3)

	rstatus := spos.NewRoundStatus()

	rstatus.SetStatus(bn.SrBlock, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitmentHash, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrBitmap, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrCommitment, spos.SsNotFinished)
	rstatus.SetStatus(bn.SrSignature, spos.SsNotFinished)

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	return cns
}

func TestNewSpos(t *testing.T) {
	cns := initConsensusState()
	assert.NotNil(t, cns)
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

	cns := spos.NewConsensusState(
		rcns,
		nil,
		nil,
	)

	assert.Equal(t, true, cns.IsNodeLeaderInCurrentRound("1"))
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

	cns2 := spos.NewConsensusState(
		rcns1,
		nil,
		nil,
	)

	cns3 := spos.NewConsensusState(
		rcns2,
		nil,
		nil,
	)

	cns4 := spos.NewConsensusState(
		rcns3,
		nil,
		nil,
	)

	leader, err := cns2.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "consensusGroup is null", err.Error())
	assert.Equal(t, "", leader)

	leader, err = cns3.GetLeader()
	assert.NotNil(t, err)
	assert.Equal(t, "consensusGroup is empty", err.Error())
	assert.Equal(t, "", leader)

	leader, err = cns4.GetLeader()
	assert.Nil(t, err)
	assert.Equal(t, "1", leader)
}

func TestSpos_IsSelfLeaderInCurrentRoundShouldReturnFalse(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	cns := spos.NewConsensusState(
		rcns,
		nil,
		nil,
	)

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestSpos_IsSelfLeaderInCurrentRoundShouldReturnTrue(t *testing.T) {
	eligibleList := []string{"1", "2", "3"}

	rcns := spos.NewRoundConsensus(
		eligibleList,
		3,
		"2")

	rcns.SetConsensusGroup(eligibleList)

	cns := spos.NewConsensusState(
		rcns,
		nil,
		nil,
	)

	assert.False(t, cns.IsSelfLeaderInCurrentRound())
}

func TestWorker_IsConsensusDataNotSetShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	cns.Data = make([]byte, 0)

	assert.False(t, cns.IsConsensusDataNotSet())
}

func TestWorker_IsConsensusDataNotSetShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.Data = nil

	assert.True(t, cns.IsConsensusDataNotSet())
}

func TestWorker_IsConsensusDataAlreadySetShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	cns.Data = nil

	assert.False(t, cns.IsConsensusDataAlreadySet())
}

func TestWorker_IsConsensusDataAlreadySetShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.Data = make([]byte, 0)

	assert.True(t, cns.IsConsensusDataAlreadySet())
}

func TestWorker_IsSelfJobDoneShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))

	cns.SetJobDone(cns.SelfPubKey(), bn.SrCommitment, true)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))

	cns.SetJobDone(cns.SelfPubKey()+"X", bn.SrBlock, true)
	assert.False(t, cns.IsSelfJobDone(bn.SrBlock))
}

func TestWorker_IsSelfJobDoneShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)

	assert.True(t, cns.IsSelfJobDone(bn.SrBlock))
}

func TestWorker_IsJobDoneShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	cns.SetJobDone("1", bn.SrBlock, false)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))

	cns.SetJobDone("1", bn.SrCommitment, true)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))

	cns.SetJobDone("2", bn.SrBlock, true)
	assert.False(t, cns.IsJobDone("1", bn.SrBlock))
}

func TestWorker_IsJobDoneShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.SetJobDone("1", bn.SrBlock, true)

	assert.True(t, cns.IsJobDone("1", bn.SrBlock))
}

func TestWorker_IsCurrentSubroundFinishedShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	cns.SetStatus(bn.SrBlock, spos.SsNotFinished)
	assert.False(t, cns.IsCurrentSubroundFinished(bn.SrBlock))

	cns.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.False(t, cns.IsCurrentSubroundFinished(bn.SrBlock))

}

func TestWorker_IsCurrentSubroundFinishedShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.SetStatus(bn.SrBlock, spos.SsFinished)
	assert.True(t, cns.IsCurrentSubroundFinished(bn.SrBlock))
}

func TestWorker_IsMessageReceivedFromItselfShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	assert.False(t, cns.IsMessageReceivedFromItself(cns.SelfPubKey()+"X"))
}

func TestWorker_IsMessageReceivedFromItselfShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	assert.True(t, cns.IsMessageReceivedFromItself(cns.SelfPubKey()))
}

func TestWorker_IsMessageReceivedTooLateShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	assert.False(t, cns.IsMessageReceivedTooLate())
}

func TestWorker_IsMessageReceivedForOtherRoundShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	assert.False(t, cns.IsMessageReceivedForOtherRound(3, 3))
}

func TestWorker_IsMessageReceivedForOtherRoundShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	assert.True(t, cns.IsMessageReceivedForOtherRound(1, 0))
}

func TestWorker_IsBlockBodyAlreadyReceivedShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	cns.BlockBody = nil

	assert.False(t, cns.IsBlockBodyAlreadyReceived())
}

func TestWorker_IsBlockBodyAlreadyReceivedShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.BlockBody = &block.TxBlockBody{}

	assert.True(t, cns.IsBlockBodyAlreadyReceived())
}

func TestWorker_IsHeaderAlreadyReceivedShouldReturnFalse(t *testing.T) {
	cns := initConsensusState()

	cns.Header = nil

	assert.False(t, cns.IsHeaderAlreadyReceived())
}

func TestWorker_IsHeaderAlreadyReceivedShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.Header = &block.Header{}

	assert.True(t, cns.IsHeaderAlreadyReceived())
}

func TestWorker_CanDoSubroundJobShouldReturnFalseWhenConsensusDataNotSet(t *testing.T) {
	cns := initConsensusState()

	cns.Data = nil

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanDoSubroundJobShouldReturnFalseWhenSelfJobIsDone(t *testing.T) {
	cns := initConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, true)

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanDoSubroundJobShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	cns := initConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	cns.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanDoSubroundJobShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cns.Data = make([]byte, 0)
	cns.SetJobDone(cns.SelfPubKey(), bn.SrBlock, false)
	cns.SetStatus(bn.SrBlock, spos.SsNotFinished)

	assert.True(t, cns.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedFromItself(t *testing.T) {
	cns := initConsensusState()

	cnsDta := &spos.ConsensusData{
		RoundIndex: 0,
		PubKey:     []byte(cns.SelfPubKey()),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestWorker_CanProcessReceivedMessageShouldReturnFalseWhenMessageIsReceivedForOtherRound(t *testing.T) {
	cns := initConsensusState()

	cnsDta := &spos.ConsensusData{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 1, bn.SrBlock))
}

func TestWorker_CanProcessReceivedMessageShouldReturnFalseWhenJobIsDone(t *testing.T) {
	cns := initConsensusState()

	cnsDta := &spos.ConsensusData{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	cns.SetJobDone("1", bn.SrBlock, true)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestWorker_CanProcessReceivedMessageShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	cns := initConsensusState()

	cnsDta := &spos.ConsensusData{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	cns.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}

func TestWorker_CanProcessReceivedMessageShouldReturnTrue(t *testing.T) {
	cns := initConsensusState()

	cnsDta := &spos.ConsensusData{
		RoundIndex: 0,
		PubKey:     []byte("1"),
	}

	assert.True(t, cns.CanProcessReceivedMessage(cnsDta, 0, bn.SrBlock))
}
