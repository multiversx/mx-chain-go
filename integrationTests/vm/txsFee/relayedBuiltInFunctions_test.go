//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestRelayedBuildInFunctionChangeOwnerCallShouldWork(t *testing.T) {
	t.Run("before relayed move balance fix", testRelayedBuildInFunctionChangeOwnerCallShouldWork(integrationTests.UnreachableEpoch))
	t.Run("after relayed move balance fix", testRelayedBuildInFunctionChangeOwnerCallShouldWork(0))
}

func testRelayedBuildInFunctionChangeOwnerCallShouldWork(relayedFixActivationEpoch uint32) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(
			config.EnableEpochs{
				PenalizedTooMuchGasEnableEpoch:   integrationTests.UnreachableEpoch,
				FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
			})
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
		testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)

		relayerAddr := []byte("12345678901234567890123456789033")
		newOwner := []byte("12345678901234567890123456789112")
		gasLimit := uint64(1000)

		txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
		innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

		rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
		rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.Ok, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		utils.CheckOwnerAddr(t, testContext, scAddress, newOwner)

		expectedBalanceRelayer := big.NewInt(16610)
		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

		expectedBalance := big.NewInt(88100)
		vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, big.NewInt(13390), accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(915), developerFees)
	}
}

func TestRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(t *testing.T) {
	t.Run("before relayed move balance fix", testRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(integrationTests.UnreachableEpoch))
	t.Run("after relayed move balance fix", testRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(0))
}

func testRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(relayedFixActivationEpoch uint32) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedMoveBalanceEnableEpoch: relayedFixActivationEpoch,
		})
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
		testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789113")
		newOwner := []byte("12345678901234567890123456789112")
		gasLimit := uint64(1000)

		txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
		innerTx := vm.CreateTransaction(1, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData)

		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

		rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
		rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)
		require.Equal(t, process.ErrFailedTransaction, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		utils.CheckOwnerAddr(t, testContext, scAddress, owner)

		expectedBalanceRelayer := big.NewInt(16610)
		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

		expectedBalance := big.NewInt(88100)
		vm.TestAccount(t, testContext.Accounts, owner, 1, expectedBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, big.NewInt(13390), accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(0), developerFees)
	}
}

func TestRelayedBuildInFunctionChangeOwnerInvalidAddressShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	newOwner := []byte("invalidAddress")
	gasLimit := uint64(1000)

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.ExecutionFailed, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContext, scAddress, owner)

	expectedBalanceRelayer := big.NewInt(17330)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	expectedBalance := big.NewInt(88100)
	vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(12670), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}

func TestRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(t *testing.T) {
	t.Run("nonce fix is disabled, should increase the sender's nonce", func(t *testing.T) {
		testRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(t,
			config.EnableEpochs{
				RelayedNonceFixEnableEpoch:       1000,
				FixRelayedMoveBalanceEnableEpoch: 1000,
			})
	})
	t.Run("nonce fix is enabled, should still increase the sender's nonce", func(t *testing.T) {
		testRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(t,
			config.EnableEpochs{
				RelayedNonceFixEnableEpoch:       0,
				FixRelayedMoveBalanceEnableEpoch: 1000,
			})
	})
}

func testRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(
	t *testing.T,
	enableEpochs config.EnableEpochs,
) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(enableEpochs)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	newOwner := []byte("12345678901234567890123456789112")

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	gasLimit := uint64(len(txData) - 1)
	innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContext, scAddress, owner)

	expectedBalanceRelayer := big.NewInt(25810)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	expectedBalance := big.NewInt(88100)
	vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(4190), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}

func TestRelayedBuildInFunctionChangeOwnerCallOutOfGasShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	newOwner := []byte("12345678901234567890123456789112")

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	gasLimit := uint64(len(txData) + 1)
	innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.ExecutionFailed, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContext, scAddress, owner)

	expectedBalanceRelayer := big.NewInt(25790)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	expectedBalance := big.NewInt(88100)
	vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(4210), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}
