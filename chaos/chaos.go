package chaos

import (
	"os"

	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("chaos")

var context chaosContext

func init() {
	context = chaosContext{}
}

type chaosContext struct {
	processTransactionCounter int
}

// In_subroundSignature_doSignatureJob_maybeCorruptSignature corrupts the signature, from time to time.
func In_subroundSignature_doSignatureJob_maybeCorruptSignature(header data.HeaderHandler, signatureShare []byte) {
	if !isChaosEnabled() {
		return
	}

	round := header.GetRound()
	nonce := header.GetNonce()

	if round%uint64(roundDivisor_maybeCorruptSignature) != 0 {
		return
	}

	log.Info("Corrupting signature", "round", round, "nonce", nonce)
	signatureShare[0] += 1
}

// In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures skips waiting for signatures, from time to time.
func In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(header data.HeaderHandler) bool {
	if !isChaosEnabled() {
		return false
	}

	round := header.GetRound()
	nonce := header.GetNonce()

	if round%uint64(roundDivisor_shouldSkipWaitingForSignatures) != 0 {
		return false
	}

	log.Info("Skipping waiting for signatures", "round", round, "nonce", nonce)
	return true
}

func In_subroundEndRound_checkSignaturesValidity_shouldReturnError(header data.HeaderHandler) bool {
	if !isChaosEnabled() {
		return false
	}

	round := header.GetRound()
	nonce := header.GetNonce()

	if round%uint64(roundDivisor_shouldReturnErrorInCheckSignaturesValidity) != 0 {
		return false
	}

	log.Info("Returning error in check signatures validity", "round", round, "nonce", nonce)
	return true
}

// In_shardProcess_processTransaction_shouldReturnError returns an error when processing a transaction, from time to time.
func In_shardProcess_processTransaction_shouldReturnError() bool {
	if !isChaosEnabled() {
		return false
	}

	context.processTransactionCounter++
	if context.processTransactionCounter%numCallsDivisor_processTransaction_shouldReturnError == 0 {
		log.Info("Returning error when processing transaction")
		return true
	}

	return false
}

func isChaosEnabled() bool {
	valueAsString := os.Getenv("CHAOS")
	return len(valueAsString) > 0
}
