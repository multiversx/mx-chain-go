//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

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
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	ret := vm.GetIntValueFromSC(nil, testContext.Accounts, scAddress, "get")
	require.Equal(t, big.NewInt(2), ret)

	expectedBalance := big.NewInt(23850)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(17950), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(807), developerFees)
}

func TestRelayedScCallContractNotFoundShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
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
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(18130)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(11870), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(0), developerFees)
}

func TestRelayedScCallInvalidMethodShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasLimit := uint64(1000)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("invalidMethod"))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(18050)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(23850), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(399), developerFees)
}

func TestRelayedScCallInsufficientGasLimitShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasLimit := uint64(5)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, _ := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(28100)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(13800), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(399), developerFees)
}

func TestRelayedScCallOutOfGasShouldConsumeGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeploy(t, testContext, "../wasm/testdata/counter/output/counter.wasm")
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	relayerAddr := []byte("12345678901234567890123456789033")
	sndAddr := []byte("12345678901234567890123456789112")
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	userTx := vm.CreateTransaction(0, big.NewInt(100), sndAddr, scAddress, gasPrice, gasLimit, []byte("increment"))

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, sndAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(27950)
	vm.TestAccount(t, testContext.Accounts, relayerAddr, 1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(13950), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(399), developerFees)
}

func TestRelayedDeployInvalidContractShouldIncrementNonceOnSender(t *testing.T) {
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
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(enableEpochs)
	require.Nil(t, err)

	relayerAddr := []byte("12345678901234567890123456789033")
	gasLimit := uint64(20)

	_, _ = vm.CreateAccount(testContext.Accounts, senderAddr, 0, big.NewInt(0))
	_, _ = vm.CreateAccount(testContext.Accounts, relayerAddr, 0, big.NewInt(30000))

	emptyAddress := make([]byte, len(senderAddr))
	userTx := vm.CreateTransaction(senderNonce, big.NewInt(100), senderAddr, emptyAddress, gasPrice, gasLimit, nil)

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + gasLimit + uint64(len(rtxData))
	rtx := vm.CreateTransaction(0, userTx.Value, relayerAddr, senderAddr, gasPrice, rTxGasLimit, rtxData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(rtx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	return testContext
}
