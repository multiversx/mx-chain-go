package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	rewardTxData "github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewAPIRewardsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		apiRewardsHandler, err := NewAPIRewardsHandler(nil)
		require.Nil(t, apiRewardsHandler)
		require.Equal(t, errors.ErrNilAddressPublicKeyConverter, err)
	})

	t.Run("should work", func(t *testing.T) {
		apiRewardsHandler, err := NewAPIRewardsHandler(&testscommon.PubkeyConverterMock{})
		require.Nil(t, err)
		require.False(t, apiRewardsHandler.IsInterfaceNil())
	})
}

func TestApiRewards_PrepareRewardTx(t *testing.T) {
	t.Parallel()

	pkConv := &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			return string(pkBytes)
		},
	}
	apiRewardsHandler, _ := NewAPIRewardsHandler(pkConv)

	apiRewardTx := apiRewardsHandler.PrepareRewardTx(nil)
	require.Equal(t, &transaction.ApiTransactionResult{}, apiRewardTx)

	rewardTx := &rewardTxData.RewardTx{
		Round:   45,
		Value:   big.NewInt(441),
		RcvAddr: []byte("rcv"),
		Epoch:   4,
	}
	apiRewardTx = apiRewardsHandler.PrepareRewardTx(rewardTx)
	require.Equal(t, &transaction.ApiTransactionResult{
		Tx:          rewardTx,
		Type:        string(transaction.TxTypeReward),
		Round:       45,
		Epoch:       4,
		Value:       "441",
		Sender:      "metachain",
		Receiver:    "rcv",
		SourceShard: core.MetachainShardId,
	}, apiRewardTx)
}
