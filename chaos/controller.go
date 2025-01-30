package chaos

import (
	"math/rand"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/chaosAdapters"
)

type chaosController struct {
	mutex                  sync.Mutex
	enabled                bool
	config                 *chaosConfig
	currentShard           uint32
	currentEpoch           uint32
	currentRound           uint64
	currentlyEligibleNodes []chaosAdapters.Validator
	currentlyWaitingNodes  []chaosAdapters.Validator

	CallsCounters *callsCounters
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

func (controller *chaosController) LearnSelfShard(shard uint32) {
	log.Info("LearnSelfShard", "shard", shard)
	controller.currentShard = shard
}

func (controller *chaosController) LearnCurrentEpoch(epoch uint32) {
	log.Info("LearnCurrentEpoch", "epoch", epoch)
	controller.currentEpoch = epoch
}

func (controller *chaosController) LearnCurrentRound(round int64) {
	log.Info("LearnCurrentRound", "round", round)
	controller.currentRound = uint64(round)
}

func (controller *chaosController) LearnNodes(eligibleNodes []chaosAdapters.Validator, waitingNodes []chaosAdapters.Validator) {
	log.Info("LearnNodes", "len(eligibleNodes)", len(eligibleNodes), "len(waitingNodes)", len(waitingNodes))

	controller.currentlyEligibleNodes = eligibleNodes
	controller.currentlyWaitingNodes = waitingNodes
}

// In_shardProcess_processTransaction_shouldReturnError returns an error when processing a transaction, from time to time.
func (controller *chaosController) In_shardProcess_processTransaction_shouldReturnError() bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	return controller.shouldFail(failureProcessTransactionShouldReturnError, circumstance)
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey corrupts the signature, from time to time.
func (controller *chaosController) In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(header data.HeaderHandler, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.blockNonce = header.GetNonce()
	if controller.shouldFail(failureMaybeCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey corrupts the signature, from time to time.
func (controller *chaosController) In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(header data.HeaderHandler, keyIndex int, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.blockNonce = header.GetNonce()
	if controller.shouldFail(failureMaybeCorruptSignature, circumstance) {
		signature[0] += 1
	}
}

// In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures skips waiting for signatures, from time to time.
func (controller *chaosController) In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(header data.HeaderHandler) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.blockNonce = header.GetNonce()
	return controller.shouldFail(failureShouldSkipWaitingForSignatures, circumstance)
}

func (controller *chaosController) In_subroundEndRound_checkSignaturesValidity_shouldReturnError(header data.HeaderHandler) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.blockNonce = header.GetNonce()
	return controller.shouldFail(failureShouldReturnErrorInCheckSignaturesValidity, circumstance)
}

// In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature corrupts the signature, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(header data.HeaderHandler, signature []byte) {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.blockNonce = header.GetNonce()
	if controller.shouldFail(failureMaybeCorruptLeaderSignature, circumstance) {
		signature[0] += 1
	}
}

// In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock skips sending a block, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(header data.HeaderHandler) bool {
	controller.mutex.Lock()
	defer controller.mutex.Unlock()

	circumstance := controller.acquireCircumstance()
	circumstance.blockNonce = header.GetNonce()
	return controller.shouldFail(failureShouldSkipSendingBlock, circumstance)
}

func (controller *chaosController) acquireCircumstance() *failureCircumstance {
	randomNumber := rand.Uint64()
	now := time.Now().Unix()

	return &failureCircumstance{
		randomNumber: randomNumber,
		now:          now,
		shard:        controller.currentShard,
		epoch:        controller.currentEpoch,
		round:        controller.currentRound,

		counterProcessTransaction: controller.CallsCounters.ProcessTransaction.GetUint64(),
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

	shouldFail, err := circumstance.evalExpression(failure.When)
	if err != nil {
		log.Warn("Failed to evaluate expression", "error", err)
		return false
	}

	if shouldFail {
		log.Info("shouldFail()", "failureName", failureName)
		return true
	}

	return false
}
