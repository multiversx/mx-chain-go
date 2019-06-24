package bls_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
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
	t.Parallel()

	bnService, _ := bls.NewConsensusService()
	messages := bnService.InitReceivedMessages()

	receivedMessages := make(map[consensus.MessageType][]*consensus.Message)
	receivedMessages[bls.MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtSignature] = make([]*consensus.Message, 0)

	assert.Equal(t, len(receivedMessages), len(messages))
	assert.NotNil(t, messages[bls.MtBlockBody])
	assert.NotNil(t, messages[bls.MtBlockHeader])
	assert.NotNil(t, messages[bls.MtSignature])
}

func TestWorker_GetMessageRangeShouldWork(t *testing.T) {
	t.Parallel()

	var v []consensus.MessageType
	blsService, _ := bls.NewConsensusService()

	messagesRange := blsService.GetMessageRange()
	for i := bls.MtBlockBody; i <= bls.MtSignature; i++ {
		v = append(v, i)
	}

	assert.NotNil(t, messagesRange)

	for i, val := range messagesRange {
		assert.Equal(t, v[i], val)
	}
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockBodyShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrStartRound, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockBody)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockBodyShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrStartRound, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockBody)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockHeaderShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrStartRound, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockHeader)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockHeaderShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrStartRound, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockHeader)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockFinishedForMtBlockHeaderShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrBlock, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtSignature)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockRoundNotFinishedForMtBlockHeaderShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrBlock, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtSignature)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWitUnkownMessageTypeShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()
	consensusState := initConsensusState()

	canProceed := blsService.CanProceed(consensusState, -1)
	assert.False(t, canProceed)
}

func TestWorker_GetSubroundName(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	r := service.GetSubroundName(bls.SrStartRound)
	assert.Equal(t, "(START_ROUND)", r)
	r = service.GetSubroundName(bls.SrBlock)
	assert.Equal(t, "(BLOCK)", r)
	r = service.GetSubroundName(bls.SrSignature)
	assert.Equal(t, "(SIGNATURE)", r)
	r = service.GetSubroundName(bls.SrEndRound)
	assert.Equal(t, "(END_ROUND)", r)
	r = service.GetSubroundName(-1)
	assert.Equal(t, "Undefined subround", r)
}

func TestWorker_GetStringValue(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	r := service.GetStringValue(bls.MtBlockBody)
	assert.Equal(t, bls.BlockBodyStringValue, r)
	r = service.GetStringValue(bls.MtBlockHeader)
	assert.Equal(t, bls.BlockHeaderStringValue, r)
	r = service.GetStringValue(bls.MtSignature)
	assert.Equal(t, bls.BlockSignatureStringValue, r)
	r = service.GetStringValue(bls.MtUnknown)
	assert.Equal(t, bls.BlockUnknownStringValue, r)
	r = service.GetStringValue(-1)
	assert.Equal(t, bls.BlockDefaultStringValue, r)
}

func TestWorker_IsMessageWithBlockHeader(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsMessageWithBlockHeader(bls.MtBlockBody)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockHeader(bls.MtBlockHeader)
	assert.True(t, ret)
}
