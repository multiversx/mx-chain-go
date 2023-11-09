package multiShard

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestRelayedBuiltInFunctionExecuteOnRelayerAndDstShardShouldWork(t *testing.T) {
	// TODO reinstate test after Wasm VM pointer fix
	if testing.Short() {
		t.Skip("cannot run with -race -short; requires Wasm VM fix")
	}

	testContextRelayer, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(
		2,
		config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: integrationTests.UnreachableEpoch,
		})
	require.Nil(t, err)
	defer testContextRelayer.Close()

	testContextInner, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(
		1,
		config.EnableEpochs{
			PenalizedTooMuchGasEnableEpoch: integrationTests.UnreachableEpoch,
		})
	require.Nil(t, err)
	defer testContextInner.Close()

	pathToContract := "../../wasm/testdata/counter/output/counter_old.wasm"
	scAddr, owner := utils.DoDeployOldCounter(t, testContextInner, pathToContract)
	gasAndFees := getZeroGasAndFees()
	testContextInner.TxFeeHandler.CreateBlockStarted(gasAndFees)
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

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
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

	expectedFees = big.NewInt(7000)
	accumulatedFees = testContextInner.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedFees, accumulatedFees)

	expectedRelayerBalance = big.NewInt(4610)
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, expectedRelayerBalance)
}
