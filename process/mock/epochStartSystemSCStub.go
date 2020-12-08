package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartSystemSCStub -
type EpochStartSystemSCStub struct {
	ProcessSystemSmartContractCalled func(validatorInfos map[uint32][]*state.ValidatorInfo, nonce uint64, epoch uint32) error
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
func (e *EpochStartSystemSCStub) ProcessSystemSmartContract(validatorInfos map[uint32][]*state.ValidatorInfo, nonce uint64, epoch uint32) error {
	if e.ProcessSystemSmartContractCalled != nil {
		return e.ProcessSystemSmartContractCalled(validatorInfos, nonce, epoch)
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
