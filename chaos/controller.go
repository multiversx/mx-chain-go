package chaos

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
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
		log.Warn("Could not load chaos config", "error", err)
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
	return controller.shouldFail(failureProcessTransactionShouldReturnError, circumstance)
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey corrupts the signature, from time to time.
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(consensusState spos.ConsensusStateHandler, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = consensusState.GetHeader().GetNonce()
	if controller.shouldFail(failureShouldCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey corrupts the signature, from time to time.
func (controller *chaosController) In_V1_and_V2_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(consensusState spos.ConsensusStateHandler, nodePublicKey string, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, nodePublicKey)
	circumstance.blockNonce = consensusState.GetHeader().GetNonce()
	if controller.shouldFail(failureShouldCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V1_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures skips waiting for signatures, from time to time.
func (controller *chaosController) In_V1_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(consensusState spos.ConsensusStateHandler) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = consensusState.GetHeader().GetNonce()
	return controller.shouldFail(failureShouldSkipWaitingForSignatures, circumstance)
}

func (controller *chaosController) In_V1_subroundEndRound_checkSignaturesValidity_shouldReturnError(consensusState spos.ConsensusStateHandler) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = consensusState.GetHeader().GetNonce()
	return controller.shouldFail(failureShouldReturnErrorInCheckSignaturesValidity, circumstance)
}

// In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature corrupts the signature, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(consensusState spos.ConsensusStateHandler, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = consensusState.GetHeader().GetNonce()
	if controller.shouldFail(failureShouldCorruptLeaderSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock skips sending a block, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(consensusState spos.ConsensusStateHandler) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance(consensusState, "")
	circumstance.blockNonce = consensusState.GetHeader().GetNonce()
	return controller.shouldFail(failureShouldSkipSendingBlock, circumstance)
}

func (controller *chaosController) acquireCircumstance(consensusState spos.ConsensusStateHandler, nodePublicKey string) *failureCircumstance {
	randomNumber := rand.Uint64()
	now := time.Now().Unix()

	// For simplificty, we get the current shard, epoch and round from the logger correlation facility.
	loggerCorrelation := logger.GetCorrelation()

	shard, err := core.ConvertShardIDToUint32(loggerCorrelation.Shard)
	if err != nil {
		shard = math.MaxInt16
	}

	circumstance := &failureCircumstance{
		// Always available:
		nodeDisplayName: controller.nodeDisplayName,
		randomNumber:    randomNumber,
		now:             now,
		shard:           shard,
		epoch:           loggerCorrelation.Epoch,
		round:           uint64(loggerCorrelation.Round),

		// Not always available:
		nodeIndex:       -1,
		nodePublicKey:   nil,
		blockNonce:      0,
		transactionHash: nil,
	}

	if !check.IfNil(consensusState) {
		if nodePublicKey == "" {
			nodePublicKey = consensusState.SelfPubKey()
		}

		nodeIndex, err := consensusState.ConsensusGroupIndex(nodePublicKey)
		if err != nil {
			circumstance.nodeIndex = nodeIndex
		}

		circumstance.nodePublicKey = []byte(nodePublicKey)
		circumstance.blockNonce = consensusState.GetHeader().GetNonce()
	}

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
