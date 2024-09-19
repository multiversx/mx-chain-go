package v1_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	v1 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v1"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func createEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, string([]byte{byte(i + 65)}))
	}
	return eligibleList
}

func initConsensusState() *spos.ConsensusState {
	return initConsensusStateWithKeysHandler(&testscommon.KeysHandlerStub{})
}

func initConsensusStateWithKeysHandler(keysHandler consensus.KeysHandler) *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := createEligibleList(consensusGroupSize)

	eligibleNodesPubKeys := make(map[string]struct{})
	for _, key := range eligibleList {
		eligibleNodesPubKeys[key] = struct{}{}
	}

	indexLeader := 1
	rcns, _ := spos.NewRoundConsensus(
		eligibleNodesPubKeys,
		consensusGroupSize,
		eligibleList[indexLeader],
		keysHandler,
	)

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

	service, err := v1.NewConsensusService()
	assert.Nil(t, err)
	assert.False(t, check.IfNil(service))
}

func TestWorker_InitReceivedMessagesShouldWork(t *testing.T) {
	t.Parallel()

	bnService, _ := v1.NewConsensusService()
	messages := bnService.InitReceivedMessages()

	receivedMessages := make(map[consensus.MessageType][]*consensus.Message)
	receivedMessages[v1.MtBlockBodyAndHeader] = make([]*consensus.Message, 0)
	receivedMessages[v1.MtBlockBody] = make([]*consensus.Message, 0)
	receivedMessages[v1.MtBlockHeader] = make([]*consensus.Message, 0)
	receivedMessages[v1.MtSignature] = make([]*consensus.Message, 0)
	receivedMessages[v1.MtBlockHeaderFinalInfo] = make([]*consensus.Message, 0)
	receivedMessages[v1.MtInvalidSigners] = make([]*consensus.Message, 0)

	assert.Equal(t, len(receivedMessages), len(messages))
	assert.NotNil(t, messages[v1.MtBlockBodyAndHeader])
	assert.NotNil(t, messages[v1.MtBlockBody])
	assert.NotNil(t, messages[v1.MtBlockHeader])
	assert.NotNil(t, messages[v1.MtSignature])
	assert.NotNil(t, messages[v1.MtBlockHeaderFinalInfo])
	assert.NotNil(t, messages[v1.MtInvalidSigners])
}

func TestWorker_GetMessageRangeShouldWork(t *testing.T) {
	t.Parallel()

	v := make([]consensus.MessageType, 0)
	blsService, _ := v1.NewConsensusService()

	messagesRange := blsService.GetMessageRange()
	assert.NotNil(t, messagesRange)

	for i := v1.MtBlockBodyAndHeader; i <= v1.MtInvalidSigners; i++ {
		v = append(v, i)
	}
	assert.NotNil(t, v)

	for i, val := range messagesRange {
		assert.Equal(t, v[i], val)
	}
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockBodyAndHeaderShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrStartRound, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockBodyAndHeader)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockBodyAndHeaderShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrStartRound, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockBodyAndHeader)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockBodyShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrStartRound, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockBody)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockBodyShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrStartRound, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockBody)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundFinishedForMtBlockHeaderShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrStartRound, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockHeader)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrStartRoundNotFinishedForMtBlockHeaderShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrStartRound, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockHeader)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockFinishedForMtBlockHeaderShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrBlock, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtSignature)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrBlockRoundNotFinishedForMtBlockHeaderShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrBlock, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtSignature)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWithSrSignatureFinishedForMtBlockHeaderFinalInfoShouldWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrSignature, spos.SsFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockHeaderFinalInfo)
	assert.True(t, canProceed)
}

func TestWorker_CanProceedWithSrSignatureRoundNotFinishedForMtBlockHeaderFinalInfoShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()

	consensusState := initConsensusState()
	consensusState.SetStatus(v1.SrSignature, spos.SsNotFinished)

	canProceed := blsService.CanProceed(consensusState, v1.MtBlockHeaderFinalInfo)
	assert.False(t, canProceed)
}

func TestWorker_CanProceedWitUnkownMessageTypeShouldNotWork(t *testing.T) {
	t.Parallel()

	blsService, _ := v1.NewConsensusService()
	consensusState := initConsensusState()

	canProceed := blsService.CanProceed(consensusState, -1)
	assert.False(t, canProceed)
}

func TestWorker_GetSubroundName(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	r := service.GetSubroundName(v1.SrStartRound)
	assert.Equal(t, "(START_ROUND)", r)
	r = service.GetSubroundName(v1.SrBlock)
	assert.Equal(t, "(BLOCK)", r)
	r = service.GetSubroundName(v1.SrSignature)
	assert.Equal(t, "(SIGNATURE)", r)
	r = service.GetSubroundName(v1.SrEndRound)
	assert.Equal(t, "(END_ROUND)", r)
	r = service.GetSubroundName(-1)
	assert.Equal(t, "Undefined subround", r)
}

func TestWorker_GetStringValue(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	r := service.GetStringValue(v1.MtBlockBodyAndHeader)
	assert.Equal(t, v1.BlockBodyAndHeaderStringValue, r)
	r = service.GetStringValue(v1.MtBlockBody)
	assert.Equal(t, v1.BlockBodyStringValue, r)
	r = service.GetStringValue(v1.MtBlockHeader)
	assert.Equal(t, v1.BlockHeaderStringValue, r)
	r = service.GetStringValue(v1.MtSignature)
	assert.Equal(t, v1.BlockSignatureStringValue, r)
	r = service.GetStringValue(v1.MtBlockHeaderFinalInfo)
	assert.Equal(t, v1.BlockHeaderFinalInfoStringValue, r)
	r = service.GetStringValue(v1.MtUnknown)
	assert.Equal(t, v1.BlockUnknownStringValue, r)
	r = service.GetStringValue(-1)
	assert.Equal(t, v1.BlockDefaultStringValue, r)
}

func TestWorker_IsMessageWithBlockBodyAndHeader(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsMessageWithBlockBodyAndHeader(v1.MtBlockBody)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockBodyAndHeader(v1.MtBlockHeader)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockBodyAndHeader(v1.MtBlockBodyAndHeader)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithBlockBody(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsMessageWithBlockBody(v1.MtBlockHeader)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockBody(v1.MtBlockBody)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithBlockHeader(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsMessageWithBlockHeader(v1.MtBlockBody)
	assert.False(t, ret)

	ret = service.IsMessageWithBlockHeader(v1.MtBlockHeader)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithSignature(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsMessageWithSignature(v1.MtBlockBodyAndHeader)
	assert.False(t, ret)

	ret = service.IsMessageWithSignature(v1.MtSignature)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithFinalInfo(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsMessageWithFinalInfo(v1.MtSignature)
	assert.False(t, ret)

	ret = service.IsMessageWithFinalInfo(v1.MtBlockHeaderFinalInfo)
	assert.True(t, ret)
}

func TestWorker_IsMessageWithInvalidSigners(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsMessageWithInvalidSigners(v1.MtBlockHeaderFinalInfo)
	assert.False(t, ret)

	ret = service.IsMessageWithInvalidSigners(v1.MtInvalidSigners)
	assert.True(t, ret)
}

func TestWorker_IsSubroundSignature(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsSubroundSignature(v1.SrEndRound)
	assert.False(t, ret)

	ret = service.IsSubroundSignature(v1.SrSignature)
	assert.True(t, ret)
}

func TestWorker_IsSubroundStartRound(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsSubroundStartRound(v1.SrSignature)
	assert.False(t, ret)

	ret = service.IsSubroundStartRound(v1.SrStartRound)
	assert.True(t, ret)
}

func TestWorker_IsMessageTypeValid(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()

	ret := service.IsMessageTypeValid(v1.MtBlockBody)
	assert.True(t, ret)

	ret = service.IsMessageTypeValid(666)
	assert.False(t, ret)
}

func TestWorker_GetMaxNumOfMessageTypeAccepted(t *testing.T) {
	t.Parallel()

	service, _ := v1.NewConsensusService()
	t.Run("message type signature", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, v1.MaxNumOfMessageTypeSignatureAccepted, service.GetMaxNumOfMessageTypeAccepted(v1.MtSignature))
	})
	t.Run("other message types", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, v1.DefaultMaxNumOfMessageTypeAccepted, service.GetMaxNumOfMessageTypeAccepted(v1.MtUnknown))
		assert.Equal(t, v1.DefaultMaxNumOfMessageTypeAccepted, service.GetMaxNumOfMessageTypeAccepted(v1.MtBlockBody))
		assert.Equal(t, v1.DefaultMaxNumOfMessageTypeAccepted, service.GetMaxNumOfMessageTypeAccepted(v1.MtBlockHeader))
		assert.Equal(t, v1.DefaultMaxNumOfMessageTypeAccepted, service.GetMaxNumOfMessageTypeAccepted(v1.MtBlockBodyAndHeader))
		assert.Equal(t, v1.DefaultMaxNumOfMessageTypeAccepted, service.GetMaxNumOfMessageTypeAccepted(v1.MtBlockHeaderFinalInfo))
	})
}
