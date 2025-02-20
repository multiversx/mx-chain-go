package transactionAPI

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	rewardTxData "github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"

	"github.com/multiversx/mx-chain-go/errors"
)

type apiRewards struct {
	addressPubKeyConverter core.PubkeyConverter
}

// NewAPIRewardsHandler creates an api rewards handler for normal run type
func NewAPIRewardsHandler(addressPubKeyConverter core.PubkeyConverter) (*apiRewards, error) {
	if check.IfNil(addressPubKeyConverter) {
		return nil, errors.ErrNilAddressPublicKeyConverter
	}

	return &apiRewards{
		addressPubKeyConverter: addressPubKeyConverter,
	}, nil
}

// PrepareRewardTx prepares an api reward tx with sender and receiver shard as metachain
func (sar *apiRewards) PrepareRewardTx(tx *rewardTxData.RewardTx) *transaction.ApiTransactionResult {
	if check.IfNil(tx) {
		return &transaction.ApiTransactionResult{}
	}

	return &transaction.ApiTransactionResult{
		Tx:          tx,
		Type:        string(transaction.TxTypeReward),
		Round:       tx.GetRound(),
		Epoch:       tx.GetEpoch(),
		Value:       tx.GetValue().String(),
		Sender:      "metachain",
		Receiver:    sar.addressPubKeyConverter.SilentEncode(tx.GetRcvAddr(), log),
		SourceShard: core.MetachainShardId,
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (sar *apiRewards) IsInterfaceNil() bool {
	return sar == nil
}
