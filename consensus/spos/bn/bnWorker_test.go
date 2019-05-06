package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

func TestWorker_InitReceivedMessagesShouldWork(t *testing.T) {
	bnService, _ := bn.NewConsensusService()
	messages := bnService.InitReceivedMessages()

	receivedMessages := make(map[consensus.MessageType][]*consensus.Message)
	receivedMessages[bn.MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[bn.MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[bn.MtCommitmentHash] = make([]*consensus.Message, 0)
	receivedMessages[bn.MtBitmap] = make([]*consensus.Message, 0)
	receivedMessages[bn.MtCommitment] = make([]*consensus.Message, 0)
	receivedMessages[bn.MtSignature] = make([]*consensus.Message, 0)

	assert.Equal(t, len(receivedMessages), len(messages))
	assert.NotNil(t, messages[bn.MtBlockBody])
	assert.NotNil(t, messages[bn.MtBlockHeader])
	assert.NotNil(t, messages[bn.MtCommitmentHash])
	assert.NotNil(t, messages[bn.MtBitmap])
	assert.NotNil(t, messages[bn.MtCommitment])
	assert.NotNil(t, messages[bn.MtSignature])
}

func TestWorker_GetMessageRangeShouldWork(t *testing.T) {
	var v []consensus.MessageType
	bnService, _ := bn.NewConsensusService()

	messagesRange := bnService.GetMessageRange()
	for i := bn.MtBlockBody; i <= bn.MtSignature; i++ {
		v = append(v, i)
	}

	assert.NotNil(t, messagesRange)

	for i, val := range messagesRange {
		assert.Equal(t, v[i], val)
	}
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockBodyShouldWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrStartRound, spos.SsFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtBlockBody)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockBodyShouldNotWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrStartRound, spos.SsNotFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtBlockBody)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockHeaderShouldWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrStartRound, spos.SsFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtBlockHeader)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockHeaderShouldNotWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrStartRound, spos.SsNotFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtBlockHeader)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockFinishedForMtCommitmentHashShouldWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBlock, spos.SsFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtCommitmentHash)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockNotFinishedForMtCommitmentHashShouldNotWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBlock, spos.SsNotFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtCommitmentHash)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockFinishedForMtBitmapShouldWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBlock, spos.SsFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtBitmap)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockNotFinishedForMtBitmaphouldNotWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBlock, spos.SsNotFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtBitmap)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrBitmapFinishedForMtCommitmentShouldWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBitmap, spos.SsFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtCommitment)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrBitmapNotFinishedForMtCommitmentShouldNotWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtCommitment)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrBitmapFinishedMtSignatureShouldWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBitmap, spos.SsFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtSignature)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSSrBitmapNotFinishedForMtSignatureShouldNotWork(t *testing.T) {

	bnService, _ := bn.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bn.SrBitmap, spos.SsNotFinished)

	canProceed := bnService.CanProceed(consensusState, bn.MtSignature)
	assert.False(t, canProceed)
}

func TestWorker_GetSubroundName(t *testing.T) {
	t.Parallel()

	service, _ := bn.NewConsensusService()

	r := service.GetSubroundName(bn.SrStartRound)
	assert.Equal(t, "(START_ROUND)", r)
	r = service.GetSubroundName(bn.SrBlock)
	assert.Equal(t, "(BLOCK)", r)
	r = service.GetSubroundName(bn.SrCommitmentHash)
	assert.Equal(t, "(COMMITMENT_HASH)", r)
	r = service.GetSubroundName(bn.SrBitmap)
	assert.Equal(t, "(BITMAP)", r)
	r = service.GetSubroundName(bn.SrCommitment)
	assert.Equal(t, "(COMMITMENT)", r)
	r = service.GetSubroundName(bn.SrSignature)
	assert.Equal(t, "(SIGNATURE)", r)
	r = service.GetSubroundName(bn.SrEndRound)
	assert.Equal(t, "(END_ROUND)", r)
	r = service.GetSubroundName(-1)
	assert.Equal(t, "Undefined subround", r)
}
