package chaos

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type chaosController struct {
	mutex           sync.Mutex
	enabled         bool
	config          *chaosConfig
	nodeDisplayName string
	CallsCounters   *callsCounters
}

type callsCounters struct {
	ProcessTransaction atomic.Counter
}

func newChaosController(configFilePath string) *chaosController {
	config, err := newChaosConfigFromFile(configFilePath)
	if err != nil {
		log.Error("Could not load chaos config", "error", err)
		return &chaosController{enabled: false}
	}

	return &chaosController{
		mutex:         sync.Mutex{},
		enabled:       true,
		config:        config,
		CallsCounters: &callsCounters{},
	}
}

// LearnNodeDisplayName learns the display name of the current node.
func (controller *chaosController) LearnNodeDisplayName(displayName string) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	log.Info("LearnNodeDisplayName", "displayName", displayName)
	controller.nodeDisplayName = displayName
}

// In_shardProcess_processTransaction_shouldReturnError returns an error when processing a transaction, from time to time.
func (controller *chaosController) In_shardProcess_processTransaction_shouldReturnError() bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(nil, "")
	return controller.shouldFail(failureProcessingTransactionError, circumstance)
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey corrupts the signature, from time to time.
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(consensusState spos.ConsensusStateHandler, signature []byte) {
	log.Trace("In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")

	if controller.shouldFail(failureConsensusCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey corrupts the signature, from time to time.
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(consensusState spos.ConsensusStateHandler, nodePublicKey string, signature []byte) {
	log.Trace("In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, nodePublicKey)

	if controller.shouldFail(failureConsensusCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures skips waiting for signatures, from time to time.
func (controller *chaosController) In_V1_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(consensusState spos.ConsensusStateHandler) bool {
	log.Trace("In_V1_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	return controller.shouldFail(failureConsensusV1SkipWaitingForSignatures, circumstance)
}

func (controller *chaosController) In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError(consensusState spos.ConsensusStateHandler) bool {
	log.Trace("In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	return controller.shouldFail(failureConsensusV1ReturnErrorInCheckSignaturesValidity, circumstance)
}

// In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcastingFinalBlock delays the broadcast of the block, from time to time.
func (controller *chaosController) In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcastingFinalBlock(consensusState spos.ConsensusStateHandler) {
	log.Trace("In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcast")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")

	if controller.shouldFail(failureConsensusV1DelayBroadcastingFinalBlockAsLeader, circumstance) {
		duration := controller.config.getFailureParameterAsFloat64(failureConsensusV1DelayBroadcastingFinalBlockAsLeader, "duration")
		time.Sleep(time.Duration(duration))
	}
}

// In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature corrupts the signature, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(consensusState spos.ConsensusStateHandler, signature []byte) {
	log.Trace("In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")

	if controller.shouldFail(failureConsensusV2CorruptLeaderSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature delays the leader signature, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature(consensusState spos.ConsensusStateHandler) {
	log.Trace("In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")

	if controller.shouldFail(failureConsensusV2DelayLeaderSignature, circumstance) {
		duration := controller.config.getFailureParameterAsFloat64(failureConsensusV2DelayLeaderSignature, "duration")
		time.Sleep(time.Duration(duration))
	}
}

// In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock skips sending a block, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(consensusState spos.ConsensusStateHandler) bool {
	log.Trace("In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	return controller.shouldFail(failureConsensusV2SkipSendingBlock, circumstance)
}

func (controller *chaosController) acquireCircumstance(consensusState spos.ConsensusStateHandler, nodePublicKey string) *failureCircumstance {
	circumstance := newFailureCircumstance()
	circumstance.nodeDisplayName = controller.nodeDisplayName
	circumstance.enrichWithLoggerCorrelation(logger.GetCorrelation())
	circumstance.enrichWithConsensusState(consensusState, nodePublicKey)

	return circumstance
}

func (controller *chaosController) shouldFail(failureName failureName, circumstance *failureCircumstance) bool {
	if !controller.enabled {
		return false
	}

	failure, configured := controller.config.getFailureByName(failureName)
	if !configured {
		return false
	}
	if !failure.Enabled {
		return false
	}

	shouldFail := circumstance.anyExpression(failure.Triggers)
	if shouldFail {
		log.Info("shouldFail()", "failureName", failureName)
		return true
	}

	return false
}
