package hardfork

import (
	"sync"
)

// ExcludedInterval defines a [Low, High] round range that must be ignored by consumers.
type ExcludedInterval struct {
	Low  uint64
	High uint64
}

// shardExclusions holds all intervals declared for a single shard in the TOML config.
type shardExclusions struct {
	ShardID   uint32
	Intervals []ExcludedInterval
}

// hardforkExclusionsSection mirrors the top-level TOML section.
type hardforkExclusionsSection struct {
	Exclusions []shardExclusions
}

// HardforkExclusionsConfig is the TOML-parseable form of the exclusions file.
type HardforkExclusionsConfig struct {
	HardforkExclusions hardforkExclusionsSection
}

var (
	mutHfExcludedIntervals sync.RWMutex
	// HfExcludedIntervals contains the round ranges in which incoming headers and proofs must be rejected.
	// Populated at node startup from the TOML config file (see cmd/node/config/hardforkExclusions.toml).
	// Consumers: hdrInterceptorProcessor.checkTestNetHardfork, interceptedEquivalentProof.CheckValidity and
	// baseForkDetector (processReceivedBlock, processReceivedProof, append, computeProbableHighestNonce).
	HfExcludedIntervals = map[uint32][]ExcludedInterval{}
)

// SetHfExcludedIntervals replaces the in-memory exclusion map. Intended to be called once at startup
// from the config loader. Safe for concurrent use with IntervalForRound/IsRoundExcluded readers.
func SetHfExcludedIntervals(intervals map[uint32][]ExcludedInterval) {
	mutHfExcludedIntervals.Lock()
	defer mutHfExcludedIntervals.Unlock()
	HfExcludedIntervals = intervals
}

// ApplyConfig converts a parsed HardforkExclusionsConfig into the shard->intervals map and installs it.
func ApplyConfig(cfg *HardforkExclusionsConfig) {
	if cfg == nil {
		SetHfExcludedIntervals(map[uint32][]ExcludedInterval{})
		return
	}

	intervals := make(map[uint32][]ExcludedInterval, len(cfg.HardforkExclusions.Exclusions))
	for _, se := range cfg.HardforkExclusions.Exclusions {
		if len(se.Intervals) == 0 {
			continue
		}
		intervals[se.ShardID] = append(intervals[se.ShardID], se.Intervals...)
	}
	SetHfExcludedIntervals(intervals)
}

// IntervalForRound returns the first interval defined for `shardID` that contains `round`, or nil.
func IntervalForRound(shardID uint32, round uint64) *ExcludedInterval {
	mutHfExcludedIntervals.RLock()
	defer mutHfExcludedIntervals.RUnlock()

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
