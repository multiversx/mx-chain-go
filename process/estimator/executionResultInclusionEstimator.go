package estimator

import (
	"math/bits"

	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/executionResultInclusionEstimator")

// Time per gas unit on minimum‑spec hardware (1 gas = 1ns)
var tGas = uint64(1)

// ExecutionResultMetaData is a lightweight summary EIE requires.
type ExecutionResultMetaData struct {
	HeaderHash   [32]byte // Link to full header in DB / cache
	HeaderNonce  uint64   // Monotonic within shard
	HeaderTimeMs uint64   // Milliseconds since Unix epoch, from header
	GasUsed      uint64   // Units actually consumed (post‑execution)
}

// ExecutionResultInclusionEstimator (EIE) is a deterministic component shipped with the MultiversX *Supernova*
// node. It determines, at proposal‑time and at validation‑time, whether one or more pending execution results can be
// safely embedded in the block that is being produced / verified.
type ExecutionResultInclusionEstimator struct {
	cfg           config.ExecutionResultInclusionEstimatorConfig // immutable after construction
	tGas          uint64                                         // time per gas unit on minimum‑spec hardware - 1 ns per gas unit
	GenesisTimeMs uint64                                         // required if lastNotarised == nil
	// TODO add also max estimated block gas capacity - used gas must be lower than this
}

// NewExecutionResultInclusionEstimator returns a new instance of EIE
func NewExecutionResultInclusionEstimator(cfg config.ExecutionResultInclusionEstimatorConfig, GenesisTimeMs uint64) *ExecutionResultInclusionEstimator {
	return &ExecutionResultInclusionEstimator{
		cfg:           cfg,
		tGas:          tGas,
		GenesisTimeMs: GenesisTimeMs,
	}
}

// Decide returns the prefix of `pending` that may be inserted into the block currently being built / verified.
// Return value: `allowed` is the count of leading entries in `pending` deemed safe. The caller slices `pending[:allowed]` and embeds them.
func (erie *ExecutionResultInclusionEstimator) Decide(lastNotarised *ExecutionResultMetaData,
	pending []ExecutionResultMetaData,
	currentHdrTsMs uint64,
) (allowed int) {
	allowed = 0

	if len(pending) == 0 {
		return allowed
	}

	var tBase uint64
	// lastNotarised is nil if genesis.
	if lastNotarised == nil {
		tBase = convertMsToNs(erie.GenesisTimeMs)
	} else {
		tBase = convertMsToNs(lastNotarised.HeaderTimeMs)
	}

	currentHdrTsNs := convertMsToNs(currentHdrTsMs)

	// accumulated execution time in ns (1 gas = 1ns)
	estimatedTime := uint64(0)
	for i, executionResultMeta := range pending {
		var previousExecutionResultMeta *ExecutionResultMetaData
		if i > 0 {
			previousExecutionResultMeta = &pending[i-1]
		}
		ok := erie.checkSanity(executionResultMeta, previousExecutionResultMeta, lastNotarised, currentHdrTsNs)
		if !ok {
			return i
		}

		overflow, currentEstimatedTime := bits.Mul64(executionResultMeta.GasUsed, erie.tGas)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in currentEstimatedTime",
				"currentEstimatedTime", currentEstimatedTime,
				"gasUsed", executionResultMeta.GasUsed,
				"tGas", erie.tGas,
			)
			return i
		}

		estimatedTime, overflow = bits.Add64(estimatedTime, currentEstimatedTime, 0)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in block transactions time estimation",
				"estimatedTime", estimatedTime,
				"currentEstimatedTime", currentEstimatedTime)
			return i
		}

		// Apply safety margin
		overflow, estimatedTimeWithMargin := bits.Mul64(estimatedTime, erie.cfg.SafetyMargin)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in estimated time with margin",
				"estimatedTime", estimatedTime,
				"safetyMargin", erie.cfg.SafetyMargin)
			return i
		}
		estimatedTimeWithMargin /= 100

		tDone, overflow := bits.Add64(tBase, estimatedTimeWithMargin, 0)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in total estimated time",
				"tBase", tBase,
				"estimatedTimeWithMargin", estimatedTimeWithMargin)
			return i
		}

		// check for time cap reached, cannot include current pending item or anything after
		if tDone > currentHdrTsNs {
			log.Debug("ExecutionResultInclusionEstimator: estimated time exceeds current header timestamp",
				"tDone", tDone,
				"currentHdrTsNs", currentHdrTsNs)
			return i
		}

		// check for number of results cap reached, including current pending item. MaxResultsPerBlock = 0 means no cap.
		if erie.cfg.MaxResultsPerBlock != 0 && uint64(i+1) >= erie.cfg.MaxResultsPerBlock {
			log.Debug("ExecutionResultInclusionEstimator: reached MaxResultsPerBlock cap",
				"maxResultsPerBlock", erie.cfg.MaxResultsPerBlock,
				"currentIndex", i+1)
			return i + 1
		}
	}

	// If we reach here, all pending items are safe to include
	return len(pending)
}

func (erie ExecutionResultInclusionEstimator) checkSanity(currentExecutionResultMeta ExecutionResultMetaData,
	previousExecutionResultMeta *ExecutionResultMetaData,
	lastNotarised *ExecutionResultMetaData,
	currentHdrTsNs uint64,
) bool {
	// Check for strict nonce monotonicity
	if previousExecutionResultMeta != nil && currentExecutionResultMeta.HeaderNonce != previousExecutionResultMeta.HeaderNonce+1 {
		log.Debug("ExecutionResultInclusionEstimator: non-monotonic HeaderNonce detected",
			"currentHeaderNonce", currentExecutionResultMeta.HeaderNonce,
			"previousHeaderNonce", previousExecutionResultMeta.HeaderNonce,
			"currentHeaderTimeMs", currentExecutionResultMeta.HeaderTimeMs,
			"previousHeaderTimeMs", previousExecutionResultMeta.HeaderTimeMs,
		)
		return false
	}
	// Check for monotonicity in time
	if previousExecutionResultMeta != nil && currentExecutionResultMeta.HeaderTimeMs < previousExecutionResultMeta.HeaderTimeMs {
		log.Debug("ExecutionResultInclusionEstimator: non-monotonic HeaderTimeMs detected",
			"currentHeaderTimeMs", currentExecutionResultMeta.HeaderTimeMs,
			"previousHeaderTimeMs", previousExecutionResultMeta.HeaderTimeMs,
		)
		return false
	}
	// Check for time before genesis time
	if currentExecutionResultMeta.HeaderTimeMs < erie.GenesisTimeMs {
		log.Debug("ExecutionResultInclusionEstimator: HeaderTimeMs before genesis detected",
			"headerNonce", currentExecutionResultMeta.HeaderNonce,
			"headerTimeMs", currentExecutionResultMeta.HeaderTimeMs,
			"genesisTimeMs", erie.GenesisTimeMs,
		)
		return false
	}
	// Check for time before last notarised
	if lastNotarised != nil && currentExecutionResultMeta.HeaderTimeMs < lastNotarised.HeaderTimeMs {
		log.Debug("ExecutionResultInclusionEstimator: HeaderTimeMs before last notarised detected",
			"headerNonce", currentExecutionResultMeta.HeaderNonce,
			"headerTimeMs", currentExecutionResultMeta.HeaderTimeMs,
			"lastNotarisedTimeMs", lastNotarised.HeaderTimeMs,
		)
		return false
	}

	// Check for results in the future
	if convertMsToNs(currentExecutionResultMeta.HeaderTimeMs) > currentHdrTsNs {
		log.Debug("ExecutionResultInclusionEstimator: HeaderTimeMs in the future detected",
			"headerNonce", currentExecutionResultMeta.HeaderNonce,
			"headerTimeMs", currentExecutionResultMeta.HeaderTimeMs,
			"currentHdrTsNs", currentHdrTsNs,
		)
		return false
	}
	return true
}

// TODO check for overflow
func convertMsToNs(ms uint64) uint64 {
	// Convert milliseconds to nanoseconds
	return ms * 1_000_000
}
