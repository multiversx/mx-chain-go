package bls_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func createEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, string([]byte{byte(i + 65)}))
	}
	return eligibleList
}

func initConsensusState() *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := createEligibleList(consensusGroupSize)

	eligibleNodesPubKeys := make(map[string]struct{})
	for _, key := range eligibleList {
		eligibleNodesPubKeys[key] = struct{}{}
	}

	indexLeader := 1
	rcns := spos.NewRoundConsensus(
		eligibleNodesPubKeys,
		consensusGroupSize,
		eligibleList[indexLeader])

	rcns.SetConsensusGroup(eligibleList)
	rcns.ResetRoundState()

	pBFTThreshold := consensusGroupSize*2/3 + 1
	pBFTFallbackThreshold := consensusGroupSize*1/2 + 1

	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(1, 1)
	rthr.SetThreshold(2, pBFTThreshold)
	rthr.SetFallbackThreshold(1, 1)
	rthr.SetFallbackThreshold(2, pBFTFallbackThreshold)

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

func TestWorker_NewConsensusServiceShouldWork(t *testing.T) {
	t.Parallel()

	service, err := bls.NewConsensusService()
	assert.Nil(t, err)
	assert.False(t, check.IfNil(service))
}

func TestWorker_InitReceivedMessagesShouldWork(t *testing.T) {
	t.Parallel()

	bnService, _ := bls.NewConsensusService()
	messages := bnService.InitReceivedMessages()

	receivedMessages := make(map[consensus.MessageType][]*consensus.Message)
	receivedMessages[bls.MtBlockBodyAndHeader] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtSignature] = make([]*consensus.Message, 0)
	receivedMessages[bls.MtBlockHeaderFinalInfo] = make([]*consensus.Message, 0)

	assert.Equal(t, len(receivedMessages), len(messages))
	assert.NotNil(t, messages[bls.MtBlockBodyAndHeader])
	assert.NotNil(t, messages[bls.MtBlockBody])
	assert.NotNil(t, messages[bls.MtBlockHeader])
	assert.NotNil(t, messages[bls.MtSignature])
	assert.NotNil(t, messages[bls.MtBlockHeaderFinalInfo])
}

func TestWorker_GetMessageRangeShouldWork(t *testing.T) {
	t.Parallel()

	v := make([]consensus.MessageType, 0)
	blsService, _ := bls.NewConsensusService()

	messagesRange := blsService.GetMessageRange()
	assert.NotNil(t, messagesRange)

	for i := bls.MtBlockBodyAndHeader; i <= bls.MtBlockHeaderFinalInfo; i++ {
		v = append(v, i)
	}
	assert.NotNil(t, v)

	for i, val := range messagesRange {
		assert.Equal(t, v[i], val)
	}
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockBodyAndHeaderShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrStartRound, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockBodyAndHeader)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockBodyAndHeaderShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrStartRound, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockBodyAndHeader)
	assert.False(t, canProceed)
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

func TestWorker_CanProceedWithSrSignatureFinishedForMtBlockHeaderFinalInfoShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrSignature, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockHeaderFinalInfo)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrSignatureRoundNotFinishedForMtBlockHeaderFinalInfoShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := bls.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(bls.SrSignature, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, bls.MtBlockHeaderFinalInfo)
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

	r := service.GetStringValue(bls.MtBlockBodyAndHeader)
	assert.Equal(t, bls.BlockBodyAndHeaderStringValue, r)
	r = service.GetStringValue(bls.MtBlockBody)
	assert.Equal(t, bls.BlockBodyStringValue, r)
	r = service.GetStringValue(bls.MtBlockHeader)
	assert.Equal(t, bls.BlockHeaderStringValue, r)
	r = service.GetStringValue(bls.MtSignature)
	assert.Equal(t, bls.BlockSignatureStringValue, r)
	r = service.GetStringValue(bls.MtBlockHeaderFinalInfo)
	assert.Equal(t, bls.BlockHeaderFinalInfoStringValue, r)
	r = service.GetStringValue(bls.MtUnknown)
	assert.Equal(t, bls.BlockUnknownStringValue, r)
	r = service.GetStringValue(-1)
	assert.Equal(t, bls.BlockDefaultStringValue, r)
}

func TestWorker_IsMessageWithBlockBodyAndHeader(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsMessageWithBlockBodyAndHeader(bls.MtBlockBody)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockBodyAndHeader(bls.MtBlockHeader)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockBodyAndHeader(bls.MtBlockBodyAndHeader)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithBlockBody(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsMessageWithBlockBody(bls.MtBlockHeader)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockBody(bls.MtBlockBody)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithBlockHeader(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsMessageWithBlockHeader(bls.MtBlockBody)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockHeader(bls.MtBlockHeader)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithSignature(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsMessageWithSignature(bls.MtBlockBodyAndHeader)
	assert.False(t, ret)

	ret = service.IsMessageWithSignature(bls.MtSignature)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithFinalInfo(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsMessageWithFinalInfo(bls.MtSignature)
	assert.False(t, ret)

	ret = service.IsMessageWithFinalInfo(bls.MtBlockHeaderFinalInfo)
	assert.True(t, ret)
}

func TestWorker_IsSubroundSignature(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsSubroundSignature(bls.SrEndRound)
	assert.False(t, ret)

	ret = service.IsSubroundSignature(bls.SrSignature)
	assert.True(t, ret)
}

func TestWorker_IsSubroundStartRound(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsSubroundStartRound(bls.SrSignature)
	assert.False(t, ret)

	ret = service.IsSubroundStartRound(bls.SrStartRound)
	assert.True(t, ret)
}

func TestWorker_IsMessageTypeValid(t *testing.T) {
	t.Parallel()

	service, _ := bls.NewConsensusService()

	ret := service.IsMessageTypeValid(bls.MtBlockBody)
	assert.True(t, ret)

	ret = service.IsMessageTypeValid(666)
	assert.False(t, ret)
}
