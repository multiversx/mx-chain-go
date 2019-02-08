package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func TestWorker_IsConsensusDataNotSetShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = make([]byte, 0)

	assert.False(t, wrk.IsConsensusDataNotSet())
}

func TestWorker_IsConsensusDataNotSetShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = nil

	assert.True(t, wrk.IsConsensusDataNotSet())
}

func TestWorker_IsConsensusDataAlreadySetShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = nil

	assert.False(t, wrk.IsConsensusDataAlreadySet())
}

func TestWorker_IsConsensusDataAlreadySetShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = make([]byte, 0)

	assert.True(t, wrk.IsConsensusDataAlreadySet())
}

func TestWorker_IsSelfJobDoneShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, false)
	assert.False(t, wrk.IsSelfJobDone(bn.SrBlock))

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrCommitment, true)
	assert.False(t, wrk.IsSelfJobDone(bn.SrBlock))

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey()+"X", bn.SrBlock, true)
	assert.False(t, wrk.IsSelfJobDone(bn.SrBlock))
}

func TestWorker_IsSelfJobDoneShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, true)

	assert.True(t, wrk.IsSelfJobDone(bn.SrBlock))
}

func TestWorker_IsJobDoneShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetJobDone("A", bn.SrBlock, false)
	assert.False(t, wrk.IsJobDone("A", bn.SrBlock))

	wrk.SPoS.SetJobDone("A", bn.SrCommitment, true)
	assert.False(t, wrk.IsJobDone("A", bn.SrBlock))

	wrk.SPoS.SetJobDone("B", bn.SrBlock, true)
	assert.False(t, wrk.IsJobDone("A", bn.SrBlock))
}

func TestWorker_IsJobDoneShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetJobDone("A", bn.SrBlock, true)

	assert.True(t, wrk.IsJobDone("A", bn.SrBlock))
}

func TestWorker_IsCurrentRoundFinishedShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsNotFinished)
	assert.False(t, wrk.IsCurrentRoundFinished(bn.SrBlock))

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsExtended)
	assert.False(t, wrk.IsCurrentRoundFinished(bn.SrBlock))

	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	assert.False(t, wrk.IsCurrentRoundFinished(bn.SrBlock))

}

func TestWorker_IsCurrentRoundFinishedShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	assert.True(t, wrk.IsCurrentRoundFinished(bn.SrBlock))
}

func TestWorker_IsMessageReceivedFromItselfShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	assert.False(t, wrk.IsMessageReceivedFromItself(wrk.SPoS.SelfPubKey()+"X"))
}

func TestWorker_IsMessageReceivedFromItselfShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.IsMessageReceivedFromItself(wrk.SPoS.SelfPubKey()))
}

func TestWorker_IsMessageReceivedTooLateShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Chr.SetClockOffset(0)

	assert.False(t, wrk.IsMessageReceivedTooLate())
}

func TestWorker_IsMessageReceivedTooLateShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	endTime := getEndTime(wrk.SPoS.Chr, bn.SrEndRound)
	wrk.SPoS.Chr.SetClockOffset(time.Duration(endTime))

	assert.True(t, wrk.IsMessageReceivedTooLate())
}

func TestWorker_IsMessageReceivedForOtherRoundShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	assert.False(t, wrk.IsMessageReceivedForOtherRound(0))
}

func TestWorker_IsMessageReceivedForOtherRoundShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.IsMessageReceivedForOtherRound(1))
}

func TestWorker_IsBlockBodyAlreadyReceivedShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.BlockBody = nil

	assert.False(t, wrk.IsBlockBodyAlreadyReceived())
}

func TestWorker_IsBlockBodyAlreadyReceivedShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.BlockBody = &block.TxBlockBody{}

	assert.True(t, wrk.IsBlockBodyAlreadyReceived())
}

func TestWorker_IsHeaderAlreadyReceivedShouldReturnFalse(t *testing.T) {
	wrk := initWorker()

	wrk.Header = nil

	assert.False(t, wrk.IsHeaderAlreadyReceived())
}

func TestWorker_IsHeaderAlreadyReceivedShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.Header = &block.Header{}

	assert.True(t, wrk.IsHeaderAlreadyReceived())
}

func TestWorker_CanDoSubroundJobShouldReturnFalseWhenConsensusDataNotSet(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = nil

	assert.False(t, wrk.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanDoSubroundJobShouldReturnFalseWhenSelfJobIsDone(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = make([]byte, 0)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, true)

	assert.False(t, wrk.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanDoSubroundJobShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = make([]byte, 0)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, false)
	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, wrk.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanDoSubroundJobShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Data = make([]byte, 0)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, false)
	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsNotFinished)

	assert.True(t, wrk.CanDoSubroundJob(bn.SrBlock))
}

func TestWorker_CanReceiveMessageShouldReturnFalseWhenMessageIsReceivedFromItself(t *testing.T) {
	wrk := initWorker()

	assert.False(t, wrk.CanReceiveMessage(wrk.SPoS.SelfPubKey(), 0, bn.SrBlock))
}

func TestWorker_CanReceiveMessageShouldReturnFalseWhenMessageIsReceivedTooLate(t *testing.T) {
	wrk := initWorker()

	endTime := getEndTime(wrk.SPoS.Chr, bn.SrEndRound)
	wrk.SPoS.Chr.SetClockOffset(time.Duration(endTime))

	assert.False(t, wrk.CanReceiveMessage("A", 0, bn.SrBlock))
}

func TestWorker_CanReceiveMessageShouldReturnFalseWhenMessageIsReceivedForOtherRound(t *testing.T) {
	wrk := initWorker()

	assert.False(t, wrk.CanReceiveMessage("A", 1, bn.SrBlock))
}

func TestWorker_CanReceiveMessageShouldReturnFalseWhenJobIsDone(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetJobDone("A", bn.SrBlock, true)

	assert.False(t, wrk.CanReceiveMessage("A", 0, bn.SrBlock))
}

func TestWorker_CanReceiveMessageShouldReturnFalseWhenCurrentRoundIsFinished(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)

	assert.False(t, wrk.CanReceiveMessage("A", 0, bn.SrBlock))
}

func TestWorker_CanReceiveMessageShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	assert.True(t, wrk.CanReceiveMessage("A", 0, bn.SrBlock))
}
