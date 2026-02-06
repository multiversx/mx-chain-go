package estimator

import (
	"errors"
	"math/bits"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

// ErrInvalidMaxResultsPerBlock signals that invalid max results per block config has been provided
var ErrInvalidMaxResultsPerBlock = errors.New("invalid max results per block config")

var log = logger.GetOrCreate("process/executionResultInclusionEstimator")

// Time per gas unit on minimum‑spec hardware (1 gas = 1ns)
var tGas = uint64(1)

// RoundHandler provides the current round index and timestamp.
type RoundHandler interface {
	GetTimeStampForRound(round uint64) uint64
	IsInterfaceNil() bool
}

type blockSizeComputationHandler interface {
	AddNumExecRes(numExecRes int)
	DecNumExecRes(numExecRes int)
	IsMaxExecResSizeReached(numNewExecRes int) bool
	IsInterfaceNil() bool
}

// ExecutionResultInclusionEstimator (EIE) is a deterministic component shipped with the MultiversX *Supernova*
// node. It determines, at proposal‑time and at validation‑time, whether one or more pending execution results can be
// safely embedded in the block that is being produced / verified.
type ExecutionResultInclusionEstimator struct {
	cfg          config.ExecutionResultInclusionEstimatorConfig // immutable after construction
	tGas         uint64                                         // time per gas unit on minimum‑spec hardware - 1 ns per gas unit
	roundHandler RoundHandler
	// TODO add also max estimated block gas capacity - used gas must be lower than this
	blockSizeComputation blockSizeComputationHandler
}

// NewExecutionResultInclusionEstimator returns a new instance of EIE
func NewExecutionResultInclusionEstimator(
	cfg config.ExecutionResultInclusionEstimatorConfig,
	roundHandler RoundHandler,
	blockSizeComputation blockSizeComputationHandler,
) (*ExecutionResultInclusionEstimator, error) {
	err := checkConfig(cfg)
	if err != nil {
		return nil, err
	}
	if check.IfNil(roundHandler) {
		return nil, process.ErrNilRoundHandler
	}
	if check.IfNil(blockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
	}

	return &ExecutionResultInclusionEstimator{
		cfg:                  cfg,
		tGas:                 tGas,
		roundHandler:         roundHandler,
		blockSizeComputation: blockSizeComputation,
	}, nil
}

func checkConfig(cfg config.ExecutionResultInclusionEstimatorConfig) error {
	if cfg.MaxResultsPerBlock == 0 {
		return ErrInvalidMaxResultsPerBlock
	}

	return nil
}

// Decide returns the prefix of `pending` that may be inserted into the block currently being built / verified.
// Return value: `allowed` is the count of leading entries in `pending` deemed safe. The caller slices `pending[:allowed]` and embeds them.
func (erie *ExecutionResultInclusionEstimator) Decide(
	lastNotarised *common.LastExecutionResultForInclusion,
	pending []data.BaseExecutionResultHandler,
	currentRound uint64,
) (allowed int) {
	allowed = 0

	if len(pending) == 0 {
		return allowed
	}

	nominalSafetyMargin := uint64(100)
	safetyMargin := erie.cfg.SafetyMargin
	if erie.cfg.SafetyMargin > nominalSafetyMargin {
		safetyMargin -= nominalSafetyMargin
	}

	var roundForTBaseCalculation uint64
	var previousExecutionResultMeta data.BaseExecutionResultHandler

	// lastNotarised is nil if genesis.
	if lastNotarised == nil {
		roundForTBaseCalculation = 0
	} else {
		roundForTBaseCalculation = lastNotarised.NotarizedInRound
	}

	tBase := convertMsToNs(erie.roundHandler.GetTimeStampForRound(roundForTBaseCalculation))

	currentHdrTsMs := erie.roundHandler.GetTimeStampForRound(currentRound)
	currentHdrTsNs := convertMsToNs(currentHdrTsMs)

	// accumulated execution tBase for each pending execution result in ns (1 gas = 1ns)
	estimatedTBase := tBase
	for pendingExecutionIndex, executionResultMeta := range pending {
		if pendingExecutionIndex > 0 {
			previousExecutionResultMeta = pending[pendingExecutionIndex-1]
		}
		ok := erie.checkSanity(executionResultMeta, previousExecutionResultMeta, lastNotarised, currentRound)
		if !ok {
			return pendingExecutionIndex
		}

		overflow, currentEstimatedTime := bits.Mul64(executionResultMeta.GetGasUsed(), erie.tGas)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in currentEstimatedTime",
				"currentEstimatedTime", currentEstimatedTime,
				"gasUsed", executionResultMeta.GetGasUsed(),
				"tGas", erie.tGas,
			)
			return pendingExecutionIndex
		}

		// Round timestamp for current execution result, since it is the time when the execution result becomes available
		tRoundNs := convertMsToNs(erie.roundHandler.GetTimeStampForRound(executionResultMeta.GetHeaderRound()))

		// Align previously estimated tBase with start of round if needed
		estimatedTBase = max(estimatedTBase, tRoundNs)
		estimatedTBase, overflow = bits.Add64(estimatedTBase, currentEstimatedTime, 0)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in block transactions time estimation",
				"estimatedTBase", estimatedTBase,
				"currentEstimatedTime", currentEstimatedTime)
			return pendingExecutionIndex
		}

		// Apply safety margin to current estimated time
		overflow, currentEstimatedTimeMargin := bits.Mul64(currentEstimatedTime, safetyMargin)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in estimated time with margin",
				"currentEstimatedTime", currentEstimatedTime,
				"safetyMargin", safetyMargin)
			return pendingExecutionIndex
		}
		currentEstimatedTimeMargin /= nominalSafetyMargin

		tDone, overflow := bits.Add64(estimatedTBase, currentEstimatedTimeMargin, 0)
		if overflow != 0 {
			log.Debug("ExecutionResultInclusionEstimator: overflow detected in total estimated time",
				"estimatedTBase", estimatedTBase,
				"estimatedTimeWithMargin", currentEstimatedTimeMargin)
			return pendingExecutionIndex
		}

		// check for time cap reached, cannot include current pending item or anything after
		if tDone > currentHdrTsNs {
			log.Debug("ExecutionResultInclusionEstimator: estimated time exceeds current header timestamp",
				"tDone", tDone,
				"currentHdrTsNs", currentHdrTsNs)
			return pendingExecutionIndex
		}

		if erie.blockSizeComputation.IsMaxExecResSizeReached(1) {
			log.Debug("ExecutionResultInclusionEstimator: estimated max size reached",
				"currentIndex", pendingExecutionIndex)
			return pendingExecutionIndex
		}
		erie.blockSizeComputation.AddNumExecRes(1)

		// check for number of results cap reached, including current pending item. MaxResultsPerBlock = 0 means no cap.
		if uint64(pendingExecutionIndex+1) >= erie.cfg.MaxResultsPerBlock {
			log.Debug("ExecutionResultInclusionEstimator: reached MaxResultsPerBlock cap",
				"maxResultsPerBlock", erie.cfg.MaxResultsPerBlock,
				"currentIndex", pendingExecutionIndex+1)
			return pendingExecutionIndex + 1
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
	lastNotarised *common.LastExecutionResultForInclusion,
	currentRound uint64,
) bool {
	// Check for genesis round
	if currentExecutionResult.GetHeaderRound() == 0 {
		log.Debug("ExecutionResultInclusionEstimator: ExecutionResultHeaderRound on genesis detected",
			"headerNonce", currentExecutionResult.GetHeaderNonce(),
		)
		return false
	}
	// Check for strict nonce monotonicity
	if previousExecutionResult != nil && currentExecutionResult.GetHeaderNonce() != previousExecutionResult.GetHeaderNonce()+1 {
		log.Debug("ExecutionResultInclusionEstimator: non-monotonic HeaderNonce detected",
			"currentHeaderNonce", currentExecutionResult.GetHeaderNonce(),
			"previousHeaderNonce", previousExecutionResult.GetHeaderNonce(),
			"currentRound", currentExecutionResult.GetHeaderRound(),
			"previousRound", previousExecutionResult.GetHeaderRound(),
		)
		return false
	}
	// Check for monotonicity of rounds
	if previousExecutionResult != nil && currentExecutionResult.GetHeaderRound() <= previousExecutionResult.GetHeaderRound() {
		log.Debug("ExecutionResultInclusionEstimator: non-monotonic rounds detected",
			"currentRound", currentExecutionResult.GetHeaderRound(),
			"previousERound", previousExecutionResult.GetHeaderRound(),
		)
		return false
	}

	// Check for round before last notarised
	if lastNotarised != nil && currentExecutionResult.GetHeaderRound() <= lastNotarised.ProposedInRound {
		log.Debug("ExecutionResultInclusionEstimator: Round before last notarised detected",
			"headerNonce", currentExecutionResult.GetHeaderNonce(),
			"lastNotarisedResultProposalRound", lastNotarised.ProposedInRound,
			"currentExecutionResultRound", currentExecutionResult.GetHeaderRound())
		return false
	}
	// Check for results in the future
	if currentExecutionResult.GetHeaderRound() >= currentRound {
		log.Debug("ExecutionResultInclusionEstimator: Execution result round in the future detected",
			"currentExecutionResultNonce", currentExecutionResult.GetHeaderNonce(),
			"currentExecutionResultRound", currentExecutionResult.GetHeaderRound(),
			"currentHeaderRound", currentRound,
		)
		return false
	}
	return true
}

// TODO check for overflow?
func convertMsToNs(ms uint64) uint64 {
	// Convert milliseconds to nanoseconds
	return ms * 1_000_000
}
