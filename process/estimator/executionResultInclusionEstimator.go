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
// `lastNotarised` is nil if genesis.
// Return value* `allowed` is the count of **leading** entries in `pending` deemed safe. The caller slices `pending[:allowed]` and embeds them.
func Decide(cfg Config,
	lastNotarised *ExecutionResultMeta,
	pending []ExecutionResultMeta,
	currentHdrTsNs uint64) (allowed int) {
	allowed = 0

	if len(pending) == 0 {
		return allowed
	}

	var tBase uint64
	if lastNotarised == nil {
		tBase = cfg.GenesisTimeMs * 1_000_000 // ms → ns
	} else {
		tBase = lastNotarised.HeaderTimeMs * 1_000_000 // ms → ns
	}

	var estimatedTime uint64 // accumulated execution time in ns (1 gas = 1ns)
	for i, executionResultMeta := range pending {
		estimatedTime += executionResultMeta.GasUsed

		// Apply safety margin: et * μ = et * cfg.SafetyMargin / 100
		estimatedTimeWithMargin := estimatedTime * cfg.SafetyMargin / 100
		tDone := tBase + estimatedTimeWithMargin

		if tDone > currentHdrTsNs {
			return i // cannot include this or anything after
		}

		if cfg.MaxResultsPerBlock != 0 && uint64(i+1) >= cfg.MaxResultsPerBlock {
			return i + 1 // reached cap
		}
	}

	return len(pending)
}

/*Return value* `allowed` is the count of **leading** entries in `pending` deemed safe. The caller
slices `pending[:allowed]` and embeds them.

---

## 5. Core Algorithm

```text
1.  let ET      ← 0                // accumulated execution time (ns)
2.  let t_base  ← (if lastNotarised == nil
                    then genesisTimestampMs * 10^6 // convert to ns
                    else lastNotarised.HeaderTimeMs * 10^6)
3.  for i, res ∈ pending (in nonce order):
4.      Δgas    ← res.GasUsed
5.      ET     += Δgas * T_gas     // 1 ns per gas
6.      t_done  ← t_base + ET * (1 + μ)
7.      if t_done > currentHdrTsMs * 10^6: // convert to ns
8.          return i               // cannot include res and beyond
9.      if cfg.MaxResultsPerBlock ≠ 0 and i+1 ≥ cfg.MaxResultsPerBlock:
10.         return i+1
11. return len(pending)
*/
