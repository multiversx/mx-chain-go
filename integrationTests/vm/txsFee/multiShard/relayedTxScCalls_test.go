//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package multiShard

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

// Test scenario
// relayer address - shard 2
// inner transaction sender - shard 0
// inner transaction receiver (smart contract address) - shard 1

// 1. Do a SC deploy on shard 1
// 2. Execute relayed transaction on shard 2
// 3. Execute relayed transaction on shard 0
// 4. Execute SCR with the smart contract call on shard 1
// 5. Execute SCR with refund on relayer shard (shard 2)
func TestRelayedTxScCallMultiShardShouldWork(t *testing.T) {
	enableEpochs := config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	}

	testContextRelayer, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, enableEpochs)
	require.Nil(t, err)
	defer testContextRelayer.Close()

	testContextInnerSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, enableEpochs)
	require.Nil(t, err)
	defer testContextInnerSource.Close()

	testContextInnerDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, enableEpochs)
	require.Nil(t, err)
	defer testContextInnerDst.Close()

	pathToContract := "../../wasm/testdata/counter/output/counter_old.wasm"
	scAddr, owner := utils.DoDeployOldCounter(t, testContextInnerDst, pathToContract)
	testContextInnerDst.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContextInnerDst)

	require.Equal(t, uint32(1), testContextInnerDst.ShardCoordinator.ComputeId(scAddr))
	require.Equal(t, uint32(1), testContextInnerDst.ShardCoordinator.ComputeId(owner))

	sndAddr := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContextInnerDst.ShardCoordinator.ComputeId(sndAddr))

	relayerAddr := []byte("12345678901234567890123456789012")
	require.Equal(t, uint32(2), testContextInnerDst.ShardCoordinator.ComputeId(relayerAddr))

	gasPrice := uint64(10)
	gasLimit := uint64(500)

	innerTx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddr, gasPrice, gasLimit, []byte("increment"))
	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	_, _ = vm.CreateAccount(testContextRelayer.Accounts, relayerAddr, 0, big.NewInt(10000))

	// execute on relayer shard
	retCode, err := testContextRelayer.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextRelayer.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(3130)
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContextRelayer.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1870), accumulatedFees)

	developerFees := testContextRelayer.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)

	// execute on inner tx sender
	retCode, err = testContextInnerSource.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextRelayer.Accounts.Commit()
	require.Nil(t, err)

	// check balance of inner tx sender
	expectedBalance = big.NewInt(0)
	utils.TestAccount(t, testContextInnerSource.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees = testContextInnerSource.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(100), accumulatedFees)

	developerFees = testContextInnerSource.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)

	txs := testContextInnerSource.GetIntermediateTransactions(t)
	scr := txs[0]

	// execute on inner tx receiver ( shard with contract )
	utils.ProcessSCRResult(t, testContextInnerDst, scr, vmcommon.Ok, nil)

	ret := vm.GetIntValueFromSC(nil, testContextInnerDst.Accounts, scAddr, "get")
	require.Equal(t, big.NewInt(2), ret)

	// check accumulated fees dest
	accumulatedFees = testContextInnerDst.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(3770), accumulatedFees)

	developerFees = testContextInnerDst.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(377), developerFees)

	txs = testContextInnerDst.GetIntermediateTransactions(t)
	scr = txs[0]

	utils.ProcessSCRResult(t, testContextRelayer, scr, vmcommon.Ok, nil)
	expectedBalance = big.NewInt(4260)
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees = testContextRelayer.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1870), accumulatedFees)

	developerFees = testContextRelayer.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}

func TestRelayedTxScCallMultiShardFailOnInnerTxDst(t *testing.T) {
	testContextRelayer, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextRelayer.Close()

	testContextInnerSource, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(0, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextInnerSource.Close()

	testContextInnerDst, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextInnerDst.Close()

	pathToContract := "../../wasm/testdata/counter/output/counter_old.wasm"
	scAddr, owner := utils.DoDeployOldCounter(t, testContextInnerDst, pathToContract)
	testContextInnerDst.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContextInnerDst)

	require.Equal(t, uint32(1), testContextInnerDst.ShardCoordinator.ComputeId(scAddr))
	require.Equal(t, uint32(1), testContextInnerDst.ShardCoordinator.ComputeId(owner))

	sndAddr := []byte("12345678901234567890123456789010")
	require.Equal(t, uint32(0), testContextInnerDst.ShardCoordinator.ComputeId(sndAddr))

	relayerAddr := []byte("12345678901234567890123456789012")
	require.Equal(t, uint32(2), testContextInnerDst.ShardCoordinator.ComputeId(relayerAddr))

	gasPrice := uint64(10)
	gasLimit := uint64(500)

	innerTx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddr, gasPrice, gasLimit, []byte("incremeno"))
	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	_, _ = vm.CreateAccount(testContextRelayer.Accounts, relayerAddr, 0, big.NewInt(10000))

	// execute on relayer shard
	retCode, err := testContextRelayer.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextRelayer.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(3130)
	utils.TestAccount(t, testContextRelayer.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContextRelayer.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1870), accumulatedFees)

	developerFees := testContextRelayer.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)

	// execute on inner tx sender
	retCode, err = testContextInnerSource.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContextRelayer.Accounts.Commit()
	require.Nil(t, err)

	// check balance of inner tx sender
	expectedBalance = big.NewInt(0)
	utils.TestAccount(t, testContextInnerSource.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees = testContextInnerSource.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(100), accumulatedFees)

	developerFees = testContextInnerSource.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)

	txs := testContextInnerSource.GetIntermediateTransactions(t)
	scr := txs[0]

	// execute on inner tx receiver ( shard with contract )
	utils.ProcessSCRResult(t, testContextInnerDst, scr, vmcommon.UserError, nil)

	ret := vm.GetIntValueFromSC(nil, testContextInnerDst.Accounts, scAddr, "get")
	require.Equal(t, big.NewInt(1), ret)

	// check accumulated fees dest
	accumulatedFees = testContextInnerDst.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(4900), accumulatedFees)

	developerFees = testContextInnerDst.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)

	expectedBalance = big.NewInt(0)
	utils.TestAccount(t, testContextInnerSource.Accounts, sndAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees = testContextInnerSource.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(100), accumulatedFees)

	developerFees = testContextInnerSource.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}
