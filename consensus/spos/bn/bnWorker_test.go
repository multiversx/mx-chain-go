package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/stretchr/testify/assert"
)

const roundTimeDuration = time.Duration(100 * time.Millisecond)

func createEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, string(i+65))
	}
	return eligibleList
}

func initConsensusState() *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := createEligibleList(consensusGroupSize)
	indexLeader := 1
	rcns := spos.NewRoundConsensus(
		eligibleList,
		consensusGroupSize,
		eligibleList[indexLeader])

	rcns.SetConsensusGroup(eligibleList)
	rcns.ResetRoundState()

	PBFTThreshold := consensusGroupSize*2/3 + 1

	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(1, 1)
	rthr.SetThreshold(2, PBFTThreshold)
	rthr.SetThreshold(3, PBFTThreshold)
	rthr.SetThreshold(4, PBFTThreshold)
	rthr.SetThreshold(5, PBFTThreshold)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	cns.Data = []byte("X")
	cns.RoundIndex = 0
	return cns
}

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

func TestWorker_IsMessageWithBlockHeader(t *testing.T) {
	t.Parallel()

	service, _ := bn.NewConsensusService()

	ret := service.IsMessageWithBlockHeader(bn.MtBlockBody)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockHeader(bn.MtBlockHeader)
	assert.True(t, ret)
}
