package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
)

// RewardTxProcessorMock -
type RewardTxProcessorMock struct {
	ProcessRewardTransactionCalled func(rTx *rewardTx.RewardTx) error
}

// ProcessRewardTransaction -
func (scrp *RewardTxProcessorMock) ProcessRewardTransaction(rTx *rewardTx.RewardTx) error {
	if scrp.ProcessRewardTransactionCalled == nil {
		return nil
	}

	return scrp.ProcessRewardTransactionCalled(rTx)
}

// IsInterfaceNil -
func (scrp *RewardTxProcessorMock) IsInterfaceNil() bool {
	return scrp == nil
}
