package chaos

import (
	"os"

	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("chaos")

// From time to time, as leader, consider good signatures as being bad.
// From time to time, as leader, incorrectly process some transactions.
// From time to time, as anyone, forget to sign a block.
// From time to time, as anyone, sign badly.
// From time to time, panic.

func In_subroundSignature_doSignatureJob_maybeCorruptSignature(header data.HeaderHandler, signatureShare []byte) {
	if !isChaosEnabled() {
		return
	}

	nonce := header.GetNonce()
	round := header.GetRound()

	if nonce%5 != 0 {
		return
	}

	log.Info("Corrupting signature", "round", round, "nonce", nonce)
	signatureShare[0] += 1
}

func In_subroundSignature_completeSignatureSubRound_shouldSkipWaitingForSignatures(header data.HeaderHandler) bool {
	if !isChaosEnabled() {
		return false
	}

	nonce := header.GetNonce()
	round := header.GetRound()

	if nonce%7 != 0 {
		return false
	}

	log.Info("Skipping waiting for signatures", "round", round, "nonce", nonce)
	return true
}

func isChaosEnabled() bool {
	valueAsString := os.Getenv("CHAOS")
	return len(valueAsString) > 0
}
