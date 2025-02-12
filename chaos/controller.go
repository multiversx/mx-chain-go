package chaos

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/data"
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

	circumstance := controller.acquireCircumstance()
	return controller.shouldFail(failureProcessTransactionShouldReturnError, circumstance)
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey corrupts the signature, from time to time.
func (controller *chaosController) In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(header data.HeaderHandler, nodeIndex int, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.nodeIndex = nodeIndex
	circumstance.blockNonce = header.GetNonce()
	if controller.shouldFail(failureShouldCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey corrupts the signature, from time to time.
func (controller *chaosController) In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(header data.HeaderHandler, nodeIndex int, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.nodeIndex = nodeIndex
	circumstance.blockNonce = header.GetNonce()
	if controller.shouldFail(failureShouldCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures skips waiting for signatures, from time to time.
func (controller *chaosController) In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(header data.HeaderHandler, nodeIndex int) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.nodeIndex = nodeIndex
	circumstance.blockNonce = header.GetNonce()
	return controller.shouldFail(failureShouldSkipWaitingForSignatures, circumstance)
}

func (controller *chaosController) In_subroundEndRound_checkSignaturesValidity_shouldReturnError(header data.HeaderHandler, nodeIndex int) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.nodeIndex = nodeIndex
	circumstance.blockNonce = header.GetNonce()
	return controller.shouldFail(failureShouldReturnErrorInCheckSignaturesValidity, circumstance)
}

// In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature corrupts the signature, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(header data.HeaderHandler, nodeIndex int, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.nodeIndex = nodeIndex
	circumstance.blockNonce = header.GetNonce()
	if controller.shouldFail(failureShouldCorruptLeaderSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock skips sending a block, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(header data.HeaderHandler, nodeIndex int) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.nodeIndex = nodeIndex
	circumstance.blockNonce = header.GetNonce()
	return controller.shouldFail(failureShouldSkipSendingBlock, circumstance)
}

func (controller *chaosController) acquireCircumstance() *failureCircumstance {
	randomNumber := rand.Uint64()
	now := time.Now().Unix()

	// For simplificty, we get the current shard, epoch and round from the logger correlation facility.
	loggerCorrelation := logger.GetCorrelation()

	shard, err := core.ConvertShardIDToUint32(loggerCorrelation.Shard)
	if err != nil {
		shard = math.MaxInt16
	}

	return &failureCircumstance{
		// Always available:
		nodeDisplayName: controller.nodeDisplayName,
		randomNumber:    randomNumber,
		now:             now,
		shard:           shard,
		epoch:           loggerCorrelation.Epoch,
		round:           uint64(loggerCorrelation.Round),

		// Always available (counters):
		counterProcessTransaction: controller.CallsCounters.ProcessTransaction.GetUint64(),

		// WHEN WE KNOW INDEX, WE SHOULD NOT PUBLIC KEY, RIGHT???
		// from learned keys? nu pot se iau ambele colectii acolo?

		// Not always available:
		nodeIndex:       -1,
		blockNonce:      0,
		nodePublicKey:   nil,
		transactionHash: nil,
	}
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
