package process

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
)

// GasProcessingPolicy holds candidate-specific gas processing overrides.
type GasProcessingPolicy struct {
	maxGasLimitPerBlock         uint64
	hasMaxGasLimitPerBlockValue bool
}

// ResolveGasProcessingPolicy resolves gas processing overrides from the candidate header.
func ResolveGasProcessingPolicy(
	header data.HeaderHandler,
	enableEpochsHandler common.EnableEpochsHandler,
	enableRoundsHandler common.EnableRoundsHandler,
	feeHandler FeeHandler,
	shardID uint32,
) (GasProcessingPolicy, error) {
	if check.IfNil(header) {
		return GasProcessingPolicy{}, ErrNilBlockHeader
	}
	if check.IfNil(enableEpochsHandler) {
		return GasProcessingPolicy{}, ErrNilEnableEpochsHandler
	}
	if check.IfNil(enableRoundsHandler) {
		return GasProcessingPolicy{}, ErrNilEnableRoundsHandler
	}
	if check.IfNil(feeHandler) {
		return GasProcessingPolicy{}, ErrNilEconomicsFeeHandler
	}

	if header.IsHeaderV3() || !common.IsInSupernovaDrainWindowForEpochAndRound(
		enableEpochsHandler,
		enableRoundsHandler,
		header.GetEpoch(),
		header.GetRound(),
	) {
		return GasProcessingPolicy{}, nil
	}

	supernovaEpoch := enableEpochsHandler.GetActivationEpoch(common.SupernovaFlag)
	if supernovaEpoch == 0 {
		return GasProcessingPolicy{}, nil
	}

	return GasProcessingPolicy{
		maxGasLimitPerBlock:         feeHandler.MaxGasLimitPerBlockInEpoch(shardID, supernovaEpoch-1),
		hasMaxGasLimitPerBlockValue: true,
	}, nil
}

// HasMaxGasLimitPerBlock returns true when the policy overrides the total block gas limit.
func (policy GasProcessingPolicy) HasMaxGasLimitPerBlock() bool {
	return policy.hasMaxGasLimitPerBlockValue
}

// MaxGasLimitPerBlock returns the candidate-specific total block gas limit.
func (policy GasProcessingPolicy) MaxGasLimitPerBlock() uint64 {
	return policy.maxGasLimitPerBlock
}
