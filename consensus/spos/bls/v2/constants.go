package v2

import (
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("consensus/spos/bls/v2")

// waitingAllSigsMaxTimeThreshold specifies the max allocated time for waiting all signatures from the total time of the subround signature
const waitingAllSigsMaxTimeThreshold = 0.5

// competingBlockSignDelay is the fraction of the full round time to wait before signing
// a competing block for the same nonce, giving the previous block's proof time to arrive.
const competingBlockSignDelay = 0.5

// competingProofSendDelay is the fraction of the full round time to wait before sending
// a proof, giving the previous block's proof time to arrive.
const competingProofSendDelay = 0.25

// acceptedClockSkew is the clock-skew tolerance applied on both ends of the invalid signers timestamp window
const acceptedClockSkew = time.Second

// significantEvidenceFraction is the fraction of the signature threshold above which observed
// shares for the previous round's block trigger the wait tier before signing a competing block
const significantEvidenceFraction = 0.5
