package api

import (
	rewardTxData "github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// APIRewardsHandlerStub -
type APIRewardsHandlerStub struct {
	PrepareRewardTxCalled func(tx *rewardTxData.RewardTx) *transaction.ApiTransactionResult
}

// PrepareRewardTx -
func (stub *APIRewardsHandlerStub) PrepareRewardTx(tx *rewardTxData.RewardTx) *transaction.ApiTransactionResult {
	if stub.PrepareRewardTxCalled != nil {
		return stub.PrepareRewardTxCalled(tx)
	}

	return &transaction.ApiTransactionResult{}
}

// IsInterfaceNil -
func (stub *APIRewardsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
