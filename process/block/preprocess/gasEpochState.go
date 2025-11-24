package preprocess

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

type gasEpochState struct {
	economicsFee         process.FeeHandler
	enableEpochsHandler  common.EnableEpochsHandler
	enableRoundsHandler  common.EnableRoundsHandler
	mut                  sync.RWMutex
	overEstimationFactor uint64
	epochForLimits       uint32
	roundForLimits       uint64
}

func newGasEpochState(
	economicsFee process.FeeHandler,
	enableEpochsHandler common.EnableEpochsHandler,
	enableRoundsHandler common.EnableRoundsHandler,
) (*gasEpochState, error) {
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(enableRoundsHandler) {
		return nil, process.ErrNilEnableRoundsHandler
	}
	return &gasEpochState{
		economicsFee:         economicsFee,
		enableEpochsHandler:  enableEpochsHandler,
		enableRoundsHandler:  enableRoundsHandler,
		overEstimationFactor: noOverestimationFactor,
	}, nil
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (ges *gasEpochState) EpochConfirmed(epoch uint32) {
	ges.mut.Lock()
	defer ges.mut.Unlock()

	ges.epochForLimits = epoch

	// if already computed, only store the new epoch
	if ges.overEstimationFactor != noOverestimationFactor {
		return
	}

	isEpochFlagEnabled := ges.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, ges.epochForLimits)
	if !isEpochFlagEnabled {
		return
	}

	isRoundFlagEnabled := ges.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, ges.roundForLimits)
	if !isRoundFlagEnabled && epoch > 0 {
		ges.epochForLimits = epoch - 1 // use the previous epoch until activation round
	}
}

// RoundConfirmed is called whenever a new round is confirmed
func (ges *gasEpochState) RoundConfirmed(round uint64) {
	ges.mut.Lock()
	defer ges.mut.Unlock()

	ges.roundForLimits = round

	// if already computed, only store the new round
	if ges.overEstimationFactor != noOverestimationFactor {
		return
	}

	isRoundFlagEnabled := ges.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, ges.roundForLimits)
	if !isRoundFlagEnabled {
		return
	}

	// new limits and overestimation should be enabled once the Supernova round is active
	ges.overEstimationFactor = ges.economicsFee.BlockCapacityOverestimationFactor()
	// epoch was previously held at (currentEpoch - 1) until the Supernova round activated.
	// now that the round is active, we advance epochForLimits to the real epoch.
	ges.epochForLimits = ges.epochForLimits + 1
}

// GetEpochForLimitsAndOverEstimationFactor returns the epoch for limits and the overestimation factor
func (ges *gasEpochState) GetEpochForLimitsAndOverEstimationFactor() (uint32, uint64) {
	ges.mut.RLock()
	defer ges.mut.RUnlock()

	return ges.epochForLimits, ges.overEstimationFactor
}

// IsInterfaceNil returns true if there is no value under the interface
func (ges *gasEpochState) IsInterfaceNil() bool {
	return ges == nil
}
