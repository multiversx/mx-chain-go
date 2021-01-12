package multiShard

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/stretchr/testify/require"
)

func TestRelayedBuiltInFunctionExecuteOnRelayerAndDstShardShouldWork(t *testing.T) {
	testContextRelayer := vm.CreatePreparedTxProcessorWithVMsMultiShard(
		t,
		2,
		vm.ArgEnableEpoch{
			PenalizedTooMuchGasEnableEpoch: 100,
		})
	defer testContextRelayer.Close()

	testContextInner := vm.CreatePreparedTxProcessorWithVMsMultiShard(
		t,
		1,
		vm.ArgEnableEpoch{
			PenalizedTooMuchGasEnableEpoch: 100,
		})
	defer testContextInner.Close()

	pathToContract := "../../arwen/testdata/counter/output/counter.wasm"
	scAddr, owner := utils.DoDeploy(t, testContextInner, pathToContract)
	testContextInner.TxFeeHandler.CreateBlockStarted()
	utils.CleanAccumulatedIntermediateTransactions(t, testContextInner)

	require.Equal(t, uint32(1), testContextInner.ShardCoordinator.ComputeId(scAddr))
	require.Equal(t, uint32(1), testContextInner.ShardCoordinator.ComputeId(owner))

	relayerAddr := []byte("12345678901234567890123456789012")
	require.Equal(t, uint32(2), testContextInner.ShardCoordinator.ComputeId(relayerAddr))

	gasPrice := uint64(10)
	gasLimit := uint64(700)
	newOwner := []byte("12345678901234567890123456789112")
	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddr, gasPrice, gasLimit, txData)

	_, _ = vm.CreateAccount(testContextRelayer.Accounts, relayerAddr, 0, big.NewInt(15000))

	rtxData := utils.PrepareRelayerTxData(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

	// execute on relayer shard
	retCode, err := testContextRelayer.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	expectedRelayerBalance := big.NewInt(4610)
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, expectedRelayerBalance)

	expectedFees := big.NewInt(3390)
	accumulatedFees := testContextRelayer.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedFees, accumulatedFees)

	// execute on inner tx shard
	retCode, err = testContextInner.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContextInner, scAddr, newOwner)

	expectedFees = big.NewInt(850)
	accumulatedFees = testContextInner.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedFees, accumulatedFees)

	txs := testContextInner.GetIntermediateTransactions(t)
	scr := txs[1]
	utils.ProcessSCRResult(t, testContextRelayer, scr, vmcommon.Ok, nil)

	expectedRelayerBalance = big.NewInt(10760)
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, expectedRelayerBalance)
}
