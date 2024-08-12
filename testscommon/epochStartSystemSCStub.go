package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
)

// EpochStartSystemSCStub -
type EpochStartSystemSCStub struct {
	ProcessSystemSmartContractCalled func(validatorsInfo state.ShardValidatorsInfoMapHandler, header data.HeaderHandler) error
	ProcessDelegationRewardsCalled   func(miniBlocks block.MiniBlockSlice, txCache epochStart.TransactionCacher) error
	ToggleUnStakeUnBondCalled        func(value bool) error
}

// ToggleUnStakeUnBond -
func (e *EpochStartSystemSCStub) ToggleUnStakeUnBond(value bool) error {
	if e.ToggleUnStakeUnBondCalled != nil {
		return e.ToggleUnStakeUnBondCalled(value)
	}
	return nil
}

// ProcessSystemSmartContract -
func (e *EpochStartSystemSCStub) ProcessSystemSmartContract(
	validatorsInfo state.ShardValidatorsInfoMapHandler,
	header data.HeaderHandler,
) error {
	if e.ProcessSystemSmartContractCalled != nil {
		return e.ProcessSystemSmartContractCalled(validatorsInfo, header)
	}
	return nil
}

// ProcessDelegationRewards -
func (e *EpochStartSystemSCStub) ProcessDelegationRewards(
	miniBlocks block.MiniBlockSlice,
	txCache epochStart.TransactionCacher,
) error {
	if e.ProcessDelegationRewardsCalled != nil {
		return e.ProcessDelegationRewardsCalled(miniBlocks, txCache)
	}
	return nil
}

// IsInterfaceNil -
func (e *EpochStartSystemSCStub) IsInterfaceNil() bool {
	return e == nil
}
