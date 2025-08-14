package estimator

import (
	"math/bits"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/executionResultInclusionEstimator")

// ExecutionResultMeta is a lightweight summary EIE requires.
type ExecutionResultMeta struct {
	HeaderHash   [32]byte // Link to full header in DB / cache
	HeaderNonce  uint64   // Monotonic within shard
	HeaderTimeMs uint64   // Milliseconds since Unix epoch, from header
	GasUsed      uint64   // Units actually consumed (post‑execution)
}

// Config supplied at construction, read‑only thereafter.
type Config struct {
	SafetyMargin       uint64 // default 110
	MaxResultsPerBlock uint64 // 0 = unlimited
	GenesisTimeMs      uint64 // required if lastNotarised == nil
}

// Decide returns the prefix of `pending` that may be inserted into the block currently being built / verified.
// Return value: `allowed` is the count of leading entries in `pending` deemed safe. The caller slices `pending[:allowed]`and embeds them.
func Decide(cfg Config,
	lastNotarised *ExecutionResultMeta,
	pending []ExecutionResultMeta,
	currentHdrTsNs uint64) (allowed int) {
	allowed = 0

	// time per gas unit on **minimum‑spec** hardware - 1 ns per gas unit
	t_gas := uint64(1)

	if len(pending) == 0 {
		return allowed
	}

	var tBase uint64
	// lastNotarised is nil if genesis.
	if lastNotarised == nil {
		tBase = convertMsToNs(cfg.GenesisTimeMs)
	} else {
		tBase = convertMsToNs(lastNotarised.HeaderTimeMs)
	}

	// accumulated execution time in ns (1 gas = 1ns)
	estimatedTime := uint64(0)
	for i, executionResultMeta := range pending {
		currentEstimatedTime := executionResultMeta.GasUsed * t_gas

		estimatedTime, overflow := bits.Add64(estimatedTime, currentEstimatedTime, 0)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in block tranzactions time estimation",
				"estimatedTime", estimatedTime,
				"currentEstimatedTime", currentEstimatedTime)
			return i
		}

		// Apply safety margin
		overflow, estimatedTimeWithMargin := bits.Mul64(estimatedTime, cfg.SafetyMargin/100)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in estimated time with margin",
				"estimatedTime", estimatedTime,
				"safetyMargin", cfg.SafetyMargin)
			return i
		}

		tDone, overflow := bits.Add64(tBase, estimatedTimeWithMargin, 0)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in total estimated time",
				"tBase", tBase,
				"estimatedTimeWithMargin", estimatedTimeWithMargin)
			return i
		}

		// cannot include current pending item or anything after
		if tDone > currentHdrTsNs {
			return i
		}

		// reached cap, including current pending item
		if cfg.MaxResultsPerBlock != 0 && uint64(i+1) >= cfg.MaxResultsPerBlock {
			return i + 1
		}
	}

	// If we reach here, all pending items are safe to include
	return len(pending)
}

func convertMsToNs(ms uint64) uint64 {
	// Convert milliseconds to nanoseconds
	return ms * 1_000_000
}
