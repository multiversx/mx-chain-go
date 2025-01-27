package chaos

import (
	"os"

	"github.com/multiversx/mx-chain-core-go/data"
	chaosAdapters "github.com/multiversx/mx-chain-go/chaosAdapters"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("chaos")

var context chaosContext

func init() {
	context = chaosContext{}
}

type chaosContext struct {
	numCallsProcessTransaction        int
	numCallsDoSignatureJob            int
	numCallsCompleteSignatureSubround int
	numCallsCheckSignaturesValidity   int
}

func Seed(eligibleNodes []chaosAdapters.Validator, waitingNodes []chaosAdapters.Validator) {
	log.Info("Seeding chaos", "len(eligibleNodes)", len(eligibleNodes), "len(waitingNodes)", len(waitingNodes))
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature corrupts the signature, from time to time.
func In_subroundSignature_doSignatureJob_maybeCorruptSignature(header data.HeaderHandler, signatureShare []byte) {
	if !isChaosEnabled() {
		return
	}

	context.numCallsDoSignatureJob++
	if context.numCallsDoSignatureJob%numCallsDivisor_maybeCorruptSignature != 0 {
		return
	}

	log.Info("Corrupting signature", "round", header.GetRound(), "nonce", header.GetNonce())
	signatureShare[0] += 1
}

// In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures skips waiting for signatures, from time to time.
func In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(header data.HeaderHandler) bool {
	if !isChaosEnabled() {
		return false
	}

	context.numCallsCompleteSignatureSubround++
	if context.numCallsCompleteSignatureSubround%numCallsDivisor_shouldSkipWaitingForSignatures != 0 {
		return false
	}

	log.Info("Skipping waiting for signatures", "round", header.GetRound(), "nonce", header.GetNonce())
	return true
}

func In_subroundEndRound_checkSignaturesValidity_shouldReturnError(header data.HeaderHandler) bool {
	if !isChaosEnabled() {
		return false
	}

	context.numCallsCheckSignaturesValidity++
	if context.numCallsCheckSignaturesValidity%numCallsDivisor_shouldReturnErrorInCheckSignaturesValidity != 0 {
		return false
	}

	log.Info("Returning error in check signatures validity", "round", header.GetRound(), "nonce", header.GetNonce())
	return true
}

// In_shardProcess_processTransaction_shouldReturnError returns an error when processing a transaction, from time to time.
func In_shardProcess_processTransaction_shouldReturnError() bool {
	if !isChaosEnabled() {
		return false
	}

	context.numCallsProcessTransaction++
	if context.numCallsProcessTransaction%numCallsDivisor_processTransaction_shouldReturnError == 0 {
		log.Info("Returning error when processing transaction")
		return true
	}

	return false
}

func isChaosEnabled() bool {
	valueAsString := os.Getenv("CHAOS")
	return len(valueAsString) > 0
}
