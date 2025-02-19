package chaos

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type chaosController struct {
	mutex      sync.RWMutex
	enabled    bool
	config     *chaosConfig
	nodeConfig *config.Configs
	node       NodeHandler
}

func newChaosController() *chaosController {
	return &chaosController{enabled: false}
}

// Setup sets up the chaos controller. Make sure to call this only after logging components (file logging, as well) are set up.
func (controller *chaosController) Setup() {
	config, err := newChaosConfigFromFile(defaultConfigFilePath)
	if err != nil {
		log.Error("could not load chaos config", "error", err)
		return
	}

	controller.mutex.Lock()
	controller.config = config
	controller.enabled = true
	controller.mutex.Unlock()
}

// HandleNodeConfig -
func (controller *chaosController) HandleNodeConfig(config *config.Configs) {
	log.Info("HandleNodeConfig")

	controller.mutex.Lock()
	controller.nodeConfig = config
	controller.mutex.Unlock()
}

// HandleNode -
func (controller *chaosController) HandleNode(node NodeHandler) {
	log.Info("HandleNode")

	controller.mutex.Lock()
	controller.node = node
	controller.mutex.Unlock()

	node.GetCoreComponents().EpochNotifier().RegisterNotifyHandler(controller)
}

// EpochConfirmed -
func (controller *chaosController) EpochConfirmed(epoch uint32, timestamp uint64) {
	log.Info("EpochConfirmed", "epoch", epoch, "timestamp", timestamp)

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(nil, "")
	if controller.shouldFailNoLock(failurePanicOnEpochChange, circumstance) {
		panic("chaos: panic on epoch change")
	}
}

// In_shardBlock_CreateBlock_shouldReturnError -
func (controller *chaosController) In_shardBlock_CreateBlock_shouldReturnError() bool {
	log.Trace("In_shardBlock_CreateBlock_shouldReturnError")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(nil, "")
	return controller.shouldFailNoLock(failureCreatingBlockError, circumstance)
}

// In_shardBlock_ProcessBlock_shouldReturnError -
func (controller *chaosController) In_shardBlock_ProcessBlock_shouldReturnError() bool {
	log.Trace("In_shardBlock_ProcessBlock_shouldReturnError")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(nil, "")
	return controller.shouldFailNoLock(failureProcessingBlockError, circumstance)
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey -
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(consensusState spos.ConsensusStateHandler, signature []byte) {
	log.Trace("In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(consensusState, "")

	if controller.shouldFailNoLock(failureConsensusCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey -
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(consensusState spos.ConsensusStateHandler, nodePublicKey string, signature []byte) {
	log.Trace("In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(consensusState, nodePublicKey)

	if controller.shouldFailNoLock(failureConsensusCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError -
func (controller *chaosController) In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError(consensusState spos.ConsensusStateHandler) bool {
	log.Trace("In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(consensusState, "")
	return controller.shouldFailNoLock(failureConsensusV1ReturnErrorInCheckSignaturesValidity, circumstance)
}

// In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcastingFinalBlock -
func (controller *chaosController) In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcastingFinalBlock(consensusState spos.ConsensusStateHandler) {
	log.Trace("In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcast")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(consensusState, "")

	if controller.shouldFailNoLock(failureConsensusV1DelayBroadcastingFinalBlockAsLeader, circumstance) {
		duration := controller.config.getFailureParameterAsFloat64(failureConsensusV1DelayBroadcastingFinalBlockAsLeader, "duration")
		time.Sleep(time.Duration(duration))
	}
}

// In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature -
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(consensusState spos.ConsensusStateHandler, header data.HeaderHandler, signature []byte) {
	log.Trace("In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(consensusState, "")
	circumstance.blockNonce = header.GetNonce()

	if controller.shouldFailNoLock(failureConsensusV2CorruptLeaderSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature -
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature(consensusState spos.ConsensusStateHandler, header data.HeaderHandler) {
	log.Trace("In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(consensusState, "")
	circumstance.blockNonce = header.GetNonce()

	if controller.shouldFailNoLock(failureConsensusV2DelayLeaderSignature, circumstance) {
		duration := controller.config.getFailureParameterAsFloat64(failureConsensusV2DelayLeaderSignature, "duration")
		time.Sleep(time.Duration(duration))
	}
}

// In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock -
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(consensusState spos.ConsensusStateHandler, header data.HeaderHandler) bool {
	log.Trace("In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock")

	controller.mutex.RLock()
	defer controller.mutex.RUnlock()

	circumstance := controller.acquireCircumstanceNoLock(consensusState, "")
	circumstance.blockNonce = header.GetNonce()

	return controller.shouldFailNoLock(failureConsensusV2SkipSendingBlock, circumstance)
}

// Should only be called wi
func (controller *chaosController) acquireCircumstanceNoLock(consensusState spos.ConsensusStateHandler, nodePublicKey string) *failureCircumstance {
	circumstance := newFailureCircumstance()
	circumstance.nodeDisplayName = controller.nodeConfig.PreferencesConfig.Preferences.NodeDisplayName
	circumstance.enrichWithLoggerCorrelation(logger.GetCorrelation())
	circumstance.enrichWithConsensusState(consensusState, nodePublicKey)

	return circumstance
}

func (controller *chaosController) shouldFailNoLock(failureName failureName, circumstance *failureCircumstance) bool {
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
		log.Info("shouldFail", "failureName", failureName)
		return true
	}

	return false
}

// IsInterfaceNil -
func (controller *chaosController) IsInterfaceNil() bool {
	return controller == nil
}
