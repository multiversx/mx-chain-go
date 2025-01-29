package chaos

import (
	"github.com/multiversx/mx-chain-core-go/data"
	chaosAdapters "github.com/multiversx/mx-chain-go/chaosAdapters"
)

type chaosController struct {
	enabled bool
	config  chaosConfig

	numCallsProcessTransaction        int
	numCallsDoSignatureJob            int
	numCallsCompleteSignatureSubround int
	numCallsCheckSignaturesValidity   int
	numCallsV2DoBlockJob              int
}

func newChaosController(configFilePath string) *chaosController {
	config, err := loadChaosConfigFromFile(configFilePath)
	if err != nil {
		log.Warn("Could not load chaos config", "error", err)
		return &chaosController{enabled: false}
	}

	return &chaosController{
		enabled: true,
		config:  config,
	}
}

func (controller *chaosController) LearnNodes(eligibleNodes []chaosAdapters.Validator, waitingNodes []chaosAdapters.Validator) {
	log.Info("Seeding chaos", "len(eligibleNodes)", len(eligibleNodes), "len(waitingNodes)", len(waitingNodes))
}

// In_shardProcess_processTransaction_shouldReturnError returns an error when processing a transaction, from time to time.
func (controller *chaosController) In_shardProcess_processTransaction_shouldReturnError() bool {
	if !controller.enabled {
		return false
	}

	controller.numCallsProcessTransaction++
	if controller.numCallsProcessTransaction%controller.config.NumCallsDivisor_processTransaction_shouldReturnError == 0 {
		log.Info("Returning error when processing transaction")
		return true
	}

	return false
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey corrupts the signature, from time to time.
func (controller *chaosController) In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenSingleKey(header data.HeaderHandler, signature []byte) {
	if !controller.enabled {
		return
	}

	controller.numCallsDoSignatureJob++
	if controller.numCallsDoSignatureJob%controller.config.NumCallsDivisor_maybeCorruptSignature != 0 {
		return
	}

	log.Info("Corrupting signature", "round", header.GetRound(), "nonce", header.GetNonce())
	signature[0] += 1
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey corrupts the signature, from time to time.
func (controller *chaosController) In_subroundSignature_doSignatureJob_maybeCorruptSignature_whenMultiKey(header data.HeaderHandler, keyIndex int, signature []byte) {
	if !controller.enabled {
		return
	}

	controller.numCallsDoSignatureJob++
	if controller.numCallsDoSignatureJob%controller.config.NumCallsDivisor_maybeCorruptSignature != 0 {
		return
	}

	log.Info("Corrupting signature", "round", header.GetRound(), "nonce", header.GetNonce(), "keyIndex", keyIndex)
	signature[0] += 1
}

// In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures skips waiting for signatures, from time to time.
func (controller *chaosController) In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(header data.HeaderHandler) bool {
	if !controller.enabled {
		return false
	}

	controller.numCallsCompleteSignatureSubround++
	if controller.numCallsCompleteSignatureSubround%controller.config.NumCallsDivisor_shouldSkipWaitingForSignatures != 0 {
		return false
	}

	log.Info("Skipping waiting for signatures", "round", header.GetRound(), "nonce", header.GetNonce())
	return true
}

func (controller *chaosController) In_subroundEndRound_checkSignaturesValidity_shouldReturnError(header data.HeaderHandler) bool {
	if !controller.enabled {
		return false
	}

	controller.numCallsCheckSignaturesValidity++
	if controller.numCallsCheckSignaturesValidity%controller.config.NumCallsDivisor_shouldReturnErrorInCheckSignaturesValidity != 0 {
		return false
	}

	log.Info("Returning error in check signatures validity", "round", header.GetRound(), "nonce", header.GetNonce())
	return true
}

// In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature corrupts the signature, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_maybeCorruptLeaderSignature(header data.HeaderHandler, signature []byte) {
	if !controller.enabled {
		return
	}

	controller.numCallsV2DoBlockJob++
	if controller.numCallsV2DoBlockJob%controller.config.NumCallsDivisor_consensusV2_maybeCorruptLeaderSignature != 0 {
		return
	}

	log.Info("Corrupting leader signature", "round", header.GetRound(), "nonce", header.GetNonce())
	signature[0] += 1
}

// In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock skips sending a block, from time to time.
func (controller *chaosController) In_V2_subroundBlock_doBlockJob_shouldSkipSendingBlock(header data.HeaderHandler) bool {
	if !controller.enabled {
		return false
	}

	controller.numCallsV2DoBlockJob++
	if controller.numCallsV2DoBlockJob%controller.config.NumCallsDivisor_consensusV2_maybeCorruptLeaderSignature != 0 {
		return false
	}

	log.Info("Skip sending block", "round", header.GetRound(), "nonce", header.GetNonce())
	return true
}
