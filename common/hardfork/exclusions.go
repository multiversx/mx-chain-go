package hardfork

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// ExcludedInterval defines a [Low, High] round range that must be ignored by consumers.
type ExcludedInterval struct {
	Low  uint64
	High uint64
}

// HfExcludedIntervals contains the round ranges in which incoming headers and proofs must be rejected.
// Consumers: hdrInterceptorProcessor.checkTestNetHardfork, interceptedEquivalentProof.CheckValidity and
// baseForkDetector (processReceivedBlock, processReceivedProof, append, computeProbableHighestNonce).
var HfExcludedIntervals = map[uint32][]ExcludedInterval{
	0:                     {{Low: 6783245, High: 6783245}},
	1:                     {{Low: 5536884, High: 5536885}},
	2:                     {{Low: 6783245, High: 6783245}},
	core.MetachainShardId: {{Low: 5609515, High: 6783200}},
}

// IntervalForRound returns the first interval defined for `shardID` that contains `round`, or nil.
func IntervalForRound(shardID uint32, round uint64) *ExcludedInterval {
	for i := range HfExcludedIntervals[shardID] {
		iv := &HfExcludedIntervals[shardID][i]
		if round >= iv.Low && round <= iv.High {
			return iv
		}
	}
	return nil
}

// IsRoundExcluded returns true if `round` falls inside any interval defined for `shardID`.
func IsRoundExcluded(shardID uint32, round uint64) bool {
	return IntervalForRound(shardID, round) != nil
}
