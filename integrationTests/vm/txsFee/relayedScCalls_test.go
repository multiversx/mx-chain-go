package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayedScCallShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("before relayed base cost fix", testRelayedScCallShouldWork(integrationTests.UnreachableEpoch, big.NewInt(29982306), big.NewInt(25903), big.NewInt(1608)))
	t.Run("after relayed base cost fix", testRelayedScCallShouldWork(0, big.NewInt(29982216), big.NewInt(25993), big.NewInt(1608)))
}

func testRelayedScCallShouldWork(
	relayedFixActivationEpoch uint32,
	expectedRelayerBalance *big.Int,
	expectedAccFees *big.Int,
	expectedDevFees *big.Int,
) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
			FixRelayedBaseCostEnableEpoch:                   relayedFixActivationEpoch,
		}, gasPriceModifier)
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9991691, 8309, 39)
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789112")
		gasLimit := uint64(100000)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000000))

		userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.Ok, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
		require.Equal(t, big.NewInt(2), ret)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedRelayerBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccFees, accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, expectedDevFees, developerFees)
	}
}

func TestRelayedScCallContractNotFoundShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("before relayed fix", testRelayedScCallContractNotFoundShouldConsumeGas(integrationTests.UnreachableEpoch, big.NewInt(27130), big.NewInt(2870)))
	t.Run("after relayed fix", testRelayedScCallContractNotFoundShouldConsumeGas(0, big.NewInt(27040), big.NewInt(2960)))
}

func testRelayedScCallContractNotFoundShouldConsumeGas(
	relayedFixActivationEpoch uint32,
	expectedRelayerBalance *big.Int,
	expectedAccFees *big.Int,
) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedBaseCostEnableEpoch: relayedFixActivationEpoch,
		}, gasPriceModifier)
		require.Nil(t, err)
		defer testContext.Close()

		scAddress := "00000000000000000500dbb53e4b23392b0d6f36cce32deb2d623e9625ab3132"
		scAddrBytes, _ := hex.DecodeString(scAddress)

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789112")
		gasLimit := uint64(1000)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

		userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddrBytes, gasPrice, gasLimit, []byte("increment"))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedRelayerBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccFees, accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(0), developerFees)
	}
}

func TestRelayedScCallInvalidMethodShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("before relayed fix", testRelayedScCallInvalidMethodShouldConsumeGas(integrationTests.UnreachableEpoch, big.NewInt(26924), big.NewInt(11385)))
	t.Run("after relayed fix", testRelayedScCallInvalidMethodShouldConsumeGas(0, big.NewInt(26924), big.NewInt(11385)))
}

func testRelayedScCallInvalidMethodShouldConsumeGas(
	relayedFixActivationEpoch uint32,
	expectedRelayerBalance *big.Int,
	expectedAccFees *big.Int,
) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			RelayedNonceFixEnableEpoch: relayedFixActivationEpoch,
		}, gasPriceModifier)
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9991691, 8309, 39)
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789112")
		gasLimit := uint64(1000)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

		userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("invalidMethod"))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedRelayerBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccFees, accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(39), developerFees)
	}
}

func TestRelayedScCallInsufficientGasLimitShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("before relayed fix", testRelayedScCallInsufficientGasLimitShouldConsumeGas(integrationTests.UnreachableEpoch, big.NewInt(28140), big.NewInt(10169)))
	t.Run("after relayed fix", testRelayedScCallInsufficientGasLimitShouldConsumeGas(0, big.NewInt(28050), big.NewInt(10259)))
}

func testRelayedScCallInsufficientGasLimitShouldConsumeGas(
	relayedFixActivationEpoch uint32,
	expectedBalance *big.Int,
	expectedAccumulatedFees *big.Int,
) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			FixRelayedBaseCostEnableEpoch: relayedFixActivationEpoch,
		}, gasPriceModifier)
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9991691, 8309, 39)
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789112")
		data := "increment"
		gasLimit := minGasLimit + uint64(len(data))

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

		userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte(data))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccumulatedFees, accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(39), developerFees)
	}
}

func TestRelayedScCallOutOfGasShouldConsumeGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("before relayed fix", testRelayedScCallOutOfGasShouldConsumeGas(integrationTests.UnreachableEpoch, big.NewInt(28040), big.NewInt(10269)))
	t.Run("after relayed fix", testRelayedScCallOutOfGasShouldConsumeGas(0, big.NewInt(28040), big.NewInt(10269)))
}

func testRelayedScCallOutOfGasShouldConsumeGas(
	relayedFixActivationEpoch uint32,
	expectedRelayerBalance *big.Int,
	expectedAccFees *big.Int,
) func(t *testing.T) {
	return func(t *testing.T) {
		testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
			RelayedNonceFixEnableEpoch: relayedFixActivationEpoch,
		}, gasPriceModifier)
		require.Nil(t, err)
		defer testContext.Close()

		scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm", 9991691, 8309, 39)
		utils.CleanAccumulatedIntermediateTransactions(t, testContext)

		relayerAddr := []byte("12345678901234567890123456789033")
		sndAddr := []byte("12345678901234567890123456789112")
		gasLimit := uint64(20)

		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
		_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

		userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

		rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
		rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
		rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

		retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
		require.Equal(t, vmcommon.UserError, retCode)
		require.Nil(t, err)

		_, err = testContext.Accounts.Commit()
		require.Nil(t, err)

		vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedRelayerBalance)

		// check accumulated fees
		accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
		require.Equal(t, expectedAccFees, accumulatedFees)

		developerFees := testContext.TxFeeHandler.GetDeveloperFees()
		require.Equal(t, big.NewInt(39), developerFees)
	}
}

func TestRelayedDeployInvalidContractShouldIncrementNonceOnSender(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	senderAddr := []byte("12345678901234567890123456789011")

	t.Run("nonce fix is disabled, should increase the sender's nonce if inner tx has correct nonce", func(t *testing.T) {
		testContext := testRelayedDeployInvalidContractShouldIncrementNonceOnSender(t, config.EnableEpochs{
			RelayedNonceFixEnableEpoch: 100000,
		},
			senderAddr,
			0)
		defer testContext.Close()

		senderAccount := getAccount(t, testContext, senderAddr)
		assert.Equal(t, uint64(1), senderAccount.GetNonce())
	})
	t.Run("nonce fix is enabled, should still increase the sender's nonce if inner tx has correct nonce", func(t *testing.T) {
		testContext := testRelayedDeployInvalidContractShouldIncrementNonceOnSender(t, config.EnableEpochs{
			RelayedNonceFixEnableEpoch: 0,
		},
			senderAddr,
			0)
		defer testContext.Close()

		senderAccount := getAccount(t, testContext, senderAddr)
		assert.Equal(t, uint64(1), senderAccount.GetNonce())
	})
	t.Run("nonce fix is enabled, should not increase the sender's nonce if inner tx has higher nonce", func(t *testing.T) {
		testContext := testRelayedDeployInvalidContractShouldIncrementNonceOnSender(t, config.EnableEpochs{
			RelayedNonceFixEnableEpoch: 0,
		},
			senderAddr,
			1) // higher nonce, the current is 0
		defer testContext.Close()

		senderAccount := getAccount(t, testContext, senderAddr)
		assert.Equal(t, uint64(0), senderAccount.GetNonce())
	})
}

func testRelayedDeployInvalidContractShouldIncrementNonceOnSender(
	t *testing.T,
	enableEpochs config.EnableEpochs,
	senderAddr []byte,
	senderNonce uint64,
) *vm.VMTestContext {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(enableEpochs, 1)
	require.Nil(t, err)

	relayerAddr := []byte("12345678901234567890123456789033")
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, senderAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	emptyAddress := make([]byte, len(senderAddr))
	userTx := vm.CreateTransaction(senderNonce, big.NewInt(100), senderAddr, emptyAddress, gasPrice, gasLimit, nil)

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := minGasLimit + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, senderAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	return testContext
}
