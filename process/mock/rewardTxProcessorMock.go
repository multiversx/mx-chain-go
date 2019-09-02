package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
)

type RewardTxProcessorMock struct {
	ProcessRewardTransactionCalled func(rTx *rewardTx.RewardTx) error
	ProcessCreatedRewardTransactionCalled func(rTx *rewardTx.RewardTx) error
}

func (scrp *RewardTxProcessorMock) ProcessRewardTransaction(rTx *rewardTx.RewardTx) error {
	if scrp.ProcessRewardTransactionCalled == nil {
		return nil
	}

	return scrp.ProcessRewardTransactionCalled(rTx)
}

func (scrp *RewardTxProcessorMock) ProcessCreatedRewardTransaction(rTx *rewardTx.RewardTx) error {
	if scrp.ProcessCreatedRewardTransactionCalled == nil {
		return nil
	}

	return scrp.ProcessCreatedRewardTransactionCalled(rTx)
}