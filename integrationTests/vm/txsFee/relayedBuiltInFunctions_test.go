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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("before relayed base cost fix", testRelayedBuildInFunctionChangeOwnerCallShouldWork(integrationTests.UnreachableEpoch, big.NewInt(25610), big.NewInt(4390)))
	t.Run("after relayed base cost fix", testRelayedBuildInFunctionChangeOwnerCallShouldWork(0, big.NewInt(24854), big.NewInt(5146)))
}

func testRelayedBuildInFunctionChangeOwnerCallShouldWork(
	relayedFixActivationEpoch uint32,
	expectedBalanceRelayer *big.Int,
	expectedAccumulatedFees *big.Int,
) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(
			config.EnableEpochs{
				PenalizedTooMuchGasEnableEpoch: integrationTests.UnreachableEpoch,
				FixRelayedBaseCostEnableEpoch:  relayedFixActivationEpoch,
			}, gasPriceModifier)
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9991691, 8309, 39)
		testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)

		relayerAddr := []byte("12345678901234567890123456789033")
		newOwner := []byte("12345678901234567890123456789112")
		gasLimit := uint64(1000)

		txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
		innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

		rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.Ok, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		utils.CheckOwnerAddr(t, testContext, scAddress, newOwner)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

		expectedBalance := big.NewInt(9991691)
		vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccumulatedFees, accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(91), developerFees)
	}
}

func TestRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("before relayed base cost fix", testRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(integrationTests.UnreachableEpoch, big.NewInt(25610), big.NewInt(4390)))
	t.Run("after relayed base cost fix", testRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(0, big.NewInt(25610), big.NewInt(4390)))
}

func testRelayedBuildInFunctionChangeOwnerCallWrongOwnerShouldConsumeGas(
	relayedFixActivationEpoch uint32,
	expectedBalanceRelayer *big.Int,
	expectedAccumulatedFees *big.Int,
) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedBaseCostEnableEpoch: relayedFixActivationEpoch,
		}, gasPriceModifier)
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9991691, 8309, 39)
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
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)
		require.Equal(t, process.ErrFailedTransaction, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		utils.CheckOwnerAddr(t, testContext, scAddress, owner)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

		expectedBalance := big.NewInt(9991691)
		vm.TestAccount(t, testContext.Accounts, owner, 1, expectedBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccumulatedFees, accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(0), developerFees)
	}
}

func TestRelayedBuildInFunctionChangeOwnerInvalidAddressShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)
	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	newOwner := []byte("invalidAddress")
	gasLimit := uint64(1000)

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.ExecutionFailed, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContext, scAddress, owner)

	expectedBalanceRelayer := big.NewInt(17330)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	expectedBalance := big.NewInt(9988100)
	vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(12670), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}

func TestRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("nonce fix is disabled, should increase the sender's nonce", func(t *testing.T) {
		testRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(t,
			config.EnableEpochs{
				RelayedNonceFixEnableEpoch:    1000,
				FixRelayedBaseCostEnableEpoch: 1000,
			})
	})
	t.Run("nonce fix is enabled, should still increase the sender's nonce", func(t *testing.T) {
		testRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(t,
			config.EnableEpochs{
				RelayedNonceFixEnableEpoch:    0,
				FixRelayedBaseCostEnableEpoch: 1000,
			})
	})
}

func testRelayedBuildInFunctionChangeOwnerCallInsufficientGasLimitShouldConsumeGas(
	t *testing.T,
	enableEpochs config.EnableEpochs,
) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(enableEpochs, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)
	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	newOwner := []byte("12345678901234567890123456789112")

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	gasLimit := uint64(len(txData) - 1)
	innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContext, scAddress, owner)

	expectedBalanceRelayer := big.NewInt(25810)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	expectedBalance := big.NewInt(9988100)
	vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(4190), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}

func TestRelayedBuildInFunctionChangeOwnerCallOutOfGasShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{}, 1)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, owner := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9988100, 11900, 399)
	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	newOwner := []byte("12345678901234567890123456789112")

	txData := []byte(core.BuiltInFunctionChangeOwnerAddress + "@" + hex.EncodeToString(newOwner))
	gasLimit := uint64(len(txData)) + minGasLimit
	innerTx := vm.CreateTransaction(1, big.NewInt(0), owner, scAddress, gasPrice, gasLimit, txData)

	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	rtxData := integrationTests.PrepareRelayedTxDataV1(innerTx)
	rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, innerTx.Value, relayerAddr, owner, gasPrice, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.ExecutionFailed, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckOwnerAddr(t, testContext, scAddress, owner)

	expectedBalanceRelayer := big.NewInt(25790)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalanceRelayer)

	expectedBalance := big.NewInt(9988100)
	vm.TestAccount(t, testContext.Accounts, owner, 2, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(4210), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}
