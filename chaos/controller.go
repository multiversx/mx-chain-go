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
	mutex   sync.Mutex
	enabled bool
	config  *chaosConfig

	nodeConfig *config.Configs
	node       NodeHandler
}

func newChaosController(configFilePath string) *chaosController {
	log.Info("newChaosController", "configFilePath", configFilePath)

	config, err := newChaosConfigFromFile(configFilePath)
	if err != nil {
		log.Error("could not load chaos config", "error", err)
		return &chaosController{enabled: false}
	}

	return &chaosController{
		mutex:   sync.Mutex{},
		enabled: true,
		config:  config,
	}
}

// HandleNodeConfig -
func (controller *chaosController) HandleNodeConfig(config *config.Configs) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	log.Info("HandleNodeConfig")

	controller.nodeConfig = config
}

// HandleNode -
func (controller *chaosController) HandleNode(node NodeHandler) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	log.Info("HandleNode")

	controller.node = node
	controller.node.GetCoreComponents().EpochNotifier().RegisterNotifyHandler(controller)
}

func (controller *chaosController) EpochConfirmed(epoch uint32, timestamp uint64) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	log.Info("EpochConfirmed", "epoch", epoch, "timestamp", timestamp)

	circumstance := controller.acquireCircumstance(nil, "")

	if controller.shouldFail(failurePanicOnEpochChange, circumstance) {
		// TBD
	}
}

// In_shardBlock_CreateBlock_shouldReturnError -
func (controller *chaosController) In_shardBlock_CreateBlock_shouldReturnError() bool {
	log.Trace("In_shardBlock_CreateBlock_shouldReturnError")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(nil, "")
	return controller.shouldFail(failureCreatingBlockError, circumstance)
}

// In_shardBlock_ProcessBlock_shouldReturnError -
func (controller *chaosController) In_shardBlock_ProcessBlock_shouldReturnError() bool {
	log.Trace("In_shardBlock_ProcessBlock_shouldReturnError")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(nil, "")
	return controller.shouldFail(failureProcessingBlockError, circumstance)
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey -
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(consensusState spos.ConsensusStateHandler, signature []byte) {
	log.Trace("In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")

	if controller.shouldFail(failureConsensusCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey -
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(consensusState spos.ConsensusStateHandler, nodePublicKey string, signature []byte) {
	log.Trace("In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, nodePublicKey)

	if controller.shouldFail(failureConsensusCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError -
func (controller *chaosController) In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError(consensusState spos.ConsensusStateHandler) bool {
	log.Trace("In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	return controller.shouldFail(failureConsensusV1ReturnErrorInCheckSignaturesValidity, circumstance)
}

// In_V1_subroundEndRound_doEndRoundJobByLeader_maybeDelayBroadcastingFinalBlock -
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

// In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature -
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(consensusState spos.ConsensusStateHandler, header data.HeaderHandler, signature []byte) {
	log.Trace("In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = header.GetNonce()

	if controller.shouldFail(failureConsensusV2CorruptLeaderSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature -
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature(consensusState spos.ConsensusStateHandler, header data.HeaderHandler) {
	log.Trace("In_V2_subroundBlock_doBlockJob_maybeDelayLeaderSignature")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = header.GetNonce()

	if controller.shouldFail(failureConsensusV2DelayLeaderSignature, circumstance) {
		duration := controller.config.getFailureParameterAsFloat64(failureConsensusV2DelayLeaderSignature, "duration")
		time.Sleep(time.Duration(duration))
	}
}

// In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock -
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(consensusState spos.ConsensusStateHandler, header data.HeaderHandler) bool {
	log.Trace("In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock")

	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = header.GetNonce()

	return controller.shouldFail(failureConsensusV2SkipSendingBlock, circumstance)
}

func (controller *chaosController) acquireCircumstance(consensusState spos.ConsensusStateHandler, nodePublicKey string) *failureCircumstance {
	circumstance := newFailureCircumstance()
	circumstance.nodeDisplayName = controller.nodeConfig.PreferencesConfig.Preferences.NodeDisplayName
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
		log.Info("shouldFail", "failureName", failureName)
		return true
	}

	return false
}

// IsInterfaceNil -
func (controller *chaosController) IsInterfaceNil() bool {
	return controller == nil
}
