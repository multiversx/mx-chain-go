package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// EpochRewardsDisabled does nothing
type EpochRewardsDisabled struct{}

// NewDisabledEpochRewards will create a new instance of *EpochRewardsDisabled
func NewDisabledEpochRewards() *EpochRewardsDisabled {
	return &EpochRewardsDisabled{}
}

// GetRewardsTxs will return nothing
func (der *EpochRewardsDisabled) GetRewardsTxs(_ *block.Body) map[string]data.TransactionHandler {
	return make(map[string]data.TransactionHandler)
}

// IsInterfaceNil returns true if the underlying object is nil
func (der *EpochRewardsDisabled) IsInterfaceNil() bool {
	return der == nil
}
