package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

type epochStartSystemSCProcessor struct {
}

// NewDisabledEpochStartSystemSC creates a new disabled EpochStartSystemSCProcessor instance
func NewDisabledEpochStartSystemSC() *epochStartSystemSCProcessor {
	return &epochStartSystemSCProcessor{}
}

// ToggleUnStakeUnBond returns nil
func (e *epochStartSystemSCProcessor) ToggleUnStakeUnBond(_ bool) error {
	return nil
}

// ProcessSystemSmartContract returns nil
func (e *epochStartSystemSCProcessor) ProcessSystemSmartContract(
	_ state.ShardValidatorsInfoMapHandler,
	_ data.HeaderHandler,
) error {
	return nil
}

// ProcessDelegationRewards returns nil
func (e *epochStartSystemSCProcessor) ProcessDelegationRewards(
	_ block.MiniBlockSlice,
	_ epochStart.TransactionCacher,
) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *epochStartSystemSCProcessor) IsInterfaceNil() bool {
	return e == nil
}
