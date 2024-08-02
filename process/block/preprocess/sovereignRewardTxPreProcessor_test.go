package preprocess

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestSovereignRewardsTxPreProcFactory_receivedRewardTransaction(t *testing.T) {
	args := createArgsRewardsPreProc()

	rwdTx := &rewardTx.RewardTx{
		RcvAddr: []byte("addr"),
	}
	wasRegisterCalled := false
	wasRewardTxProcessed := false
	args.RewardTxDataPool = &testscommon.ShardedDataStub{
		RegisterOnAddedCalled: func(f func(key []byte, value interface{})) {
			funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
			require.True(t, strings.Contains(funcName, "(*sovereignRewardsTxPreProcessor).receivedRewardTransaction"))
			wasRegisterCalled = true
		},
	}
	args.RewardProcessor = &testscommon.RewardTxProcessorMock{
		ProcessRewardTransactionCalled: func(rTx *rewardTx.RewardTx) error {
			require.Equal(t, rwdTx, rTx)
			wasRewardTxProcessed = true
			return nil
		},
	}

	sovProc, err := NewSovereignRewardsTxPreProcessor(args)
	require.Nil(t, err)
	require.False(t, sovProc.IsInterfaceNil())

	txHash := []byte("txHash")
	sovProc.receivedRewardTransaction(txHash, rwdTx)
	require.Equal(t, sovProc.rewardTxsForBlock.txHashAndInfo[string(txHash)].tx, rwdTx)
	require.True(t, wasRegisterCalled)
	require.True(t, wasRewardTxProcessed)
}
