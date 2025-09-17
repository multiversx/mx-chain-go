package estimator

import (
	"math/bits"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/config"
)

var log = logger.GetOrCreate("process/executionResultInclusionEstimator")

// Time per gas unit on minimum‑spec hardware (1 gas = 1ns)
var tGas = uint64(1)

// RoundHandler provides the current round index and timestamp.
type RoundHandler interface {
	GetTimeStampForRound(round uint64) uint64
	IsInterfaceNil() bool
}

// LastExecutionResultForInclusion is a lightweight summary of the last notarized execution result EIE requires.
type LastExecutionResultForInclusion struct {
	NotarizedInRound uint64
	ProposedInRound  uint64
}

// ExecutionResultInclusionEstimator (EIE) is a deterministic component shipped with the MultiversX *Supernova*
// node. It determines, at proposal‑time and at validation‑time, whether one or more pending execution results can be
// safely embedded in the block that is being produced / verified.
type ExecutionResultInclusionEstimator struct {
	cfg          config.ExecutionResultInclusionEstimatorConfig // immutable after construction
	tGas         uint64                                         // time per gas unit on minimum‑spec hardware - 1 ns per gas unit
	roundHandler RoundHandler
	// TODO add also max estimated block gas capacity - used gas must be lower than this
}

// NewExecutionResultInclusionEstimator returns a new instance of EIE
func NewExecutionResultInclusionEstimator(cfg config.ExecutionResultInclusionEstimatorConfig, roundHandler RoundHandler) *ExecutionResultInclusionEstimator {
	if check.IfNil(roundHandler) {
		log.Error("NewExecutionResultInclusionEstimator: nil roundHandler")
		return nil
	}
	return &ExecutionResultInclusionEstimator{
		cfg:          cfg,
		tGas:         tGas,
		roundHandler: roundHandler,
	}
}

// Decide returns the prefix of `pending` that may be inserted into the block currently being built / verified.
// Return value: `allowed` is the count of leading entries in `pending` deemed safe. The caller slices `pending[:allowed]` and embeds them.
func (erie *ExecutionResultInclusionEstimator) Decide(
	lastNotarised *LastExecutionResultForInclusion,
	pending []data.BaseExecutionResultHandler,
	currentHdrTsMs uint64,
) (allowed int) {
	allowed = 0

	if len(pending) == 0 {
		return allowed
	}

	var roundForTBaseCalculation uint64

	// lastNotarised is nil if genesis.
	if lastNotarised == nil {
		roundForTBaseCalculation = 0
	} else {
		roundForTBaseCalculation = lastNotarised.NotarizedInRound
	}
	var previousExecutionResultMeta data.BaseExecutionResultHandler

	tBase := convertMsToNs(erie.roundHandler.GetTimeStampForRound(roundForTBaseCalculation))

	currentHdrTsNs := convertMsToNs(currentHdrTsMs)

	// accumulated execution time in ns (1 gas = 1ns)
	estimatedTime := uint64(0)
	for i, executionResultMeta := range pending {
		if i > 0 {
			previousExecutionResultMeta = pending[i-1]
		}
		ok := erie.checkSanity(executionResultMeta, previousExecutionResultMeta, lastNotarised, currentHdrTsNs)
		if !ok {
			return i
		}

		overflow, currentEstimatedTime := bits.Mul64(executionResultMeta.GetGasUsed(), erie.tGas)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in currentEstimatedTime",
				"currentEstimatedTime", currentEstimatedTime,
				"gasUsed", executionResultMeta.GetGasUsed(),
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

// IsInterfaceNil returns true if there is no value under the interface
func (erie *ExecutionResultInclusionEstimator) IsInterfaceNil() bool {
	return erie == nil
}

func (erie *ExecutionResultInclusionEstimator) checkSanity(
	currentExecutionResult data.BaseExecutionResultHandler,
	previousExecutionResult data.BaseExecutionResultHandler,
	lastNotarised *LastExecutionResultForInclusion,
	currentHdrTsNs uint64,
) bool {
	currentExecutionResultForProposalTimestamp := erie.roundHandler.GetTimeStampForRound(currentExecutionResult.GetHeaderRound())
	currentExecutionResultForProposalTimestampNs := convertMsToNs(currentExecutionResultForProposalTimestamp)
	// Check for strict nonce monotonicity
	if previousExecutionResult != nil && currentExecutionResult.GetHeaderNonce() != previousExecutionResult.GetHeaderNonce()+1 {
		log.Debug("ExecutionResultInclusionEstimator: non-monotonic HeaderNonce detected",
			"currentHeaderNonce", currentExecutionResult.GetHeaderNonce(),
			"previousHeaderNonce", previousExecutionResult.GetHeaderNonce(),
			"currentHeaderTimeMs", currentExecutionResultForProposalTimestamp,
			"previousHeaderTimeMs", erie.roundHandler.GetTimeStampForRound(previousExecutionResult.GetHeaderRound()),
		)
		return false
	}
	// Check for monotonicity of rounds
	if previousExecutionResult != nil && currentExecutionResult.GetHeaderRound() < previousExecutionResult.GetHeaderRound() {
		log.Debug("ExecutionResultInclusionEstimator: non-monotonic rounds detected",
			"currentHeaderTimeMs", currentExecutionResultForProposalTimestamp,
			"previousHeaderTimeMs", erie.roundHandler.GetTimeStampForRound(previousExecutionResult.GetHeaderRound()),
		)
		return false
	}
	// Check for time before genesis time
	genesisTimeMs := erie.roundHandler.GetTimeStampForRound(0)

	if currentExecutionResultForProposalTimestamp < genesisTimeMs {
		log.Debug("ExecutionResultInclusionEstimator: HeaderTimeMs before genesis detected",
			"headerNonce", currentExecutionResult.GetHeaderNonce(),
			"headerTimeMs", currentExecutionResultForProposalTimestamp,
			"genesisTimeMs", genesisTimeMs,
		)
		return false
	}
	// Check for round before last notarised
	if lastNotarised != nil && currentExecutionResult.GetHeaderRound() < lastNotarised.NotarizedInRound {
		log.Debug("ExecutionResultInclusionEstimator: HeaderTimeMs before last notarised detected",
			"headerNonce", currentExecutionResult.GetHeaderNonce(),
			"headerTimeMs", currentExecutionResultForProposalTimestamp,
			"lastNotarisedTimeMs", erie.roundHandler.GetTimeStampForRound(lastNotarised.NotarizedInRound),
		)
		return false
	}

	// Check for results in the future
	if currentExecutionResultForProposalTimestampNs > currentHdrTsNs {
		log.Debug("ExecutionResultInclusionEstimator: HeaderTimeMs in the future detected",
			"headerNonce", currentExecutionResult.GetHeaderNonce(),
			"headerTimeNs", currentExecutionResultForProposalTimestampNs,
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
