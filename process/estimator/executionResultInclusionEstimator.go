package estimator

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
	var estimatedTime uint64
	for i, executionResultMeta := range pending {
		estimatedTime += executionResultMeta.GasUsed * t_gas

		// Apply safety margin
		estimatedTimeWithMargin := estimatedTime * cfg.SafetyMargin / 100
		tDone := tBase + estimatedTimeWithMargin

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
